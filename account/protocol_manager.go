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

package account

import (
	"sync"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProtocolManager struct {
	*pkgservice.BaseProtocolManager

	// db
	dbUserLock      *types.LockMap
	userOplogMerkle *pkgservice.Merkle

	// user-node
	dbUserNodePrefix     []byte
	dbUserNodeIdxPrefix  []byte
	dbUserNodeIdx2Prefix []byte

	lockUserNodeInfo sync.RWMutex
	userNodeInfo     *UserNodeInfo

	// user-name
	dbUserNamePrefix    []byte
	dbUserNameIdxPrefix []byte

	// user-img
	dbUserImgPrefix    []byte
	dbUserImgIdxPrefix []byte

	// name-card
	dbNameCardPrefix    []byte
	dbNameCardIdxPrefix []byte
}

func newBaseProtocolManager(pm *ProtocolManager, ptt pkgservice.Ptt, entity pkgservice.Entity) *pkgservice.BaseProtocolManager {

	b, err := pkgservice.NewBaseProtocolManager(
		ptt,

		RenewOpKeySeconds,
		ExpireOpKeySeconds,
		MaxSyncRandomSeconds,
		MinSyncRandomSeconds,

		MaxMasters,

		pm.userOplogMerkle, // log0Merkle

		// sign
		nil,
		nil,
		nil,
		nil,

		pm.SetUserDB,        // setLog0DB
		pm.HandleUserOplogs, // handleLog0s

		nil, // isMaster
		nil, // isMember

		// peer-type
		nil,
		nil,
		nil,
		nil,
		nil,

		pm.SyncUserOplog, // postsyncMemberOplog

		pm.DeleteProfile,     // theDelete
		pm.postdeleteProfile, // postdelete

		entity, // entity

		dbAccount, //db
	)
	if err != nil {
		return nil
	}

	return b
}

func NewProtocolManager(profile *Profile, ptt pkgservice.Ptt) (*ProtocolManager, error) {
	dbUserLock, err := types.NewLockMap(pkgservice.SleepTimeLock)
	if err != nil {
		return nil, err
	}

	userOplogMerkle, err := pkgservice.NewMerkle(DBUserOplogPrefix, DBUserMerkleOplogPrefix, profile.ID, dbAccount, "user")
	if err != nil {
		return nil, err
	}

	pm := &ProtocolManager{
		dbUserLock:      dbUserLock,
		userOplogMerkle: userOplogMerkle,
	}
	pm.BaseProtocolManager = newBaseProtocolManager(pm, ptt, profile)

	// user-node
	entityID := profile.ID
	pm.dbUserNodePrefix = append(DBUserNodePrefix, entityID[:]...)
	pm.dbUserNodeIdxPrefix = append(DBUserNodeIdxPrefix, entityID[:]...)
	pm.dbUserNodeIdx2Prefix = common.CloneBytes(pm.dbUserNodeIdxPrefix)
	pm.dbUserNodeIdx2Prefix[pttdb.SizeDBKeyPrefix-1] = '2'

	pm.InitUserNode(entityID)

	// user-name
	pm.dbUserNamePrefix = DBUserNamePrefix
	pm.dbUserNameIdxPrefix = DBUserNameIdxPrefix

	// user-img
	pm.dbUserImgPrefix = DBUserImgPrefix
	pm.dbUserImgIdxPrefix = DBUserImgIdxPrefix

	// name-card
	pm.dbNameCardPrefix = DBNameCardPrefix
	pm.dbNameCardIdxPrefix = DBNameCardIdxPrefix

	return pm, nil
}

func (pm *ProtocolManager) Start() error {
	err := pm.BaseProtocolManager.Start()
	if err != nil {
		log.Error("Start: unable to start BaseProtocolManager", "e", err)
		return err
	}

	// oplog-merkle-tree
	syncWG := pm.SyncWG()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pkgservice.PMOplogMerkleTreeLoop(pm, pm.userOplogMerkle)
	}()

	return nil
}

func (pm *ProtocolManager) Stop() error {

	return nil
}

func (pm *ProtocolManager) Sync(peer *pkgservice.PttPeer) error {
	log.Debug("Sync: start", "entity", pm.Entity().IDString(), "peer", peer, "status", pm.Entity().GetStatus())
	if peer == nil {
		pm.SyncPendingMasterOplog(peer)
		pm.SyncPendingMemberOplog(peer)
		pm.SyncPendingUserOplog(peer)
		return nil
	}

	err := pm.SyncOplog(peer, pm.MasterMerkle(), pkgservice.SyncMasterOplogMsg)

	log.Debug("Sync: after SyncOplog", "entity", pm.Entity().IDString(), "peer", peer, "e", err)

	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) GetUserNodeInfo() *UserNodeInfo {
	return pm.userNodeInfo
}
