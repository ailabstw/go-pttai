// Copyright 2019 The go-pttai Authors
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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/ailabstw/go-pttai/raft"
	pb "github.com/ailabstw/go-pttai/raft/raftpb"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
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

	// requests to join-friend
	lockJoinBoardRequest sync.RWMutex
	joinBoardRequests    map[common.Address]*pkgservice.JoinRequest
	joinBoardSub         *event.TypeMuxSubscription

	// my-nodes
	lockJoinMeRequest sync.RWMutex
	joinMeRequests    map[common.Address]*pkgservice.JoinRequest
	joinMeSub         *event.TypeMuxSubscription

	lockMyNodes         sync.RWMutex
	MyNodes             map[uint64]*MyNode
	MyNodeByNodeSignIDs map[types.PttID]*MyNode
	totalWeight         uint32

	// master-oplog
	dbMasterLock *types.LockMap

	// me-oplog
	dbMeLock      *types.LockMap
	meOplogMerkle *pkgservice.Merkle

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

	raftNode        raft.Node
	rs              *RaftStorage
	isStartRaftNode bool

	lockRaft sync.Mutex
}

func NewProtocolManager(myInfo *MyInfo, ptt pkgservice.MyPtt, svc pkgservice.Service) (*ProtocolManager, error) {

	dbMeLock, err := types.NewLockMap(SleepTimeMeLock)
	if err != nil {
		return nil, err
	}

	dbMasterLock, err := types.NewLockMap(SleepTimeMasterLock)
	if err != nil {
		return nil, err
	}

	myID := myInfo.ID
	entityIDBytes, _ := myID.MarshalText()
	entityIDStr := string(entityIDBytes)

	meOplogMerkle, err := pkgservice.NewMerkle(DBMeOplogPrefix, DBMeMerkleOplogPrefix, myID, dbMe, "("+entityIDStr+"/"+svc.Name()+":me)")
	if err != nil {
		return nil, err
	}

	log.Debug("NewProtocolManager: start", "myInfo", myInfo.IDString())

	pm := &ProtocolManager{
		myPtt: ptt,

		// dblock
		dbMeLock:     dbMeLock,
		dbMasterLock: dbMasterLock,

		// join
		joinFriendKeyInfos: make([]*pkgservice.KeyInfo, 0),
		joinFriendRequests: make(map[common.Address]*pkgservice.JoinRequest),

		joinMeRequests: make(map[common.Address]*pkgservice.JoinRequest),

		joinBoardRequests: make(map[common.Address]*pkgservice.JoinRequest),

		// merkle
		meOplogMerkle: meOplogMerkle,

		//raft
		raftProposeC:         make(chan string),
		raftConfChangeC:      make(chan pb.ConfChange),
		raftForceConfChangeC: make(chan pb.ConfChange),
		raftCommitC:          make(chan *string),
		raftErrorC:           make(chan error),
	}

	b, err := pkgservice.NewBaseProtocolManager(
		ptt,

		RenewOpKeySeconds,
		ExpireOpKeySeconds,
		MaxSyncRandomSeconds,
		MinSyncRandomSeconds,

		MaxMasters,

		pm.meOplogMerkle, // log0Merkle

		// sign
		pm.InternalSignMyOplog,
		pm.ForceSignMyOplog,
		pm.IsValidMyOplog,
		pm.ValidateIntegrateSignMyOplog,

		pm.SetMeDB,        // setLog0DB
		pm.HandleMeOplogs, // handleLog0s

		pm.IsMaster, // isMaster
		pm.IsMember, // isMember

		// peer-type
		pm.GetPeerType,
		pm.IsMyDevice,
		pm.IsImportantPeer,
		pm.IsMemberPeer,
		pm.IsPendingPeer,

		nil, // postsyncMemberOplog

		nil, // theDelete
		nil, // postdelete

		myInfo, // entity
		svc,

		dbMe, // db
	)
	if err != nil {
		return nil, err
	}
	pm.BaseProtocolManager = b

	// master-log
	masterLogs, err := pm.GetMasterOplogList(nil, 1, pttdb.ListOrderNext, types.StatusAlive)
	if len(masterLogs) == 1 {
		pm.SetMasterLog0Hash(masterLogs[0].Hash)
	}

	// load-my-nodes
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
	log.Debug("Start: start", "me", myInfo.IDString())
	err := pm.BaseProtocolManager.Start()
	if err != nil {
		log.Error("Start: unable to start BaseProtocolManager", "e", err)
		return err
	}

	// load my profile
	allEntities := ptt.GetEntities()
	if myInfo.ProfileID != nil {
		myProfile, ok := allEntities[*myInfo.ProfileID]
		if !ok {
			log.Error("Start: my Profile not exists")
			return ErrInvalidMe
		}
		myInfo.Profile, ok = myProfile.(*account.Profile)
		if !ok {
			log.Error("Start: unable to load my profile")
			return ErrInvalidMe
		}
	}

	pm.LoadPeers()

	log.Debug("Start: to StartRaft", "status", myInfo.Status)
	myNodeType := ptt.MyNodeType()
	myNodeID := ptt.MyNodeID()
	myRaftID := ptt.MyRaftID()
	switch myInfo.Status {
	case types.StatusInit:
	case types.StatusInternalPending:
	case types.StatusSync:
		go pm.StartRaft(nil, true)
	case types.StatusPending:
		weight := pm.nodeTypeToWeight(myNodeType)
		raftPeerList := []raft.Peer{{ID: myRaftID, Weight: weight, Context: myNodeID[:]}}
		go pm.StartRaft(raftPeerList, true)
	case types.StatusAlive:
		go pm.StartRaft(nil, false)
	}

	syncWG := pm.SyncWG()

	// join-me
	pm.joinMeSub = pm.EventMux().Subscribe(&JoinMeEvent{})
	go pm.JoinMeLoop()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.CreateJoinKeyLoop()
	}()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.SyncJoinMeLoop()
	}()

	// join-friend
	pm.joinFriendSub = pm.EventMux().Subscribe(&JoinFriendEvent{})
	go pm.JoinFriendLoop()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.CreateJoinFriendKeyLoop()
	}()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.SyncJoinFriendLoop()
	}()

	// join-board
	pm.joinBoardSub = pm.EventMux().Subscribe(&JoinBoardEvent{})
	go pm.JoinBoardLoop()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.SyncJoinBoardLoop()
	}()

	// oplog-merkle-tree
	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pkgservice.PMOplogMerkleTreeLoop(pm, pm.meOplogMerkle)
	}()

	// init me info
	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.InitMeInfoLoop()
	}()

	log.Debug("Start: done")

	return nil
}

func (pm *ProtocolManager) Stop() error {
	if !pm.IsStart() {
		return nil
	}

	pm.joinFriendSub.Unsubscribe()
	pm.joinMeSub.Unsubscribe()
	pm.joinBoardSub.Unsubscribe()

	pm.StopRaft()

	return nil
}

func (pm *ProtocolManager) Sync(peer *pkgservice.PttPeer) error {
	if peer == nil {
		pm.SyncPendingMeOplog(peer)
		return nil
	}

	log.Debug("Sync: Start", "entity", pm.Entity().IDString())

	err := pm.SyncOplog(peer, pm.meOplogMerkle, SyncMeOplogMsg)
	if err != nil {
		return err
	}

	log.Debug("Sync: Done")

	return nil
}
