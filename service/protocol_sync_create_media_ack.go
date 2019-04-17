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

package service

import "encoding/json"

type SyncMediaAck struct {
	Objs []*Media `json:"o"`
}

func (pm *BaseProtocolManager) HandleSyncCreateMediaAck(
	dataBytes []byte,
	peer *PttPeer,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	broadcastLog func(oplog *BaseOplog) error,

) error {

	data := &SyncMediaAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyMedia()
	pm.SetMediaDB(origObj)
	for _, obj := range data.Objs {
		pm.SetMediaDB(obj)

		pm.HandleSyncCreateObjectAck(
			obj,
			peer,

			origObj,

			merkle,

			setLogDB,
			pm.updateSyncCreateMedia,
			nil,
			broadcastLog,
		)
	}

	return nil
}

func (pm *BaseProtocolManager) updateSyncCreateMedia(theToObj Object, theFromObj Object) error {
	toObj, ok := theToObj.(*Media)
	if !ok {
		return ErrInvalidData
	}

	fromObj, ok := theFromObj.(*Media)
	if !ok {
		return ErrInvalidData
	}

	toObj.BlockInfo = fromObj.BlockInfo
	toObj.MediaType = fromObj.MediaType
	toObj.MediaData = fromObj.MediaData

	return nil
}
