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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProtocolManager struct {
	*pkgservice.BaseProtocolManager

	// db
	dbFriendLock      *types.LockMap
	friendOplogMerkle *pkgservice.Merkle

	// message
	dbMessagePrefix    []byte
	dbMessageIdxPrefix []byte
}

func NewProtocolManager(f *Friend, ptt pkgservice.Ptt) (*ProtocolManager, error) {
	dbFriendLock, err := types.NewLockMap(pkgservice.SleepTimeLock)
	if err != nil {
		return nil, err
	}

	friendOplogMerkle, err := pkgservice.NewMerkle(DBFriendOplogPrefix, DBFriendMerkleOplogPrefix, f.ID, dbFriend)
	if err != nil {
		return nil, err
	}
	pm := &ProtocolManager{
		dbFriendLock:      dbFriendLock,
		friendOplogMerkle: friendOplogMerkle,
	}
	b, err := pkgservice.NewBaseProtocolManager(
		ptt,

		RenewOpKeySeconds,
		ExpireOpKeySeconds,
		MaxSyncRandomSeconds,
		MinSyncRandomSeconds,

		MaxMasters,

		pm.friendOplogMerkle, // log0Merkle

		// sign
		nil,
		nil,
		nil,

		pm.SetFriendDB, // setLog0DB

		nil, // isMaster
		nil,

		// peer-type
		nil,
		nil,
		nil,
		nil,
		nil,

		pm.SyncFriendOplog, // postsyncMemberOplog

		pm.DeleteFriend,     // theDelete
		pm.postdeleteFriend, // postdelete

		f, // entity

		dbFriend, // db
	)
	if err != nil {
		return nil, err
	}
	pm.BaseProtocolManager = b

	// message
	entityID := f.ID
	pm.dbMessagePrefix = append(DBMessagePrefix, entityID[:]...)
	pm.dbMessageIdxPrefix = append(DBMessageIdxPrefix, entityID[:]...)

	return pm, nil
}

func (pm *ProtocolManager) Start() error {
	err := pm.BaseProtocolManager.Start()
	if err != nil {
		log.Error("Start: unable to start BaseProtocolManager", "e", err)
		return err
	}

	pm.LoadPeers()

	// oplog-merkle-tree
	syncWG := pm.SyncWG()
	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pkgservice.PMOplogMerkleTreeLoop(pm, pm.friendOplogMerkle)
	}()

	return nil
}

func (pm *ProtocolManager) Stop() error {

	return nil
}

func (pm *ProtocolManager) Sync(peer *pkgservice.PttPeer) error {
	log.Debug("Sync: start", "entity", pm.Entity().GetID(), "peer", peer, "service", pm.Entity().Service().Name(), "status", pm.Entity().GetStatus())
	if peer == nil {
		pm.SyncPendingMasterOplog(peer)
		pm.SyncPendingMemberOplog(peer)
		pm.SyncPendingFriendOplog(peer)
		return nil
	}

	err := pm.SyncOplog(peer, pm.MasterMerkle(), pkgservice.SyncMasterOplogMsg)

	log.Debug("Sync: after SyncOplog", "entity", pm.Entity().GetID(), "peer", peer, "service", pm.Entity().Service().Name(), "e", err)

	if err != nil {
		return err
	}

	return nil
}
