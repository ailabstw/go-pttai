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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pb "github.com/ailabstw/go-pttai/raft/raftpb"
)

func (pm *ProtocolManager) raftEntriesToApply(ents []pb.Entry) ([]pb.Entry, error) {
	var newEnts []pb.Entry

	if len(ents) == 0 {
		return newEnts, nil
	}

	firstIdx := ents[0].Index

	raftAppliedIndex := pm.GetRaftAppliedIndex(false)
	newFirstIdx := raftAppliedIndex - firstIdx + 1
	log.Debug("raftEntriesToApply", "firstIdx", firstIdx, "raftAppliedIdx", raftAppliedIndex, "newFirstIdx", newFirstIdx, "ents", len(ents))
	if newFirstIdx < uint64(len(ents)) {
		newEnts = ents[newFirstIdx:]
	}

	return newEnts, nil
}

func (pm *ProtocolManager) ProposeRaftAddNode(nodeID *discover.NodeID, weight uint32) error {
	raftID, err := nodeID.ToRaftID()
	if err != nil {
		return err
	}

	ctx := pm.proposeRaftCtx(nodeID)

	cc := pb.ConfChange{
		Type:    pb.ConfChangeAddNode,
		NodeID:  raftID,
		Weight:  weight,
		Context: ctx,
	}

	pm.raftConfChangeC <- cc

	return nil
}

func (pm *ProtocolManager) proposeRaftCtx(nodeID *discover.NodeID) []byte {
	ctx := make([]byte, discover.SizeNodeID+types.SizePttID)
	copy(ctx[:], nodeID[:])
	myNodeSignID := pm.Entity().(*MyInfo).NodeSignID
	copy(ctx[discover.SizeNodeID:], myNodeSignID[:])

	return ctx
}

func (pm *ProtocolManager) ProposeRaftRemoveNode(nodeID *discover.NodeID) error {
	raftID, err := nodeID.ToRaftID()
	if err != nil {
		return err
	}

	ctx := pm.proposeRaftCtx(nodeID)

	cc := pb.ConfChange{
		Type:    pb.ConfChangeRemoveNode,
		NodeID:  raftID,
		Context: ctx,
	}

	pm.raftConfChangeC <- cc

	return nil
}

func (pm *ProtocolManager) ForceProposeRaftRemoveNode(nodeID *discover.NodeID) error {

	raftLead := pm.GetRaftLead(false)
	if raftLead != 0 {
		return ErrWithLead
	}

	pm.lockMyNodes.RLock()
	defer pm.lockMyNodes.RUnlock()

	raftID, err := nodeID.ToRaftID()
	if err != nil {
		return err
	}

	_, ok := pm.MyNodes[raftID]
	if !ok {
		return ErrInvalidNode
	}

	ctx := pm.proposeRaftCtx(nodeID)

	cc := pb.ConfChange{
		Type:    pb.ConfChangeRemoveNode,
		NodeID:  raftID,
		Context: ctx,
	}

	pm.raftForceConfChangeC <- cc

	return nil
}

func (pm *ProtocolManager) ProposeRaftRequestLead() error {
	myRaftID := pm.myPtt.MyRaftID()

	raftLead := pm.GetRaftLead(false)
	if myRaftID == raftLead {
		return nil
	}

	if raftLead == 0 {
		return ErrInvalidNode
	}

	pm.raftNode.TransferLeadership(context.TODO(), raftLead, myRaftID)

	return nil
}

