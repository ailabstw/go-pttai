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
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncUpdateUserNameAck struct {
	Objs []*UserName `json:"o"`
}

func (pm *ProtocolManager) HandleSyncUpdateUserNameAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	data := &SyncUpdateUserNameAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyUserName()
	pm.SetUserNameDB(origObj)
	for _, obj := range data.Objs {
		pm.SetUserNameDB(obj)

		pm.HandleSyncUpdateObjectAck(
			obj,
			peer,

			origObj,

			pm.userOplogMerkle,

			pm.SetUserDB,
			pm.updateSyncUserName,

			nil,
			pm.broadcastUserOplogCore,
		)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncUserName(theToSyncInfo pkgservice.SyncInfo, theFromObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {
	toSyncInfo, ok := theToSyncInfo.(*SyncUserNameInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*UserName)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// op-data
	opData := &UserOpUpdateUserName{}
	err := oplog.GetData(opData)
	if err != nil {
		return err
	}

	// logID
	toLogID := toSyncInfo.GetLogID()
	updateLogID := fromObj.GetUpdateLogID()

	if !reflect.DeepEqual(toLogID, updateLogID) {
		return pkgservice.ErrInvalidObject
	}

	// get name
	name := fromObj.Name

	hash := types.Hash(name)
	if !reflect.DeepEqual(opData.Hash, hash) {
		return pkgservice.ErrInvalidObject
	}

	toSyncInfo.Name = name

	return nil
}
