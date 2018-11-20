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
	"encoding/json"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncTitleAck struct {
	Objs []*Title `json:"o"`
}

func (pm *ProtocolManager) HandleSyncCreateTitleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	data := &SyncTitleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	if len(data.Objs) == 0 {
		return nil
	}

	origObj := NewEmptyTitle()
	pm.SetTitleDB(origObj)
	for _, obj := range data.Objs {
		pm.SetTitleDB(obj)

		pm.HandleSyncCreateObjectAck(
			obj, peer, origObj,
			pm.SetBoardDB, pm.updateSyncCreateTitle, nil, pm.broadcastBoardOplogCore)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncCreateTitle(theToObj pkgservice.Object, theFromObj pkgservice.Object) error {
	toObj, ok := theToObj.(*Title)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*Title)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	toObj.Title = fromObj.Title

	return nil
}
