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

type ProcessBoardInfo struct {
	CreateTitleInfo map[types.PttID]*pkgservice.BaseOplog
	TitleInfo       map[types.PttID]*pkgservice.BaseOplog

	CreateArticleInfo map[types.PttID]*pkgservice.BaseOplog
	ArticleInfo       map[types.PttID]*pkgservice.BaseOplog
	ArticleBlockInfo  map[types.PttID]*pkgservice.BaseOplog

	CreateCommentInfo map[types.PttID]*pkgservice.BaseOplog
	CommentInfo       map[types.PttID]*pkgservice.BaseOplog
	CommentBlockInfo  map[types.PttID]*pkgservice.BaseOplog

	CreateReplyInfo map[types.PttID]*pkgservice.BaseOplog
	ReplyInfo       map[types.PttID]*pkgservice.BaseOplog
	ReplyBlockInfo  map[types.PttID]*pkgservice.BaseOplog

	CreateMediaInfo map[types.PttID]*pkgservice.BaseOplog
	MediaInfo       map[types.PttID]*pkgservice.BaseOplog
	MediaBlockInfo  map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessBoardInfo() *ProcessBoardInfo {
	return &ProcessBoardInfo{
		CreateTitleInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		TitleInfo:       make(map[types.PttID]*pkgservice.BaseOplog),

		CreateArticleInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		ArticleInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		ArticleBlockInfo:  make(map[types.PttID]*pkgservice.BaseOplog),

		CreateCommentInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		CommentInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		CommentBlockInfo:  make(map[types.PttID]*pkgservice.BaseOplog),

		CreateReplyInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		ReplyInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		ReplyBlockInfo:  make(map[types.PttID]*pkgservice.BaseOplog),

		CreateMediaInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		MediaInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		MediaBlockInfo:  make(map[types.PttID]*pkgservice.BaseOplog),
	}
}

/**********
 * Process Oplog
 **********/

func (pm *ProtocolManager) processBoardLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessBoardInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case BoardOpTypeDeleteBoard:
	case BoardOpTypeMigrateBoard:

	case BoardOpTypeCreateTitle:
		origLogs, err = pm.handleCreateTitleLogs(oplog, info)
	case BoardOpTypeUpdateTitle:
		origLogs, err = pm.handleUpdateTitleLogs(oplog, info)

	case BoardOpTypeCreateArticle:
		origLogs, err = pm.handleCreateArticleLogs(oplog, info)
	case BoardOpTypeUpdateArticle:
		origLogs, err = pm.handleUpdateArticleLogs(oplog, info)
	case BoardOpTypeDeleteArticle:

	case BoardOpTypeCreateMedia:
	case BoardOpTypeDeleteMedia:

	case BoardOpTypeCreateComment:
	case BoardOpTypeDeleteComment:

	case BoardOpTypeCreateReply:
	case BoardOpTypeUpdateReply:
	case BoardOpTypeDeleteReply:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingBoardLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessBoardInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case BoardOpTypeDeleteBoard:
	case BoardOpTypeMigrateBoard:

	case BoardOpTypeCreateTitle:
		origLogs, err = pm.handlePendingCreateTitleLogs(oplog, info)
	case BoardOpTypeUpdateTitle:
		origLogs, err = pm.handlePendingUpdateTitleLogs(oplog, info)

	case BoardOpTypeCreateArticle:
		origLogs, err = pm.handlePendingCreateArticleLogs(oplog, info)
	case BoardOpTypeUpdateArticle:
		origLogs, err = pm.handlePendingUpdateArticleLogs(oplog, info)
	case BoardOpTypeDeleteArticle:

	case BoardOpTypeCreateMedia:
	case BoardOpTypeDeleteMedia:

	case BoardOpTypeCreateComment:
	case BoardOpTypeDeleteComment:

	case BoardOpTypeCreateReply:
	case BoardOpTypeUpdateReply:
	case BoardOpTypeDeleteReply:
	}

	return
}

/**********
 * Postprocess Oplog
 **********/

