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

func (pm *ProtocolManager) handleDeleteArticleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {

	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	opData := &BoardOpDeleteArticle{}

	return pm.HandleDeleteObjectLog(
		oplog,
		info,
		obj,
		opData,

		pm.SetBoardDB,
		pm.removeMediaInfoByBlockInfo,
		pm.postdeleteArticle,
		pm.updateArticleDeleteInfo,
	)
}

func (pm *ProtocolManager) handlePendingDeleteArticleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	opData := &BoardOpDeleteArticle{}

	return pm.HandlePendingDeleteObjectLog(
		oplog,
		info, obj,
		opData,

		pm.SetBoardDB,
		pm.removeMediaInfoByBlockInfo,
		pm.setPendingDeleteArticleSyncInfo,
		pm.updateArticleDeleteInfo,
	)
}

func (pm *ProtocolManager) setNewestDeleteArticleLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {

	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	return pm.SetNewestDeleteObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedDeleteArticleLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	return pm.HandleFailedDeleteObjectLog(oplog, obj)
}

/**********
 * Remove Media Info
 **********/

func (pm *ProtocolManager) removeMediaInfoByBlockInfo(blockInfo *pkgservice.BlockInfo, theInfo pkgservice.ProcessInfo, oplog *pkgservice.BaseOplog) {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return
	}

	if blockInfo == nil {
		return
	}
	mediaIDs := blockInfo.MediaIDs
	if mediaIDs == nil {
		return
	}

	pm.RemoveMediaInfosByOplog(info.MediaInfo, mediaIDs, oplog, BoardOpTypeDeleteMedia)

}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) updateArticleDeleteInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.ArticleInfo[*oplog.ObjID] = oplog

	return nil
}
