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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) DeleteFriend() error {
	opData := &FriendOpDeleteFriend{}

	err := pm.DeleteEntity(
		FriendOpTypeDeleteFriend,
		opData,

		types.StatusInternalDeleted,
		types.StatusPendingDeleted,
		types.StatusDeleted,

		pm.friendOplogMerkle,

		pm.NewFriendOplog,
		pm.setPendingDeleteFriendSyncInfo,
		pm.broadcastFriendOplogCore,
		pm.postdeleteFriend,
	)

	return err
}

func (pm *ProtocolManager) postdeleteFriend(theOpData pkgservice.OpData, isForce bool) error {

	// both are masters
	if !isForce {
		return nil
	}

	// friend-oplog
	pm.CleanFriendOplog()

	pm.DefaultPostdeleteEntity(theOpData, isForce)

	return nil
}

func (pm *ProtocolManager) setPendingDeleteFriendSyncInfo(theEntity pkgservice.Entity, status types.Status, oplog *pkgservice.BaseOplog) error {

	entity, ok := theEntity.(*Friend)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithOplog(status, oplog)

	entity.SetSyncInfo(syncInfo)

	return nil
}
