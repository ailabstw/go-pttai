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

import (
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type SyncObject struct {
	IDs []*SyncID
}

func (pm *BaseProtocolManager) SyncObject(op OpType, syncIDs []*SyncID, peer *PttPeer) error {
	if len(syncIDs) == 0 {
		return nil
	}

	pSyncIDs := syncIDs
	var eachSyncIDs []*SyncID
	lenEachSyncIDs := 0
	var data *SyncObject
	for len(pSyncIDs) > 0 {
		lenEachSyncIDs = MaxSyncObjectAck
		if lenEachSyncIDs > len(pSyncIDs) {
			lenEachSyncIDs = len(pSyncIDs)
		}

		eachSyncIDs, pSyncIDs = pSyncIDs[:lenEachSyncIDs], pSyncIDs[lenEachSyncIDs:]

		data = &SyncObject{
			IDs: eachSyncIDs,
		}

		err := pm.SendDataToPeer(op, data, peer)
		if err != nil {
			return err
		}

	}
	return nil
}

func (pm *BaseProtocolManager) HandleSyncCreateObject(
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
	if lenObjs == 0 {
		return nil
	}

	objs := make([]Object, 0, lenObjs)

	log.Debug("HandleSyncCreateObject: start")
	var blockInfo *BlockInfo
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

		blockInfo = newObj.GetBlockInfo()
		if blockInfo != nil {
			blockInfo.ResetIsGood()
		}
		log.Debug("HandleSyncCreateObject: (in for-loop)", "blockInfo", blockInfo)

		newObj.SetSyncInfo(nil)

		objs = append(objs, newObj)
	}

	log.Debug("HandleSyncCreateObject: to SyncObjectAck", "objs", objs)

	return pm.SyncObjectAck(objs, syncAckMsg, peer)
}
