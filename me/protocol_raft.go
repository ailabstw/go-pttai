// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package me

import (
	"context"
	"encoding/binary"
	"reflect"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/raft"
	pb "github.com/ailabstw/go-pttai/raft/raftpb"
)

func (pm *ProtocolManager) StartRaft(peers []raft.Peer, isNew bool) error {
	myID := pm.Entity().GetID()

	myRaftID := pm.myPtt.MyRaftID()
	c := &raft.Config{
		ID:                        myRaftID,
		ElectionTick:              RaftElectionTick,
		HeartbeatTick:             RaftHeartbeatTick,
		MaxSizePerMsg:             RaftMaxSizePerMsg,
		MaxInflightMsgs:           RaftMaxInflightMsgs,
		CheckQuorum:               true,
		PreVote:                   true,
		DisableProposalForwarding: false,
	}
	if !isNew {
		log.Debug("StartRaft: to RestartNode")

		rs, err := NewRaftStorage(false, myID)
		if err != nil {
			return err
		}
		c.Storage = rs
		pm.rs = rs

		pm.raftNode = raft.RestartNode(c)

		log.Debug("StartRaft: after RestartNode")
	} else {
		rs, err := NewRaftStorage(true, myID)
		if err != nil {
			return err
		}
		c.Storage = rs
		pm.rs = rs

		myInfo := pm.Entity().(*MyInfo)
		log.Debug("StartRaft: to StartNode", "status", myInfo.Status, "c", c, "peers", peers, "isNil", peers == nil)

		pm.raftNode = raft.StartNode(c, peers)
	}

	go pm.ServeRaftChannels()

	if len(pm.MyNodes) == 1 {
		log.Debug("StartRaft: to Step")
		pm.raftNode.Step(context.TODO(), pb.Message{From: myRaftID, To: myRaftID, Type: pb.MsgHup})
		log.Debug("StartRaft: done")
	}

	return nil
}

func (pm *ProtocolManager) StopRaft() error {
	// XXX not init yet.
	// for types.StatusInit
	if pm.raftNode == nil {
		return nil
	}

	log.Debug("StopRaft: start", "myID", pm.Entity().(*MyInfo).ID)

	close(pm.raftCommitC)
	close(pm.raftErrorC)
	pm.raftNode.Stop()

	return nil
}

func (pm *ProtocolManager) ServeRaftChannels() error {

	snap, err := pm.rs.Snapshot()
	log.Debug("ServeRaftChannels: after get snapshot", "e", err)
	if err != nil {
		return err
	}

	pm.SetRaftConfState(snap.Metadata.ConfState, false)
	pm.SetRaftSnapshotIndex(snap.Metadata.Index, false)
	pm.SetRaftAppliedIndex(snap.Metadata.Index, false)

	tick := time.NewTicker(RaftTickTime)
	defer tick.Stop()

	go pm.ServeRaftProposal()

	ptt := pm.Ptt()

loop:
	for {
		select {
		case <-tick.C:
			pm.raftNode.Tick()

		case rd, ok := <-pm.raftNode.Ready():
			log.Debug("ServeRaftChannels: received Ready", "rd", rd)
			if !ok {
				break loop
			}

			if rd.SoftState != nil {
				log.Debug("ServeRaftChannels: to set Leader", "lead", rd.SoftState.Lead)
				pm.SetRaftLead(rd.SoftState.Lead, false)
			}

			if !raft.IsEmptyHardState(rd.HardState) {
				pm.rs.SetHardState(rd.HardState)
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				log.Debug("ServeRaftChannels: to do Snapshot")
				pm.rs.ApplySnapshot(rd.Snapshot)
				pm.PublishRaftSnapshot(rd.Snapshot)
			}

			pm.rs.Append(rd.Entries)
			pm.SendRaftMsgs(rd.Messages)

			raftEntriesToApply, err := pm.raftEntriesToApply(rd.CommittedEntries)
			log.Debug("ServeRaftChannels: after raftEntriesToApply", "raftEntriesToAppply", raftEntriesToApply, "e", err)
			if err != nil {
				ptt.ErrChan().PassChan(err)
				break loop
			}

			if err := pm.PublishRaftEntries(raftEntriesToApply); err != nil {
				ptt.ErrChan().PassChan(err)
				break loop
			}
			pm.MaybeTriggerRaftSnapshot()
			pm.raftNode.Advance()
		case <-pm.QuitSync():
			break loop
		}
	}

	return nil
}

