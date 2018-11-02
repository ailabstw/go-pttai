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

package service

import (
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
)

type SyncCreateObject struct {
	IDs []*SyncID
}

func (pm *BaseProtocolManager) SyncCreateObject(op OpType, syncIDs []*SyncID, peer *PttPeer) error {
	err := pm.SendDataToPeer(op, &SyncCreateObject{IDs: syncIDs}, peer)
	if err != nil {
		return err
	}
	return nil
}

func (pm *BaseProtocolManager) HandleSyncCreateObject(
	dataBytes []byte,
	peer *PttPeer,
	obj Object,
	syncAckMsg OpType,
) error {
	data := &SyncCreateObject{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	lenObjs := len(data.IDs)
	objs := make([]Object, 0, lenObjs)
	for _, syncID := range data.IDs {
		newObj, err := obj.GetNewObjByID(syncID.ID, false)
		if err != nil {
			continue
		}

		if newObj.GetStatus() == types.StatusInternalSync {
			continue
		}

		if newObj.GetUpdateLogID() != nil { // with updated content
			continue
		}

		if !reflect.DeepEqual(syncID.LogID, newObj.GetLogID()) { // deleted content
			continue
		}

		newObj.SetSyncInfo(nil)

		objs = append(objs, newObj)
	}

	return pm.SyncCreateObjectAck(objs, syncAckMsg, peer)
}
