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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleUpdateUserImgLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	opData := &UserOpUpdateUserImg{}

	log.Debug("handleUpdateUserImgLogs: start", "oplog", oplog.ID, "objID", oplog.ObjID)

	return pm.HandleUpdateObjectLog(
		oplog,
		opData,

		obj,

		info,

		pm.userOplogMerkle,

		pm.syncUserImgInfoFromOplog,

		pm.SetUserDB,
		nil,

		nil,
		pm.updateUpdateUserImgInfo,
	)
}

func (pm *ProtocolManager) handlePendingUpdateUserImgLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	log.Debug("handlePendingUpdateUserImgLogs: start", "oplog", oplog.ID, "objID", oplog.ObjID)

	opData := &UserOpUpdateUserImg{}

	return pm.HandlePendingUpdateObjectLog(
		oplog,
		opData,

		obj,

		info,

		pm.userOplogMerkle,

		pm.syncUserImgInfoFromOplog,

		pm.SetUserDB,
		nil,

		nil,

		pm.updateUpdateUserImgInfo,
	)
}

func (pm *ProtocolManager) setNewestUpdateUserImgLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	return pm.SetNewestUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedUpdateUserImgLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	return pm.HandleFailedUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedValidUpdateUserImgLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) error {

	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	return pm.HandleFailedValidUpdateObjectLog(oplog, obj, info, pm.updateUpdateUserImgInfo)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) syncUserImgInfoFromOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	syncInfo := NewEmptySyncUserImgInfo()
	syncInfo.InitWithOplog(types.StatusInternalSync, oplog)

	return syncInfo, nil
}

func (pm *ProtocolManager) updateUpdateUserImgInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, opData pkgservice.OpData, origSyncInfo pkgservice.SyncInfo, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.UserImgInfo[*oplog.ObjID] = oplog

	return nil
}
