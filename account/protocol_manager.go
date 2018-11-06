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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProtocolManager struct {
	*pkgservice.BaseProtocolManager

	// db
	dbUserLock      *types.LockMap
	userOplogMerkle *pkgservice.Merkle
}

func NewProtocolManager(profile *Profile, ptt pkgservice.Ptt) (*ProtocolManager, error) {
	dbUserLock, err := types.NewLockMap(pkgservice.SleepTimeLock)
	if err != nil {
		return nil, err
	}

	userOplogMerkle, err := pkgservice.NewMerkle(DBUserOplogPrefix, DBUserMerkleOplogPrefix, profile.ID, dbAccount)
	if err != nil {
		return nil, err
	}

	pm := &ProtocolManager{
		dbUserLock:      dbUserLock,
		userOplogMerkle: userOplogMerkle,
	}
	b, err := pkgservice.NewBaseProtocolManager(
		ptt, RenewOpKeySeconds, ExpireOpKeySeconds, MaxSyncRandomSeconds, MinSyncRandomSeconds,
		nil, nil, nil, pm.SetUserDB,
		nil, nil, nil, nil, nil, nil, nil,
		pm.SyncUserOplog,
		nil,
		profile, dbAccount)
	if err != nil {
		return nil, err
	}
	pm.BaseProtocolManager = b

	return pm, nil
}

func (pm *ProtocolManager) Start() error {
	err := pm.BaseProtocolManager.Start()
	if err != nil {
		log.Error("Start: unable to start BaseProtocolManager", "e", err)
		return err
	}

	// oplog-merkle-tree
	go pkgservice.PMOplogMerkleTreeLoop(pm, pm.userOplogMerkle)

	return nil
}

func (pm *ProtocolManager) Stop() error {
	pm.BaseProtocolManager.PreStop()

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

	err := pm.SyncOplog(peer, pm.MasterMerkle(), pkgservice.SyncMasterOplogMsg)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) SyncUserOplog(peer *pkgservice.PttPeer) error {
	if peer == nil {
		return nil
	}

	log.Debug("SyncUserOplog: start")
	err := pm.SyncOplog(peer, pm.userOplogMerkle, SyncUserOplogMsg)
	log.Debug("SyncUserOplog: after SyncOplog", "e", err)
	if err != nil {
		return err
	}

	return nil
}
