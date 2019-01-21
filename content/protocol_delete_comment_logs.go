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

func (pm *ProtocolManager) handleDeleteCommentLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	opData := &BoardOpDeleteComment{}

	return pm.HandleDeleteObjectLog(
		oplog,
		info,
		obj,
		opData,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.removeMediaInfoByBlockInfo,
		pm.postdeleteComment,
		pm.updateCommentDeleteInfo,
	)
}

func (pm *ProtocolManager) handlePendingDeleteCommentLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	opData := &BoardOpDeleteComment{}

	return pm.HandlePendingDeleteObjectLog(
		oplog,
		info,
		obj,
		opData,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.removeMediaInfoByBlockInfo,
		pm.setPendingDeleteCommentSyncInfo,
		pm.updateCommentDeleteInfo,
	)
}

func (pm *ProtocolManager) setNewestDeleteCommentLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	return pm.SetNewestDeleteObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedDeleteCommentLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	return pm.HandleFailedDeleteObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedValidDeleteCommentLog(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) error {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	return pm.HandleFailedValidDeleteObjectLog(oplog, obj, info, pm.updateCommentDeleteInfo)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) updateCommentDeleteInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.CommentInfo[*oplog.ObjID] = oplog

	return nil
}
