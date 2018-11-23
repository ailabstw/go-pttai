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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type UpdateArticle struct {
	Article  [][]byte       `json:"a"`
	MediaIDs []*types.PttID `json:"m"`
}

func (pm *ProtocolManager) UpdateArticle(articleID *types.PttID, articleBytes [][]byte, mediaIDs []*types.PttID) (*Article, error) {

	data := &UpdateArticle{Article: articleBytes, MediaIDs: mediaIDs}

	origObj := NewEmptyArticle()
	pm.SetArticleDB(origObj)

	opData := &BoardOpUpdateTitle{}

	err := pm.UpdateObject(
		articleID,
		data,
		BoardOpTypeUpdateTitle,
		origObj,
		opData,

		pm.SetBoardDB,
		pm.NewBoardOplog,
		pm.inupdateArticle,
		nil,
		pm.broadcastBoardOplogCore,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return origObj, nil
}

func (pm *ProtocolManager) inupdateArticle(obj pkgservice.Object, theData pkgservice.UpdateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	data, ok := theData.(*UpdateArticle)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*BoardOpUpdateArticle)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// block-info
	blockInfoID, blockHashs, err := pm.SplitContentBlocks(nil, oplog.ObjID, data.Article, NFirstLineInBlock)
	if err != nil {
		log.Error("inupdateArticle: Unable to SplitContentBlocks", "e", err)
		return nil, err
	}

	blockInfo, err := pkgservice.NewBlockInfo(blockInfoID, blockHashs, data.MediaIDs, oplog.CreatorID)
	if err != nil {
		return nil, err
	}
	blockInfo.SetIsAllGood()

	// op-data
	opData.BlockInfoID = blockInfoID
	opData.NBlock = blockInfo.NBlock
	opData.Hashs = blockHashs
	opData.MediaIDs = data.MediaIDs

	// sync-info
	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithOplog(oplog.ToStatus(), oplog)
	syncInfo.SetBlockInfo(blockInfo)

	return syncInfo, nil
}
