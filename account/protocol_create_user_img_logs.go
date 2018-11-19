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

func (pm *ProtocolManager) handleCreateUserImgLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	opData := &UserOpCreateUserImg{}

	return pm.HandleCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateUserImg, pm.newUserImgWithOplog, nil, pm.updateCreateUserImgInfo)
}

func (pm *ProtocolManager) handlePendingCreateUserImgLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	opData := &UserOpCreateUserImg{}

	return pm.HandlePendingCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateUserImg, pm.newUserImgWithOplog, nil, pm.updateCreateUserImgInfo)
}

func (pm *ProtocolManager) setNewestCreateUserImgLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedCreateUserImgLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) newUserImgWithOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) pkgservice.Object {

	obj := NewEmptyUserImg()
	pm.SetUserImgDB(obj)
	pkgservice.NewObjectWithOplog(obj, oplog)

	return obj
}

func (pm *ProtocolManager) existsInInfoCreateUserImg(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateUserImgInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) updateCreateUserImgInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.CreateUserImgInfo[*oplog.ObjID] = oplog

	return nil
}
