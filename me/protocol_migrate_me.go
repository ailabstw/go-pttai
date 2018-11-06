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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) MigrateMe(newMyInfo *MyInfo) error {
	opData := &MeOpMigrateMe{ID: newMyInfo.ID}

	return pm.DeleteEntity(MeOpTypeMigrateMe, opData, types.StatusMigrated, pm.NewMeOplog, pm.broadcastMeOplogCore, pm.postdeleteMigrateMe)
}

/*
postdeleteMigrateMe deals with ops after deletingMigrateMe. Assuming entity already locked (in DeleteEntity and DeleteEntityLogs).
*/
func (pm *ProtocolManager) postdeleteMigrateMe(theOpData pkgservice.OpData) (err error) {
	myInfo := pm.Entity().(*MyInfo)

	opData, ok := theOpData.(*MeOpMigrateMe)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	log.Debug("postdeleteMigrateMe: to AddOwnerID", "e", pm.Entity().GetID(), "opData", opData.ID, "ownerIDs", pm.Entity().GetOwnerIDs(), "myOwnerIDs", myInfo.GetOwnerIDs())

	myInfo.AddOwnerID(opData.ID)

	log.Debug("postdeleteMigrateMe: after AddOwnerID", "e", pm.Entity().GetID(), "opData", opData.ID, "ownerIDs", pm.Entity().GetOwnerIDs())

	return myInfo.Save(true)
}