func (pm *ProtocolManager) PublishRaftEntries(ents []pb.Entry) error {
	var err error
	//log.Debug("PublishRaftEntries: start", "ents", ents)
	for i := range ents {
		switch ents[i].Type {
		case pb.EntryNormal:
			// XXX should be no meaningful EntryNormal
			if len(ents[i].Data) != 0 {
				log.Warn("PublishRaftEntries: EntryNormal", "ent", ents[i])
			}
		case pb.EntryConfChange:
			var cc pb.ConfChange
			cc.Unmarshal(ents[i].Data)

			log.Info("PublishRaftEntries: ConfChange", "cc", cc, "ent", ents[i])

			switch cc.Type {
			case pb.ConfChangeAddNode:
				err = pm.publishEntriesAddNode(&ents[i], &cc)
				if err != nil {
					log.Warn("publishEntriesAddNode: failed", "e", err)
				}

			case pb.ConfChangeRemoveNode:
				err = pm.publishEntriesRemoveNode(&ents[i], &cc)
				if err != nil {
					log.Warn("publishEntriesRemoveNode: failed", "e", err)
				}
			}

			cs := *pm.raftNode.ApplyConfChange(cc)
			pm.SetRaftConfState(cs, false)
		}

		// after commit, update appliedIndex
		pm.SetRaftAppliedIndex(ents[i].Index, false)

		// special nil commit to signal replay has finished
		// XXX should be no meaningful EntryNormal
	}
	return nil
}

func (pm *ProtocolManager) publishEntriesAddNode(ent *pb.Entry, cc *pb.ConfChange) error {
	ptt := pm.myPtt

	myInfo := pm.Entity().(*MyInfo)

	nodeID := &discover.NodeID{}
	copy(nodeID[:], cc.Context[:discover.SizeNodeID])

	fromID := &types.PttID{}
	copy(fromID[:], cc.Context[discover.SizeNodeID:])

	raftID, err := nodeID.ToRaftID()
	log.Debug("publishEntriesAddNode: after ToRaftID", "nodeID", nodeID, "e", err)
	if err != nil {
		return err
	}
	if raftID != cc.NodeID {
		return ErrInvalidEntry
	}

	weight := cc.Weight

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	// master-oplog and my-node
	oplog, err := pm.publishEntriesAddNodeCreateMasterOplogAndSetMyNode(ts, raftID, nodeID, weight, ent, fromID)
	log.Debug("publishEntriesAddNode: after SetMyNode", "e", err)
	if err != nil {
		return err
	}

	// create-me-or-add-dial
	myNodeID := ptt.MyNodeID()
	if reflect.DeepEqual(nodeID, myNodeID) {
		// create-me
		switch myInfo.Status {
		case types.StatusPending:
			err = pm.CreateFullMe(oplog)
			log.Debug("publishEntriesAddNode: after CreateFullMe", "e", err)
			if err != nil {
				return err
			}
		case types.StatusSync:
			err = pm.InternalSyncToAlive(oplog, weight)
			log.Debug("publishEntriesAddNode: after InternalSyncToAlive", "e", err)
			if err != nil {
				return err
			}
		}

	} else {
		opKey, err := pm.GetOldestOpKey(false)
		if err != nil {
			return err
		}
		ptt.AddDial(nodeID, opKey.Hash)
	}

	// oplog-save
	err = oplog.Save(false)
	log.Debug("publishEntriesAddNode: after oplog.Save", "e", err)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) publishEntriesAddNodeCreateMasterOplogAndSetMyNode(ts types.Timestamp, raftID uint64, nodeID *discover.NodeID, weight uint32, ent *pb.Entry, fromID *types.PttID) (*MasterOplog, error) {

	myInfo := pm.Entity().(*MyInfo)

	// lock
	pm.lockMyNodes.Lock()
	defer pm.lockMyNodes.Unlock()

	masters := make(map[discover.NodeID]uint32)
	for _, eachNode := range pm.MyNodes {
		if eachNode.Weight == 0 {
			continue
		}
		masters[*eachNode.NodeID] = eachNode.Weight
	}
	masters[*nodeID] = weight

	var totalWeight uint32
	for _, eachWeight := range masters {
		totalWeight += eachWeight
	}

	opData := &MasterOpAddMaster{
		ID:      nodeID,
		From:    fromID,
		Masters: masters,
		Weight:  totalWeight,
	}

	oplog, err := pm.CreateMasterOplog(ent.Index, ts, MasterOpTypeAddMaster, opData)
	if err != nil {
		return nil, err
	}

	myNode, ok := pm.MyNodes[raftID]
	if !ok {
		myNode, err = NewMyNode(ts, myInfo.ID, nodeID, 0)
		log.Debug("publishEntriesAddNode: after NewMyNode", "e", err)
		if err != nil {
			return nil, err
		}
		myNode.Status = types.StatusInternalPending
		pm.MyNodes[raftID] = myNode
	}

	if myNode.Status == types.StatusInit {
		log.Debug("publishEntriesAddNode: Status change from Init to InternalPending", "raftID", raftID, "NodeID", myNode.NodeID)
		myNode.Status = types.StatusInternalPending
	}

	origWeight := myNode.Weight
	myNode.Weight = weight
	myNode.UpdateTS = ts
	myNode.LogID = oplog.ID
	pm.totalWeight += weight - origWeight

	nodeSignID, err := setNodeSignID(nodeID, myInfo.ID)
	if err != nil {
		return nil, err
	}

	pm.MyNodeByNodeSignIDs[*nodeSignID] = myNode

	_, err = myNode.Save()
	log.Debug("publishEntriesAddNode: after myNode.Save", "e", err, "myNode", myNode)
	if err != nil {
		return nil, err
	}

	return oplog, nil
}

