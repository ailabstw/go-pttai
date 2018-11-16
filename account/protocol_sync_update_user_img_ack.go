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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncUpdateUserImgAck struct {
	Objs []*UserImg `json:"o"`
}

func (pm *ProtocolManager) HandleSyncUpdateUserImgAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	data := &SyncUpdateUserImgAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyUserImg()
	pm.SetUserImgDB(origObj)
	for _, obj := range data.Objs {
		pm.SetUserImgDB(obj)

		pm.HandleSyncUpdateObjectAck(
			obj,
			peer,

			origObj,
			pm.SetUserDB, pm.updateSyncUserImg, nil, pm.broadcastUserOplogCore)

		log.Debug("HandleSyncUpdateUserImgAck: after HandleSyncUpdateObjectAck", "obj.SyncInfo", obj.SyncInfo)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncUserImg(theToSyncInfo pkgservice.SyncInfo, theFromObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {
	toSyncInfo, ok := theToSyncInfo.(*SyncUserImgInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*UserImg)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// op-data
	opData := &UserOpUpdateUserImg{}
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

	// sync-info
	str := fromObj.Str

	hash := types.Hash([]byte(str))
	if !reflect.DeepEqual(opData.Hash, hash) {
		return pkgservice.ErrInvalidObject
	}

	toSyncInfo.ImgType = fromObj.ImgType
	toSyncInfo.Width = fromObj.Width
	toSyncInfo.Height = fromObj.Height
	toSyncInfo.Str = str

	return nil
}
