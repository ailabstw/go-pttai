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

type SyncUpdateArticleAck struct {
	Objs []*Article `json:"o"`
}

func (pm *ProtocolManager) HandleSyncUpdateArticleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	data := &SyncUpdateArticleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyArticle()
	pm.SetArticleDB(origObj)
	for _, obj := range data.Objs {
		pm.SetArticleDB(obj)

		pm.HandleSyncUpdateObjectAck(
			obj,
			peer,

			origObj,

			pm.boardOplogMerkle,

			pm.SetBoardDB,
			pm.updateSyncArticle,
			nil,
			pm.broadcastBoardOplogCore,
		)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncArticle(theToSyncInfo pkgservice.SyncInfo, theFromObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {
	toSyncInfo, ok := theToSyncInfo.(*SyncArticleInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*Article)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// op-data
	opData := &BoardOpUpdateArticle{}
	err := oplog.GetData(opData)
	if err != nil {
		return err
	}

	// validate title
	titleHash := types.Hash(fromObj.Title)
	if !reflect.DeepEqual(titleHash, opData.TitleHash) {
		return pkgservice.ErrInvalidObject
	}

	// logID
	toLogID := toSyncInfo.GetLogID()
	updateLogID := fromObj.GetUpdateLogID()

	if !reflect.DeepEqual(toLogID, updateLogID) {
		return pkgservice.ErrInvalidObject
	}

	// get block-info
	origBlockInfo := toSyncInfo.GetBlockInfo()

	blockInfo := fromObj.GetBlockInfo()
	if blockInfo == nil {
		return pkgservice.ErrInvalidData
	}

	blockInfo.IsGood = origBlockInfo.IsGood
	blockInfo.IsAllGood = origBlockInfo.IsAllGood

	toSyncInfo.SetBlockInfo(blockInfo)

	// title
	toSyncInfo.Title = fromObj.Title

	return nil
}
