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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleCreateTitleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	opData := &BoardOpCreateTitle{}

	log.Debug("handleCreateTitleLogs: to HandleCreateObjectLog", "oplog", oplog, "obj", oplog.ObjID)

	return pm.HandleCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateTitle, pm.newTitleWithOplog, nil, pm.updateCreateTitleInfo)
}

func (pm *ProtocolManager) handlePendingCreateTitleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	opData := &BoardOpCreateTitle{}

	log.Debug("handlePendingCreateTitleLogs: to HandleCreateObjectLog", "oplog", oplog, "obj", oplog.ObjID)

	return pm.HandlePendingCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateTitle, pm.newTitleWithOplog, nil, pm.updateCreateTitleInfo)
}

func (pm *ProtocolManager) setNewestCreateTitleLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedCreateTitleLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) newTitleWithOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) pkgservice.Object {

	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)
	pkgservice.NewObjectWithOplog(obj, oplog)

	return obj
}

func (pm *ProtocolManager) existsInInfoCreateTitle(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateTitleInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) updateCreateTitleInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.CreateTitleInfo[*oplog.ObjID] = oplog

	return nil
}
