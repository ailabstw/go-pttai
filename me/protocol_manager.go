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
	"sync"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/event"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/raft"
	pb "github.com/ailabstw/go-pttai/raft/raftpb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProtocolManager struct {
	*pkgservice.BaseProtocolManager

	myPtt pkgservice.MyPtt

	// key-infos for providing join-friend
	lockJoinFriendKeyInfo sync.RWMutex
	joinFriendKeyInfos    []*pkgservice.KeyInfo

	// requests to join-friend
	lockJoinFriendRequest sync.RWMutex
	joinFriendRequests    map[common.Address]*pkgservice.JoinRequest
	joinFriendSub         *event.TypeMuxSubscription

	// merkle
	meOplogMerkle *pkgservice.Merkle

	// my-nodes
	lockJoinMeRequest sync.RWMutex
	joinMeRequests    map[common.Address]*pkgservice.JoinRequest
	joinMeSub         *event.TypeMuxSubscription

	lockMyNodes         sync.RWMutex
	MyNodes             map[uint64]*MyNode
	MyNodeByNodeSignIDs map[types.PttID]*MyNode
	totalWeight         uint32

	// dbLock
	dbMeLock     *types.LockMap
	dbMasterLock *types.LockMap

	// raft
	raftProposeC         chan string
	raftConfChangeC      chan pb.ConfChange
	raftForceConfChangeC chan pb.ConfChange
	raftCommitC          chan *string
	raftErrorC           chan error

	lockRaftLead sync.RWMutex
	raftLead     uint64

	lockRaftLastIndex sync.RWMutex
	raftLastIndex     uint64

	lockRaftConfState sync.RWMutex
	raftConfState     pb.ConfState

	lockRaftSnapshotIndex sync.RWMutex
	raftSnapshotIndex     uint64

	lockRaftAppliedIndex sync.RWMutex
	raftAppliedIndex     uint64

	raftNode raft.Node
	rs       *RaftStorage
}

func NewProtocolManager(myInfo *MyInfo, ptt pkgservice.MyPtt) (*ProtocolManager, error) {

	dbMeLock, err := types.NewLockMap(SleepTimeMeLock)
	if err != nil {
		return nil, err
	}

	dbMasterLock, err := types.NewLockMap(SleepTimeMasterLock)
	if err != nil {
		return nil, err
	}

	myID := myInfo.ID
	meOplogMerkle, err := pkgservice.NewMerkle(DBMeOplogPrefix, DBMeMerkleOplogPrefix, myID, dbMe)
	if err != nil {
		return nil, err
	}

	pm := &ProtocolManager{
		myPtt: ptt,

		// dblock
		dbMeLock:     dbMeLock,
		dbMasterLock: dbMasterLock,

		// join
		joinFriendKeyInfos: make([]*pkgservice.KeyInfo, 0),
		joinFriendRequests: make(map[common.Address]*pkgservice.JoinRequest),

		joinMeRequests: make(map[common.Address]*pkgservice.JoinRequest),

		// merkle
		meOplogMerkle: meOplogMerkle,

		//raft
		raftProposeC:         make(chan string),
		raftConfChangeC:      make(chan pb.ConfChange),
		raftForceConfChangeC: make(chan pb.ConfChange),
		raftCommitC:          make(chan *string),
		raftErrorC:           make(chan error),
	}

	b, err := pkgservice.NewBaseProtocolManager(ptt, RenewOpKeySeconds, ExpireOpKeySeconds, MaxSyncRandomSeconds, MinSyncRandomSeconds, pm.IsValidOplog, pm.IsMaster, myInfo, dbMe)
	if err != nil {
		return nil, err
	}
	pm.BaseProtocolManager = b

	err = pm.LoadMyNodes()
	if err != nil {
		log.Error("NewProtocolManager: unable to LoadMyNodes", "e", err)
		return nil, err
	}

	return pm, nil
}

func (pm *ProtocolManager) Start() error {
	ptt := pm.myPtt
	myInfo := pm.Entity().(*MyInfo)

	// start
	log.Debug("Start: start", "me", myInfo.GetID())
	err := pm.BaseProtocolManager.Start()
	if err != nil {
		return err
	}

	pm.LoadPeers()

	log.Debug("Start: to StartRaft", "status", myInfo.Status)
	myNodeType := ptt.MyNodeType()
	myNodeID := ptt.MyNodeID()
	myRaftID := ptt.MyRaftID()
	switch myInfo.Status {
	case types.StatusInit:
	case types.StatusInternalPending:
	case types.StatusInternalSync:
		go pm.StartRaft(nil, true)
	case types.StatusPending:
		weight := pm.nodeTypeToWeight(myNodeType)
		raftPeerList := []raft.Peer{{ID: myRaftID, Weight: weight, Context: myNodeID[:]}}
		go pm.StartRaft(raftPeerList, true)
	case types.StatusAlive:
		go pm.StartRaft(nil, false)
	}

	// join-me
	pm.joinMeSub = pm.EventMux().Subscribe(&JoinMeEvent{})
	go pm.JoinMeLoop()

	go pm.CreateJoinKeyInfoLoop()
	go pm.SyncJoinMeLoop()

	// oplog-merkle-tree

	go pkgservice.PMOplogMerkleTreeLoop(pm, pm.meOplogMerkle)

	// init me info
	go pm.InitMeInfoLoop()

	log.Debug("Start: done")

	return nil
}

func (pm *ProtocolManager) Stop() error {
	pm.BaseProtocolManager.PreStop()

	pm.StopRaft()

	//pm.joinFriendSub.Unsubscribe()
	pm.joinMeSub.Unsubscribe()

	err := pm.BaseProtocolManager.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) Sync(peer *pkgservice.PttPeer) error {
	if peer == nil {
		return nil
	}

	err := pm.SyncOplog(peer, pm.meOplogMerkle, SyncMeOplogMsg)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) GetJoinType(hash *common.Address) (pkgservice.JoinType, error) {
	if pm.IsJoinMeKeyHash(hash) {
		return pkgservice.JoinTypeMe, nil
	}

	return pkgservice.JoinTypeInvalid, pkgservice.ErrInvalidData
}

func (pm *ProtocolManager) IsMaster(id *types.PttID) bool {
	return true
}
