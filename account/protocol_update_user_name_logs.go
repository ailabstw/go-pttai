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

func (pm *ProtocolManager) handleUpdateUserNameLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	opData := &UserOpUpdateUserName{}

	log.Debug("handleUpdateUserNameLogs: to HandleUpdateObjectLog")

	return pm.HandleUpdateObjectLog(
		oplog, opData, obj, info,
		pm.syncUserNameInfoFromOplog, pm.SetUserDB, nil, nil, pm.updateUpdateUserNameInfo)
}

func (pm *ProtocolManager) handlePendingUpdateUserNameLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	opData := &UserOpUpdateUserName{}

	return pm.HandlePendingUpdateObjectLog(oplog, opData, obj, info, pm.syncUserNameInfoFromOplog, pm.SetUserDB, nil, nil, pm.updateUpdateUserNameInfo)
}

func (pm *ProtocolManager) setNewestUpdateUserNameLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	return pm.SetNewestUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedUpdateUserNameLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	return pm.HandleFailedUpdateObjectLog(oplog, obj)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) syncUserNameInfoFromOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	syncInfo := NewEmptySyncUserNameInfo()
	syncInfo.InitWithOplog(types.StatusInternalSync, oplog)

	return syncInfo, nil
}

func (pm *ProtocolManager) updateUpdateUserNameInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, opData pkgservice.OpData, origSyncInfo pkgservice.SyncInfo, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.UserNameInfo[*oplog.ObjID] = oplog

	return nil
}
