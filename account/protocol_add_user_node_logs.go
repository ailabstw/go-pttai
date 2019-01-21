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

func (pm *ProtocolManager) handleAddUserNodeLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {

	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	opData := &UserOpAddUserNode{}

	log.Debug("handleAddUserNodeLog: start", "oplog", oplog.ID)

	return pm.HandleCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoAddUserNode, pm.newUserNodeWithOplog, pm.postcreateUserNode, pm.updateAddUserNodeInfo)
}

func (pm *ProtocolManager) handlePendingAddUserNodeLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	opData := &UserOpAddUserNode{}

	log.Debug("handlePendingAddUserNodeLog: start", "oplog", oplog.ID)

	return pm.HandlePendingCreateObjectLog(oplog, obj, opData, info, pm.existsInInfoAddUserNode, pm.newUserNodeWithOplog, pm.postcreateUserNode, pm.updateAddUserNodeInfo)
}

func (pm *ProtocolManager) setNewestAddUserNodeLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedAddUserNodeLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

func (pm *ProtocolManager) handleFailedValidAddUserNodeLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) error {

	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	return pm.HandleFailedValidCreateObjectLog(oplog, obj, nil)
}

/**********
 * Customize
 **********/

/***
 * handleCreateObject
 ***/

func (pm *ProtocolManager) existsInInfoAddUserNode(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateUserNodeInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) newUserNodeWithOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) pkgservice.Object {
	opData := theOpData.(*UserOpAddUserNode)
	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)
	pkgservice.NewObjectWithOplog(obj, oplog)

	obj.NodeID = opData.NodeID
	obj.UserID = pm.Entity().(*Profile).MyID

	return obj
}

func (pm *ProtocolManager) updateAddUserNodeInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.CreateUserNodeInfo[*oplog.ObjID] = oplog

	return nil
}