func (pm *ProtocolManager) ServeRaftProposal() {
	myID := pm.Entity().GetID()
	MyID := pm.Entity().Service().SPM().(*ServiceProtocolManager).MyInfo.ID
	if !reflect.DeepEqual(myID, MyID) {
		return
	}

	confChangeCount := uint64(0)

loop:
	for {
		select {
		case prop, ok := <-pm.raftProposeC:
			if !ok {
				break loop
			}
			pm.raftNode.Propose(context.TODO(), []byte(prop))
		case cc, ok := <-pm.raftConfChangeC:
			if !ok {
				break loop
			}

			confChangeCount++
			cc.ID = confChangeCount
			pm.raftNode.ProposeConfChange(context.TODO(), cc)
		case cc, ok := <-pm.raftForceConfChangeC:
			if !ok {
				break loop
			}

			confChangeCount++
			cc.ID = confChangeCount
			pm.raftNode.ForceProposeConfChange(context.TODO(), cc)
		case <-pm.QuitSync():
			break loop
		}
	}
}

// raft-applied-index
func (pm *ProtocolManager) SetRaftAppliedIndex(idx uint64, isLocked bool) error {
	if !isLocked {
		pm.lockRaftAppliedIndex.Lock()
		defer pm.lockRaftAppliedIndex.Unlock()
	}

	key, err := pm.MarshalRaftAppliedIndexKey()
	if err != nil {
		return err
	}

	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, idx)

	err = dbMe.Put(key, val)
	if err != nil {
		return err
	}

	pm.raftAppliedIndex = idx

	return nil
}

func (pm *ProtocolManager) MarshalRaftAppliedIndexKey() ([]byte, error) {
	myID := pm.Entity().(*MyInfo).ID
	return common.Concat([][]byte{DBKeyRaftAppliedIndex, myID[:]})
}

func (pm *ProtocolManager) GetRaftAppliedIndex(isLocked bool) uint64 {
	if !isLocked {
		pm.lockRaftAppliedIndex.RLock()
		defer pm.lockRaftAppliedIndex.RUnlock()
	}

	return pm.raftAppliedIndex
}

func (pm *ProtocolManager) LoadRaftAppliedIndex(isLocked bool) error {
	if !isLocked {
		pm.lockRaftAppliedIndex.RLock()
		defer pm.lockRaftAppliedIndex.RUnlock()
	}

	key, err := pm.MarshalRaftAppliedIndexKey()
	if err != nil {
		return err
	}

	val, err := dbMe.Get(key)
	if err != nil {
		return err
	}

	pm.raftAppliedIndex = binary.BigEndian.Uint64(val)

	return nil

}

// raft-snapshot-index

func (pm *ProtocolManager) SetRaftSnapshotIndex(idx uint64, isLocked bool) error {
	if !isLocked {
		pm.lockRaftSnapshotIndex.Lock()
		defer pm.lockRaftSnapshotIndex.Unlock()
	}

	key, err := pm.MarshalRaftSnapshotIndexKey()
	if err != nil {
		return err
	}

	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, idx)

	err = dbMe.Put(key, val)
	if err != nil {
		return err
	}

	pm.raftSnapshotIndex = idx

	return nil
}

func (pm *ProtocolManager) MarshalRaftSnapshotIndexKey() ([]byte, error) {
	myID := pm.Entity().(*MyInfo).ID
	return common.Concat([][]byte{DBKeyRaftSnapshotIndex, myID[:]})
}

func (pm *ProtocolManager) GetRaftSnapshotIndex(isLocked bool) uint64 {
	if !isLocked {
		pm.lockRaftSnapshotIndex.RLock()
		defer pm.lockRaftSnapshotIndex.RUnlock()
	}

	return pm.raftSnapshotIndex
}

