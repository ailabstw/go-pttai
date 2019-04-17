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

	"github.com/ailabstw/go-pttai/common/types"
)

func (pm *BaseProtocolManager) HandleForceSyncMediaAck(
	dataBytes []byte,
	peer *PttPeer,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	syncMediaBlockMsg OpType,
) error {

	data := &SyncMediaAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyMedia()
	pm.SetMediaDB(origObj)

	blockIDs := make([]*SyncBlockID, 0, len(data.Objs))
	var blockInfo *BlockInfo
	var logID *types.PttID
	for _, obj := range data.Objs {
		pm.SetMediaDB(obj)

		err = pm.HandleForceSyncObjectAck(
			obj,
			peer,

			origObj,

			merkle,

			setLogDB,
		)

		if err != nil {
			continue
		}

		if obj.GetStatus() >= types.StatusDeleted {
			continue
		}

		blockInfo = obj.GetBlockInfo()

		logID = obj.LogID
		if obj.GetUpdateLogID() != nil {
			logID = obj.GetUpdateLogID()
		}

		blockIDs = append(blockIDs, &SyncBlockID{ID: blockInfo.ID, ObjID: obj.ID, LogID: logID})

	}

	if len(blockIDs) != 0 {
		pm.SyncBlock(syncMediaBlockMsg, blockIDs, peer)
	}

	return nil
}
