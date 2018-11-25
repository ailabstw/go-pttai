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
		origLogs, err = pm.handleDeleteArticleLogs(oplog, info)

	case BoardOpTypeCreateMedia:
		origLogs, err = pm.handleCreateMediaLogs(oplog, info)
	case BoardOpTypeDeleteMedia:
		origLogs, err = pm.handleDeleteMediaLogs(oplog, info)

	case BoardOpTypeCreateComment:
		origLogs, err = pm.handleCreateCommentLogs(oplog, info)
	case BoardOpTypeDeleteComment:
		origLogs, err = pm.handleDeleteCommentLogs(oplog, info)

	case BoardOpTypeCreateReply:
	case BoardOpTypeUpdateReply:
	case BoardOpTypeDeleteReply:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingBoardLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (isToSign types.Bool, origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessBoardInfo)
	if !ok {
		return false, nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case BoardOpTypeDeleteBoard:
	case BoardOpTypeMigrateBoard:

	case BoardOpTypeCreateTitle:
		isToSign, origLogs, err = pm.handlePendingCreateTitleLogs(oplog, info)
	case BoardOpTypeUpdateTitle:
		isToSign, origLogs, err = pm.handlePendingUpdateTitleLogs(oplog, info)

	case BoardOpTypeCreateArticle:
		isToSign, origLogs, err = pm.handlePendingCreateArticleLogs(oplog, info)
	case BoardOpTypeUpdateArticle:
		isToSign, origLogs, err = pm.handlePendingUpdateArticleLogs(oplog, info)
	case BoardOpTypeDeleteArticle:
		isToSign, origLogs, err = pm.handlePendingDeleteArticleLogs(oplog, info)

	case BoardOpTypeCreateMedia:
		isToSign, origLogs, err = pm.handlePendingCreateMediaLogs(oplog, info)
	case BoardOpTypeDeleteMedia:
		isToSign, origLogs, err = pm.handlePendingDeleteMediaLogs(oplog, info)

	case BoardOpTypeCreateComment:
		isToSign, origLogs, err = pm.handlePendingCreateCommentLogs(oplog, info)
	case BoardOpTypeDeleteComment:
		isToSign, origLogs, err = pm.handlePendingDeleteCommentLogs(oplog, info)

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
	createBlockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.ArticleBlockInfo, BoardOpTypeCreateArticle)
	pm.SyncArticle(SyncCreateArticleMsg, createArticleIDs, peer)
	pm.SyncBlock(SyncCreateArticleBlockMsg, createBlockIDs, peer)

	updateArticleIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateArticleInfo, BoardOpTypeUpdateArticle)
	updateBlockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.ArticleBlockInfo, BoardOpTypeUpdateArticle)
	pm.SyncArticle(SyncUpdateArticleMsg, updateArticleIDs, peer)
	pm.SyncBlock(SyncUpdateArticleBlockMsg, updateBlockIDs, peer)

	var deleteArticleLogs []*pkgservice.BaseOplog
	if isPending {
		deleteArticleLogs = pkgservice.ProcessInfoToLogs(info.ArticleInfo, BoardOpTypeDeleteArticle)
	}

	// comment
	createCommentIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateCommentInfo, BoardOpTypeCreateComment)
	createCommentBlockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.CommentBlockInfo, BoardOpTypeCreateComment)
	pm.SyncComment(SyncCreateCommentMsg, createCommentIDs, peer)
	pm.SyncBlock(SyncCreateCommentBlockMsg, createCommentBlockIDs, peer)

	var deleteCommentLogs []*pkgservice.BaseOplog
	if isPending {
		deleteCommentLogs = pkgservice.ProcessInfoToLogs(info.CommentInfo, BoardOpTypeDeleteComment)
	}

	// media
	createMediaIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateMediaInfo, BoardOpTypeCreateMedia)
	createMediaBlockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.MediaBlockInfo, BoardOpTypeCreateMedia)
	pm.SyncMedia(SyncCreateMediaMsg, createMediaIDs, peer)
	pm.SyncBlock(SyncCreateMediaBlockMsg, createMediaBlockIDs, peer)

	var deleteMediaLogs []*pkgservice.BaseOplog
	if isPending {
		deleteMediaLogs = pkgservice.ProcessInfoToLogs(info.MediaInfo, BoardOpTypeDeleteMedia)
	}

	// broadcast
	if isPending {
		toBroadcastLogAry := [][]*pkgservice.BaseOplog{
			toBroadcastLogs,
			deleteArticleLogs,
			deleteCommentLogs,
			deleteMediaLogs,
		}
		toBroadcastLogs, err = pkgservice.ConcatLog(toBroadcastLogAry)
		if err != nil {
			return
		}
	}

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
		isNewer, err = pm.setNewestDeleteArticleLog(oplog)

	case BoardOpTypeCreateMedia:
		isNewer, err = pm.SetNewestCreateMediaLog(oplog)
	case BoardOpTypeDeleteMedia:
		isNewer, err = pm.SetNewestDeleteMediaLog(oplog)

	case BoardOpTypeCreateComment:
		isNewer, err = pm.setNewestCreateCommentLog(oplog)
	case BoardOpTypeDeleteComment:
		isNewer, err = pm.setNewestDeleteCommentLog(oplog)

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
		err = pm.handleFailedDeleteArticleLog(oplog)

	case BoardOpTypeCreateMedia:
		err = pm.HandleFailedCreateMediaLog(oplog)
	case BoardOpTypeDeleteMedia:
		err = pm.HandleFailedDeleteMediaLog(oplog)

	case BoardOpTypeCreateComment:
		err = pm.handleFailedCreateCommentLog(oplog)
	case BoardOpTypeDeleteComment:
		err = pm.handleFailedDeleteCommentLog(oplog)

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
