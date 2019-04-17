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

func (pm *ProtocolManager) HandleForceSyncArticleAck(
	dataBytes []byte,
	peer *pkgservice.PttPeer,
) error {

	data := &SyncArticleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyArticle()
	pm.SetArticleDB(origObj)

	lenObj := len(data.Objs)

	blockIDs := make([]*pkgservice.SyncBlockID, 0, lenObj)
	mediaIDs := make([]*pkgservice.ForceSyncID, 0, lenObj)
	articleIDs := make([]*types.PttID, 0, lenObj)
	var blockInfo *pkgservice.BlockInfo
	var logID *types.PttID
	for _, obj := range data.Objs {
		pm.SetArticleDB(obj)

		err = pm.HandleForceSyncObjectAck(
			obj,
			peer,

			origObj,

			pm.boardOplogMerkle,

			pm.SetBoardDB,
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

		// block-ids
		blockIDs = append(blockIDs, &pkgservice.SyncBlockID{ID: blockInfo.ID, ObjID: obj.ID, LogID: logID})

		// media-ids
		if blockInfo.MediaIDs != nil {
			for _, eachID := range blockInfo.MediaIDs {
				mediaIDs = append(mediaIDs, &pkgservice.ForceSyncID{ID: eachID, TS: types.MaxTimestamp})
			}
		}

		// article-ids
		if reflect.DeepEqual(origObj.ID, obj.ID) && handleForceSyncArticleAckIsDeleted(origObj, pm.SetBoardDB) {
			articleIDs = append(articleIDs, obj.ID)
		}

	}

	if len(blockIDs) != 0 {
		pm.SyncBlock(SyncCreateArticleBlockMsg, blockIDs, peer)
	}

	if len(mediaIDs) != 0 {
		pm.ForceSyncMedia(mediaIDs, peer, ForceSyncMediaMsg)
	}

	if len(articleIDs) != 0 {
		pm.ForceSyncArticleCommentList(articleIDs, peer)
	}

	return nil
}

func handleForceSyncArticleAckIsDeleted(
	obj *Article,

	setLogDB func(oplog *pkgservice.BaseOplog),
) bool {

	logID := obj.LogID
	oplog := &pkgservice.BaseOplog{ID: logID}
	setLogDB(oplog)

	err := oplog.Lock()
	if err != nil {
		return true
	}
	defer oplog.Unlock()

	err = oplog.Get(logID, true)
	if err != nil {
		return true
	}

	return false
}
