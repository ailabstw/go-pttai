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

func (pm *ProtocolManager) DeleteProfile() error {
	opData := &UserOpDeleteProfile{}

	log.Debug("DeleteProfile: start")

	err := pm.DeleteEntity(
		UserOpTypeDeleteProfile, opData,
		types.StatusInternalDeleted, types.StatusPendingDeleted, types.StatusDeleted,
		pm.NewUserOplog, pm.setPendingDeleteAccountSyncInfo, pm.broadcastUserOplogCore, pm.postdeleteProfile)

	log.Debug("DeleteProfile: after DeleteEntity", "e", err)
	return err
}

func (pm *ProtocolManager) postdeleteProfile(theOpData pkgservice.OpData, isForce bool) error {
	myID := pm.Ptt().GetMyEntity().GetID()

	log.Debug("postdeleteProfile: start", "isForce", isForce)

	if !isForce && pm.IsMaster(myID, false) {
		return nil
	}

	// user-oplog
	pm.CleanUserOplog()

	// user-node
	pm.CleanUserNode()

	pm.DefaultPostdeleteEntity(theOpData, isForce)

	return nil
}

func (pm *ProtocolManager) setPendingDeleteAccountSyncInfo(theEntity pkgservice.Entity, status types.Status, oplog *pkgservice.BaseOplog) error {

	entity, ok := theEntity.(*Profile)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithOplog(status, oplog)

	entity.SetSyncInfo(syncInfo)

	return nil
}