func (pm *ProtocolManager) publishEntriesRemoveNode(ent *pb.Entry, cc *pb.ConfChange) error {

	nodeID := &discover.NodeID{}
	copy(nodeID[:], cc.Context[:discover.SizeNodeID])

	fromID := &types.PttID{}
	copy(fromID[:], cc.Context[discover.SizeNodeID:])

	raftID, err := nodeID.ToRaftID()
	if err != nil {
		return err
	}
	if raftID != cc.NodeID {
		return ErrInvalidEntry
	}

	masters, totalWeight, myNode, err := pm.publishEntriesRemoveNodeDealWithMyNodes(raftID, nodeID)
	if err != nil {
		return err
	}
	pm.totalWeight = totalWeight

	opData := &MasterOpRevokeMaster{
		ID:      nodeID,
		From:    fromID,
		Masters: masters,
	}

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	oplog, err := pm.CreateMasterOplog(ent.Index, ts, MasterOpTypeRevokeMaster, opData)
	if err != nil {
		return err
	}

	myNode.Status = types.StatusDeleted
	myNode.LogID = oplog.ID

	_, err = myNode.Save()
	if err != nil {
		return err
	}

	err = oplog.Save(false)
	if err != nil {
		return err
	}

	myRaftID := pm.myPtt.MyRaftID()
	if raftID == myRaftID {
		return pm.HandleRevokeMyNode(oplog, false, true)
	}

	return pm.HandleRevokeOtherNode(oplog, myNode, fromID)
}

func (pm *ProtocolManager) publishEntriesRemoveNodeDealWithMyNodes(raftID uint64, nodeID *discover.NodeID) (map[discover.NodeID]uint32, uint32, *MyNode, error) {

	pm.lockMyNodes.Lock()
	defer pm.lockMyNodes.Unlock()

	myNode, ok := pm.MyNodes[raftID]
	if !ok {
		return nil, 0, nil, ErrInvalidNode
	}

	myInfo := pm.Entity().(*MyInfo)
	nodeSignID, err := setNodeSignID(nodeID, myInfo.ID)
	if err != nil {
		return nil, 0, nil, ErrInvalidNode
	}

	delete(pm.MyNodes, raftID)
	delete(pm.MyNodeByNodeSignIDs, *nodeSignID)

	// reset masters and totalWeight
	masters := make(map[discover.NodeID]uint32)
	var totalWeight uint32
	for _, eachNode := range pm.MyNodes {
		totalWeight += eachNode.Weight
		masters[*eachNode.NodeID] = eachNode.Weight
	}

	return masters, totalWeight, myNode, nil
}