func (pm *ProtocolManager) LoadRaftSnapshotIndex(isLocked bool) error {
	if !isLocked {
		pm.lockRaftSnapshotIndex.RLock()
		defer pm.lockRaftSnapshotIndex.RUnlock()
	}

	key, err := pm.MarshalRaftSnapshotIndexKey()
	if err != nil {
		return err
	}

	val, err := dbMe.Get(key)
	if err != nil {
		return err
	}

	pm.raftSnapshotIndex = binary.BigEndian.Uint64(val)

	return nil

}

// raft-last-index

func (pm *ProtocolManager) SetRaftLastIndex(idx uint64, isLocked bool) error {
	if !isLocked {
		pm.lockRaftLastIndex.Lock()
		defer pm.lockRaftLastIndex.Unlock()
	}

	key, err := pm.MarshalRaftLastIndexKey()
	if err != nil {
		return err
	}

	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, idx)

	err = dbMe.Put(key, val)
	if err != nil {
		return err
	}

	pm.raftLastIndex = idx

	return nil
}

func (pm *ProtocolManager) MarshalRaftLastIndexKey() ([]byte, error) {
	myID := pm.Entity().(*MyInfo).ID
	return common.Concat([][]byte{DBKeyRaftLastIndex, myID[:]})
}

func (pm *ProtocolManager) GetRaftLastIndex(isLocked bool) uint64 {
	if !isLocked {
		pm.lockRaftLastIndex.RLock()
		defer pm.lockRaftLastIndex.RUnlock()
	}

	return pm.raftLastIndex
}

func (pm *ProtocolManager) LoadRaftLastIndex(isLocked bool) error {
	if !isLocked {
		pm.lockRaftLastIndex.RLock()
		defer pm.lockRaftLastIndex.RUnlock()
	}

	key, err := pm.MarshalRaftLastIndexKey()
	if err != nil {
		return err
	}

	val, err := dbMe.Get(key)
	if err != nil {
		return err
	}

	pm.raftLastIndex = binary.BigEndian.Uint64(val)

	return nil

}

// raft-lead

func (pm *ProtocolManager) SetRaftLead(idx uint64, isLocked bool) error {
	if !isLocked {
		pm.lockRaftLead.Lock()
		defer pm.lockRaftLead.Unlock()
	}

	pm.raftLead = idx

	return nil
}

func (pm *ProtocolManager) GetRaftLead(isLocked bool) uint64 {
	if !isLocked {
		pm.lockRaftLead.RLock()
		defer pm.lockRaftLead.RUnlock()
	}

	return pm.raftLead
}

// raft-conf-state

func (pm *ProtocolManager) SetRaftConfState(cs pb.ConfState, isLocked bool) error {
	if !isLocked {
		pm.lockRaftConfState.Lock()
		defer pm.lockRaftConfState.Unlock()
	}

	key, err := pm.MarshalRaftConfStateKey()
	if err != nil {
		return err
	}

	val, err := cs.Marshal()
	if err != nil {
		return err
	}

	err = dbMe.Put(key, val)
	if err != nil {
		return err
	}

	pm.raftConfState = cs

	return nil
}

func (pm *ProtocolManager) MarshalRaftConfStateKey() ([]byte, error) {
	myID := pm.Entity().(*MyInfo).ID
	return common.Concat([][]byte{DBKeyRaftConfState, myID[:]})
}

func (pm *ProtocolManager) GetRaftConfState(isLocked bool) pb.ConfState {
	if !isLocked {
		pm.lockRaftConfState.RLock()
		defer pm.lockRaftConfState.RUnlock()
	}

	return pm.raftConfState
}

func (pm *ProtocolManager) LoadRaftConfState(isLocked bool) error {
	if !isLocked {
		pm.lockRaftConfState.RLock()
		defer pm.lockRaftConfState.RUnlock()
	}

	key, err := pm.MarshalRaftConfStateKey()
	if err != nil {
		return err
	}

	val, err := dbMe.Get(key)
	if err != nil {
		return err
	}

	err = pm.raftConfState.Unmarshal(val)
	if err != nil {
		return err
	}

	return nil
}
