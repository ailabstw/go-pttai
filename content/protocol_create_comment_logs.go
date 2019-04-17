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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleCreateCommentLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	opData := &BoardOpCreateComment{}

	log.Debug("handleCreateCommentLogs: to HandleCreateObjectLog")
	return pm.HandleCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateComment, pm.newCommentWithOplog, pm.postcreateComment, pm.updateCreateCommentInfo)
}

func (pm *ProtocolManager) handlePendingCreateCommentLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	opData := &BoardOpCreateComment{}

	return pm.HandlePendingCreateObjectLog(
		oplog, obj, opData, info,
		pm.existsInInfoCreateComment, pm.newCommentWithOplog, pm.postcreateComment, pm.updateCreateCommentInfo)
}

func (pm *ProtocolManager) setNewestCreateCommentLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedCreateCommentLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

func (pm *ProtocolManager) handleFailedValidCreateCommentLog(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) error {

	obj := NewEmptyComment()
	pm.SetCommentDB(obj)

	return pm.HandleFailedValidCreateObjectLog(oplog, obj, nil)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) newCommentWithOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) pkgservice.Object {

	opData, ok := theOpData.(*BoardOpCreateComment)
	if !ok {
		return nil
	}

	// comment
	obj := NewEmptyComment()
	pm.SetCommentDB(obj)
	pkgservice.NewObjectWithOplog(obj, oplog)

	obj.ArticleID = opData.ArticleID

	// block info
	blockInfo, err := pkgservice.NewBlockInfo(opData.BlockInfoID, opData.Hashs, opData.MediaIDs, oplog.CreatorID)
	if err != nil {
		return nil
	}
	pm.SetBlockInfoDB(blockInfo, obj.ID)
	blockInfo.InitIsGood()
	obj.SetBlockInfo(blockInfo)

	return obj
}

func (pm *ProtocolManager) existsInInfoCreateComment(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateCommentInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) updateCreateCommentInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	blockInfo := obj.GetBlockInfo()
	if blockInfo == nil {
		return pkgservice.ErrInvalidData
	}

	info.CreateCommentInfo[*oplog.ObjID] = oplog
	info.CommentBlockInfo[*blockInfo.ID] = oplog

	return nil
}
