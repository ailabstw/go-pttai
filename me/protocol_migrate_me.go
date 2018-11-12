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
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) MigrateMe(newMyInfo *MyInfo) error {
	opData := &MeOpMigrateMe{ID: newMyInfo.ID}

	return pm.DeleteEntity(
		MeOpTypeMigrateMe, opData,
		types.StatusInternalDeleted, types.StatusPendingMigrate, types.StatusMigrated,
		pm.NewMeOplog, pm.setPendingDeleteMeSyncInfo, pm.broadcastMeOplogCore, pm.postdeleteMigrateMe)
}

/*
postdeleteMigrateMe deals with ops after deletingMigrateMe. Assuming entity already locked (in DeleteEntity and DeleteEntityLogs).
*/
func (pm *ProtocolManager) postdeleteMigrateMe(theOpData pkgservice.OpData, isForce bool) error {
	opData, ok := theOpData.(*MeOpMigrateMe)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	log.Debug("postdeleteMigrateMe: start")

	myInfo := pm.Entity().(*MyInfo)
	myID := myInfo.ID

	newMyID := opData.ID

	var err error
	var entityPM pkgservice.ProtocolManager

	// add member
	entities := pm.myPtt.GetEntities()
	for _, entity := range entities {
		if entity == myInfo {
			continue
		}

		log.Debug("postdeleteMigrateMe: (in-for-loop)", "entity", entity.GetID(), "service", entity.Service().Name())

		if !entity.IsOwner(myID) {
			log.Debug("postdeleteMigrateMe: not owner", "entity", entity.GetID(), "owners", entity.GetOwnerIDs()[0])
			continue
		}

		entityPM = entity.PM()
		_, _, err = entityPM.AddMember(newMyID, true)
		log.Debug("postdeleteMigrateMe: after add member", "entity", entity.GetID(), "e", err)
		if err != nil {
			continue
		}
	}

	// delete user-profile
	myInfo.Profile.PM().(*account.ProtocolManager).Delete()

	// transfer
	for _, entity := range entities {
		if entity == myInfo {
			continue
		}

		if !entity.IsOwner(myID) {
			log.Debug("postdeleteMigrateMe: not owner", "entity", entity.GetID(), "owners", entity.GetOwnerIDs()[0])
			continue
		}

		entityPM = entity.PM()

		if entityPM.IsMaster(myID, false) {
			err = entityPM.TransferMaster(newMyID)
			log.Debug("postdeleteMigrateMe: after transfer master", "entity", entity.GetID(), "e", err)
			if err != nil {
				continue
			}
		}

		err = entityPM.TransferMember(myID, newMyID)
		log.Debug("postdeleteMigrateMe: after transfer member", "entity", entity.GetID(), "e", err)
		if err != nil {
			continue
		}
	}

	log.Debug("postdeleteMigrateMe: after for-loop")

	myInfo.AddOwnerID(newMyID)

	return myInfo.Save(true)
}

func (pm *ProtocolManager) setPendingDeleteMeSyncInfo(theMyInfo pkgservice.Entity, status types.Status, oplog *pkgservice.BaseOplog) error {

	myInfo, ok := theMyInfo.(*MyInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithDeleteOplog(status, oplog)

	myInfo.SetSyncInfo(syncInfo)

	return nil
}
