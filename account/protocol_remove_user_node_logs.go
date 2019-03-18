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

func (pm *ProtocolManager) handleRemoveUserNodeLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	opData := &UserOpRemoveUserNode{}

	log.Debug("handleRemoveUserNodeLog: to HandleDeleteObjectLog", "entity", pm.Entity().IDString(), "oplog", oplog)

	toBroadcastLogs, err := pm.HandleDeleteObjectLog(
		oplog,
		info,

		obj,
		opData,

		pm.userOplogMerkle,

		pm.SetUserDB,

		nil,
		pm.postdeleteUserNode,
		pm.updateDeleteUserNodeInfo,
	)
	log.Debug("handleRemoveUserNodeLog: after HandleDeleteObjectLog", "entity", pm.Entity().IDString(), "oplog", oplog, "e", err)
	if err != nil {
		return nil, err
	}

	return toBroadcastLogs, nil
}

func (pm *ProtocolManager) handlePendingRemoveUserNodeLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	opData := &UserOpRemoveUserNode{}

	return pm.HandlePendingDeleteObjectLog(
		oplog,
		info,

		obj,
		opData,

		pm.userOplogMerkle,

		pm.SetUserDB,

		nil,
		pm.setPendingDeleteUserNodeSyncInfo,
		pm.updateDeleteUserNodeInfo,
	)
}

func (pm *ProtocolManager) setNewestRemoveUserNodeLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	return pm.SetNewestDeleteObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedRemoveUserNodeLog(oplog *pkgservice.BaseOplog) error {
	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	return pm.HandleFailedDeleteObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedValidRemoveUserNodeLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) error {
	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	return pm.HandleFailedValidDeleteObjectLog(oplog, obj, info, pm.updateDeleteUserNodeInfo)
}

func (pm *ProtocolManager) updateDeleteUserNodeInfo(theUserNode pkgservice.Object, oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.CreateUserNodeInfo[*oplog.ObjID] = oplog
	info.UserNodeInfo[*oplog.ObjID] = oplog

	return nil
}
