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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleUpdateTitleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	opData := &BoardOpUpdateTitle{}

	return pm.HandleUpdateObjectLog(
		oplog,
		opData,

		obj,
		info,

		pm.boardOplogMerkle,

		pm.syncTitleInfoFromOplog,
		pm.SetBoardDB,
		nil,
		nil,
		pm.updateUpdateTitleInfo,
	)
}

func (pm *ProtocolManager) handlePendingUpdateTitleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	opData := &BoardOpUpdateTitle{}

	return pm.HandlePendingUpdateObjectLog(
		oplog,
		opData,

		obj,
		info,
		pm.boardOplogMerkle,

		pm.syncTitleInfoFromOplog,
		pm.SetBoardDB,
		nil,
		nil,
		pm.updateUpdateTitleInfo,
	)
}

func (pm *ProtocolManager) setNewestUpdateTitleLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	return pm.SetNewestUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedUpdateTitleLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	return pm.HandleFailedUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedValidUpdateTitleLog(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) error {

	obj := NewEmptyTitle()
	pm.SetTitleDB(obj)

	return pm.HandleFailedValidUpdateObjectLog(oplog, obj, info, pm.updateUpdateTitleInfo)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) syncTitleInfoFromOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	syncInfo := NewEmptySyncTitleInfo()
	syncInfo.InitWithOplog(types.StatusInternalSync, oplog)

	return syncInfo, nil
}

func (pm *ProtocolManager) updateUpdateTitleInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, opData pkgservice.OpData, origSyncInfo pkgservice.SyncInfo, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.TitleInfo[*oplog.ObjID] = oplog

	return nil
}
