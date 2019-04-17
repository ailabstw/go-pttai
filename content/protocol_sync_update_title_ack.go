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

package content

import (
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncUpdateTitleAck struct {
	Objs []*Title `json:"o"`
}

func (pm *ProtocolManager) HandleSyncUpdateTitleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	data := &SyncUpdateTitleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyTitle()
	pm.SetTitleDB(origObj)
	for _, obj := range data.Objs {
		pm.SetTitleDB(obj)

		pm.HandleSyncUpdateObjectAck(
			obj,
			peer,

			origObj,

			pm.boardOplogMerkle,

			pm.SetBoardDB,
			pm.updateSyncTitle,
			nil,
			pm.broadcastBoardOplogCore,
		)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncTitle(theToSyncInfo pkgservice.SyncInfo, theFromObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {
	toSyncInfo, ok := theToSyncInfo.(*SyncTitleInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*Title)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// op-data
	opData := &BoardOpUpdateTitle{}
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

	// get title
	title := fromObj.Title

	hash := types.Hash(title)
	if !reflect.DeepEqual(opData.TitleHash, hash) {
		return pkgservice.ErrInvalidObject
	}

	toSyncInfo.Title = title

	return nil
}