func (pm *ProtocolManager) postprocessBoardOplogs(processInfo pkgservice.ProcessInfo, toBroadcastLogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isPending bool) (err error) {
	info, ok := processInfo.(*ProcessBoardInfo)
	if !ok {
		err = pkgservice.ErrInvalidData
	}

	// title
	createTitleIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateTitleInfo, BoardOpTypeCreateTitle)

	updateTitleIDs := pkgservice.ProcessInfoToSyncIDList(info.TitleInfo, BoardOpTypeUpdateTitle)

	pm.SyncTitle(SyncCreateTitleMsg, createTitleIDs, peer)
	pm.SyncTitle(SyncUpdateTitleMsg, updateTitleIDs, peer)

	// article
	createArticleIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateArticleInfo, BoardOpTypeCreateArticle)
	updateArticleIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateArticleInfo, BoardOpTypeUpdateArticle)

	createblockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.ArticleBlockInfo, BoardOpTypeCreateArticle)
	updateBlockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.ArticleBlockInfo, BoardOpTypeUpdateArticle)

	pm.SyncArticle(SyncCreateArticleMsg, createArticleIDs, peer)
	pm.SyncArticle(SyncUpdateArticleMsg, updateArticleIDs, peer)

	pm.SyncBlock(SyncCreateArticleBlockMsg, createblockIDs, peer)
	pm.SyncBlock(SyncUpdateArticleBlockMsg, updateBlockIDs, peer)

	// broadcast
	pm.broadcastBoardOplogsCore(toBroadcastLogs)

	return
}

/**********
 * Set Newest Oplog
 **********/

func (pm *ProtocolManager) SetNewestBoardOplog(oplog *pkgservice.BaseOplog) (err error) {
	var isNewer types.Bool

	switch oplog.Op {
	case BoardOpTypeDeleteBoard:
	case BoardOpTypeMigrateBoard:

	case BoardOpTypeCreateTitle:
		isNewer, err = pm.setNewestCreateTitleLog(oplog)
	case BoardOpTypeUpdateTitle:
		isNewer, err = pm.setNewestUpdateTitleLog(oplog)

	case BoardOpTypeCreateArticle:
		isNewer, err = pm.setNewestCreateArticleLog(oplog)
	case BoardOpTypeUpdateArticle:
		isNewer, err = pm.setNewestUpdateArticleLog(oplog)
	case BoardOpTypeDeleteArticle:

	case BoardOpTypeCreateMedia:
	case BoardOpTypeDeleteMedia:

	case BoardOpTypeCreateComment:
	case BoardOpTypeDeleteComment:

	case BoardOpTypeCreateReply:
	case BoardOpTypeUpdateReply:
	case BoardOpTypeDeleteReply:
	}

	oplog.IsNewer = isNewer

	return
}

/**********
 * Handle Failed Oplog
 **********/

func (pm *ProtocolManager) HandleFailedBoardOplog(oplog *pkgservice.BaseOplog) (err error) {

	switch oplog.Op {
	case BoardOpTypeDeleteBoard:
	case BoardOpTypeMigrateBoard:

	case BoardOpTypeCreateTitle:
		err = pm.handleFailedCreateTitleLog(oplog)
	case BoardOpTypeUpdateTitle:
		err = pm.handleFailedUpdateTitleLog(oplog)

	case BoardOpTypeCreateArticle:
		err = pm.handleFailedCreateArticleLog(oplog)
	case BoardOpTypeUpdateArticle:
		err = pm.handleFailedUpdateArticleLog(oplog)
	case BoardOpTypeDeleteArticle:

	case BoardOpTypeCreateMedia:
	case BoardOpTypeDeleteMedia:

	case BoardOpTypeCreateComment:
	case BoardOpTypeDeleteComment:

	case BoardOpTypeCreateReply:
	case BoardOpTypeUpdateReply:
	case BoardOpTypeDeleteReply:
	}

	return
}

/**********
 * Postsync Oplog
 **********/

func (pm *ProtocolManager) postsyncBoardOplogs(peer *pkgservice.PttPeer) (err error) {
	err = pm.SyncPendingBoardOplog(peer)

	return
}
