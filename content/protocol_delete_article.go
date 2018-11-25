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
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) DeleteArticle(id *types.PttID) error {

	article := NewEmptyArticle()
	pm.SetArticleDB(article)

	opData := &BoardOpDeleteArticle{}

	return pm.DeleteObject(
		id,

		BoardOpTypeDeleteArticle,
		article,
		opData,

		pm.SetBoardDB,
		pm.NewBoardOplog,
		nil,
		pm.setPendingDeleteArticleSyncInfo,

		pm.broadcastBoardOplogCore,
		pm.postdeleteArticle,
	)
}

func (pm *ProtocolManager) setPendingDeleteArticleSyncInfo(obj pkgservice.Object, status types.Status, oplog *pkgservice.BaseOplog) error {

	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithOplog(status, oplog)

	obj.SetSyncInfo(syncInfo)

	return nil
}

func (pm *ProtocolManager) postdeleteArticle(id *types.PttID, oplog *pkgservice.BaseOplog, opData pkgservice.OpData, obj pkgservice.Object, blockInfo *pkgservice.BlockInfo) error {

	// comment
	comment := NewEmptyComment()
	pm.SetCommentDB(comment)

	iter, err := comment.GetCrossObjIterWithObj(id[:], nil, pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	var key []byte
	var eachID *types.PttID
	for iter.Next() {
		key = iter.Key()
		eachID, err = comment.KeyToID(key)
		comment.SetID(eachID)
		comment.GetAndDeleteAll(false)
	}

	return nil
}
