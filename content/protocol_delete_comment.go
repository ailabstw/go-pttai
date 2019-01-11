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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) DeleteComment(id *types.PttID) error {

	comment := NewEmptyComment()
	pm.SetCommentDB(comment)

	opData := &BoardOpDeleteComment{}

	return pm.DeleteObject(
		id,

		BoardOpTypeDeleteComment,
		comment,
		opData,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.NewBoardOplog,
		nil,
		pm.setPendingDeleteCommentSyncInfo,

		pm.broadcastBoardOplogCore,
		pm.postdeleteComment,
	)
}

func (pm *ProtocolManager) setPendingDeleteCommentSyncInfo(obj pkgservice.Object, status types.Status, oplog *pkgservice.BaseOplog) error {

	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithOplog(status, oplog)

	obj.SetSyncInfo(syncInfo)

	return nil
}

func (pm *ProtocolManager) postdeleteComment(id *types.PttID, oplog *pkgservice.BaseOplog, opData pkgservice.OpData, obj pkgservice.Object, blockInfo *pkgservice.BlockInfo) error {

	return nil
}
