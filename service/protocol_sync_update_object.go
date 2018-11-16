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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) HandleSyncUpdateObject(
	dataBytes []byte,
	peer *PttPeer,
	obj Object,
	syncAckMsg OpType,
) error {

	data := &SyncObject{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	lenObjs := len(data.IDs)
	objs := make([]Object, 0, lenObjs)
	var syncInfo SyncInfo
	var blockInfo *BlockInfo

	log.Debug("HandleSyncUpdateObject: start")
	for _, syncID := range data.IDs {
		newObj, err := obj.GetNewObjByID(syncID.ID, false)
		if err != nil {
			continue
		}

		// deleted content
		if newObj.GetStatus() > types.StatusAlive {
			continue
		}

		// in the obj.
		if reflect.DeepEqual(newObj.GetUpdateLogID(), syncID.LogID) {
			newObj.SetSyncInfo(nil)
			objs = append(objs, newObj)
			continue
		}

		// check sync-info
		syncInfo = newObj.GetSyncInfo()
		if syncInfo == nil {
			continue
		}

		if !syncInfo.GetIsGood() {
			continue
		}

		if !reflect.DeepEqual(syncInfo.GetLogID(), syncID.LogID) {
			continue
		}

		blockInfo = syncInfo.GetBlockInfo()
		if blockInfo != nil {
			blockInfo.ResetIsGood()
		}

		// We let the sync-info be in the main-meta of the newobj.
		// To simplify the process in HandleSyncUpdateObjectAck. (no need to worry whether it's from main-meta or from sync-info.)
		err = syncInfo.ToObject(newObj)
		if err != nil {
			continue
		}

		objs = append(objs, newObj)
	}

	log.Debug("HandleSyncUpdateObject: to SyncObjectAck", "objs", objs)

	return pm.SyncObjectAck(objs, syncAckMsg, peer)
}
