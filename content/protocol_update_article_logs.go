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

func (pm *ProtocolManager) handleUpdateArticleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	opData := &BoardOpUpdateArticle{}

	log.Debug("handleUpdateArticleLogs: to HandleUpdateObjectLog", "entity", pm.Entity().IDString(), "IsSync", oplog.IsSync)

	return pm.HandleUpdateObjectLog(
		oplog,
		opData,

		obj,
		info,
		pm.boardOplogMerkle,

		pm.syncArticleInfoFromOplog,
		pm.SetBoardDB,
		nil,
		nil,
		pm.updateUpdateArticleInfo,
	)
}

func (pm *ProtocolManager) handlePendingUpdateArticleLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	opData := &BoardOpUpdateArticle{}

	log.Debug("handlePendingUpdateArticleLogs: to HandlePendingUpdateObjectLog", "entity", pm.Entity().IDString())

	return pm.HandlePendingUpdateObjectLog(
		oplog,
		opData,

		obj,
		info,
		pm.boardOplogMerkle,

		pm.syncArticleInfoFromOplog,
		pm.SetBoardDB,
		nil,
		nil,
		pm.updateUpdateArticleInfo,
	)
}

func (pm *ProtocolManager) setNewestUpdateArticleLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	return pm.SetNewestUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedUpdateArticleLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	return pm.HandleFailedUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedValidUpdateArticleLog(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) error {

	obj := NewEmptyArticle()
	pm.SetArticleDB(obj)

	return pm.HandleFailedValidUpdateObjectLog(oplog, obj, info, pm.updateUpdateArticleInfo)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) syncArticleInfoFromOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	opData, ok := theOpData.(*BoardOpUpdateArticle)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	syncInfo := NewEmptySyncArticleInfo()
	syncInfo.InitWithOplog(types.StatusInternalSync, oplog)

	blockInfo, err := pkgservice.NewBlockInfo(opData.BlockInfoID, opData.Hashs, opData.MediaIDs, oplog.CreatorID)
	if err != nil {
		return nil, err
	}
	pm.SetBlockInfoDB(blockInfo, oplog.ObjID)
	blockInfo.InitIsGood()
	syncInfo.SetBlockInfo(blockInfo)

	return syncInfo, nil
}

func (pm *ProtocolManager) updateUpdateArticleInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, origSyncInfo pkgservice.SyncInfo, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*BoardOpUpdateArticle)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.ArticleInfo[*oplog.ObjID] = oplog
	info.ArticleBlockInfo[*opData.BlockInfoID] = oplog

	return nil
}
