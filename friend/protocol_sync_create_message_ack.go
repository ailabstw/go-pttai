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

package friend

import (
	"encoding/json"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncMessageAck struct {
	Objs []*Message `json:"o"`
}

func (pm *ProtocolManager) HandleSyncCreateMessageAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	data := &SyncMessageAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyMessage()
	pm.SetMessageDB(origObj)
	for _, obj := range data.Objs {
		pm.SetMessageDB(obj)

		pm.HandleSyncCreateObjectAck(
			obj,
			peer,
			origObj,

			pm.friendOplogMerkle,

			pm.SetFriendDB,
			pm.updateSyncCreateMessage,
			pm.postcreateMessage,
			pm.broadcastFriendOplogCore,
		)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncCreateMessage(theToObj pkgservice.Object, theFromObj pkgservice.Object) error {
	toObj, ok := theToObj.(*Message)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*Message)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	toObj.BlockInfo = fromObj.BlockInfo

	return nil
}
