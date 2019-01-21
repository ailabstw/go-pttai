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

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleCreateUserNameLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	opData := &UserOpCreateUserName{}

	return pm.HandleCreateObjectLog(
		oplog,
		obj,

		opData,
		info,

		pm.existsInInfoCreateUserName,
		pm.newUserNameWithOplog,
		nil,
		pm.updateCreateUserNameInfo,
	)
}

func (pm *ProtocolManager) handlePendingCreateUserNameLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	opData := &UserOpCreateUserName{}

	return pm.HandlePendingCreateObjectLog(
		oplog,
		obj,
		opData,
		info,

		pm.existsInInfoCreateUserName,
		pm.newUserNameWithOplog,
		nil,
		pm.updateCreateUserNameInfo,
	)
}

func (pm *ProtocolManager) setNewestCreateUserNameLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedCreateUserNameLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

func (pm *ProtocolManager) handleFailedValidCreateUserNameLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) error {

	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)

	return pm.HandleFailedValidCreateObjectLog(oplog, obj, nil)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) newUserNameWithOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) pkgservice.Object {

	obj := NewEmptyUserName()
	pm.SetUserNameDB(obj)
	pkgservice.NewObjectWithOplog(obj, oplog)

	return obj
}

func (pm *ProtocolManager) existsInInfoCreateUserName(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateUserNameInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) updateCreateUserNameInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.CreateUserNameInfo[*oplog.ObjID] = oplog

	return nil
}
