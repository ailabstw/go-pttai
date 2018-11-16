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

type UpdateUserName struct {
	Name []byte `json:"n"`
}

func (pm *ProtocolManager) UpdateUserName(name []byte) (*UserName, error) {
	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	data := &UpdateUserName{Name: name}

	origObj := NewEmptyUserName()
	pm.SetUserNameDB(origObj)

	opData := &UserOpUpdateUserName{}

	log.Debug("UpdateUserName: to update UserName", "myID", myID, "name", name)

	err := pm.UpdateObject(
		myID, data, UserOpTypeUpdateUserName, origObj, opData,

		pm.SetUserDB, pm.NewUserOplog, pm.inupdateUserName, nil, pm.broadcastUserOplogCore, nil)
	if err != nil {
		return nil, err
	}

	return origObj, nil
}

func (pm *ProtocolManager) inupdateUserName(obj pkgservice.Object, theData pkgservice.UpdateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	data, ok := theData.(*UpdateUserName)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*UserOpUpdateUserName)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// op-data
	opData.Hash = types.Hash(data.Name)

	// sync-info
	syncInfo := NewEmptySyncUserNameInfo()
	syncInfo.InitWithOplog(oplog.ToStatus(), oplog)

	syncInfo.Name = data.Name

	return syncInfo, nil
}
