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

func (pm *ProtocolManager) handleCreateMessageLogs(oplog *pkgservice.BaseOplog, info *ProcessFriendInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	opData := &FriendOpCreateMessage{}

	return pm.HandleCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateMessage, pm.newMessageWithOplog, pm.postcreateMessage, pm.updateCreateMessageInfo)
}

func (pm *ProtocolManager) handlePendingCreateMessageLogs(oplog *pkgservice.BaseOplog, info *ProcessFriendInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	opData := &FriendOpCreateMessage{}

	log.Debug("handlePendingCreateMessageLogs: start", "oplog", oplog.ID, "objID", oplog.ObjID)

	return pm.HandlePendingCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateMessage, pm.newMessageWithOplog, pm.postcreateMessage, pm.updateCreateMessageInfo)
}

func (pm *ProtocolManager) setNewestCreateMessageLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedCreateMessageLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

func (pm *ProtocolManager) handleFailedValidCreateMessageLog(oplog *pkgservice.BaseOplog, info *ProcessFriendInfo) error {

	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	return pm.HandleFailedValidCreateObjectLog(oplog, obj, nil)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) newMessageWithOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) pkgservice.Object {

	opData, ok := theOpData.(*FriendOpCreateMessage)
	if !ok {
		return nil
	}

	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)
	pkgservice.NewObjectWithOplog(obj, oplog)

	blockInfo, err := pkgservice.NewBlockInfo(opData.BlockInfoID, opData.Hashs, opData.MediaIDs, oplog.CreatorID)
	if err != nil {
		return nil
	}
	pm.SetBlockInfoDB(blockInfo, obj.ID)
	blockInfo.InitIsGood()
	obj.SetBlockInfo(blockInfo)

	return obj
}

func (pm *ProtocolManager) existsInInfoCreateMessage(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessFriendInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateMessageInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) updateCreateMessageInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessFriendInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	blockInfo := obj.GetBlockInfo()
	if blockInfo == nil {
		return pkgservice.ErrInvalidData
	}

	info.CreateMessageInfo[*oplog.ObjID] = oplog
	info.BlockInfo[*blockInfo.ID] = oplog

	return nil
}
