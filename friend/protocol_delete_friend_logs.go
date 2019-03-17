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

func (pm *ProtocolManager) handleDeleteFriendLogs(oplog *pkgservice.BaseOplog, info *ProcessFriendInfo) ([]*pkgservice.BaseOplog, error) {

	opData := &FriendOpDeleteFriend{}

	log.Debug("handleDeleteFriendLogs: start", "entity", pm.Entity().IDString())

	return pm.HandleDeleteEntityLog(
		oplog,
		info,

		opData,
		types.StatusTerminal,

		pm.friendOplogMerkle,

		pm.SetFriendDB,
		nil,
		pm.updateFriendDeleteInfo,
	)
}

func (pm *ProtocolManager) handlePendingDeleteFriendLogs(oplog *pkgservice.BaseOplog, info *ProcessFriendInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	opData := &FriendOpDeleteFriend{}

	return pm.HandlePendingDeleteEntityLog(
		oplog,
		info,

		types.StatusInternalTerminal,
		types.StatusPendingTerminal,
		FriendOpTypeDeleteFriend,
		opData,

		pm.friendOplogMerkle,

		pm.SetFriendDB,
		pm.setPendingDeleteFriendSyncInfo,
		pm.updateFriendDeleteInfo,
	)
}

func (pm *ProtocolManager) setNewestDeleteFriendLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {

	return false, nil
}

func (pm *ProtocolManager) handleFailedDeleteFriendLog(oplog *pkgservice.BaseOplog) error {

	return pm.HandleFailedDeleteEntityLog(oplog)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) updateFriendDeleteInfo(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessFriendInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.FriendInfo[*oplog.ObjID] = oplog

	return nil
}
