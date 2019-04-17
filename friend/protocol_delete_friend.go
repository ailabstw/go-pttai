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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) DeleteFriend() error {
	opData := &FriendOpDeleteFriend{}

	err := pm.DeleteEntity(
		FriendOpTypeDeleteFriend,
		opData,

		types.StatusInternalTerminal,
		types.StatusPendingTerminal,
		types.StatusTerminal,

		pm.friendOplogMerkle,

		pm.NewFriendOplog,
		pm.setPendingDeleteFriendSyncInfo,
		pm.broadcastFriendOplogCore,
		pm.postdeleteFriend,
	)

	log.Debug("DeleteFriend: after DeleteEntity", "e", err, "entity", pm.Entity().GetID())

	return err
}

func (pm *ProtocolManager) postdeleteFriend(theOpData pkgservice.OpData, isForce bool) error {

	// both are masters
	/*
		if !isForce {
			return nil
		}
	*/

	f := pm.Entity().(*Friend)
	friendID := f.FriendID

	myEntity := pm.Ptt().GetMyEntity()
	myProfilePM := myEntity.GetProfile().PM()
	myProfilePM.DeleteMember(friendID)

	myBoardPM := myEntity.GetBoard().PM()
	myBoardPM.DeleteMember(friendID)

	myID := myEntity.GetID()
	friendProfilePM := f.Profile.PM()
	friendProfilePM.DeleteMember(myID)

	friendBoardPM := f.Board.PM()
	friendBoardPM.DeleteMember(myID)

	pm.CleanObject()

	pm.DefaultPostdeleteEntity(theOpData, isForce)

	origFriend, err := pm.Entity().Service().SPM().(*ServiceProtocolManager).GetFriendByFriendID(friendID)
	log.Debug("postdeleteFriend: after GetFriendByFriendID", "e", err, "origID", origFriend.GetID(), "id", f.GetID(), "status", origFriend.Status)

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
