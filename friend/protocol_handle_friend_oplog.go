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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProcessFriendInfo struct {
	CreateMessageInfo map[types.PttID]*pkgservice.BaseOplog

	CreateMediaInfo map[types.PttID]*pkgservice.BaseOplog

	BlockInfo map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessFriendInfo() *ProcessFriendInfo {
	return &ProcessFriendInfo{
		CreateMessageInfo: make(map[types.PttID]*pkgservice.BaseOplog),

		CreateMediaInfo: make(map[types.PttID]*pkgservice.BaseOplog),

		BlockInfo: make(map[types.PttID]*pkgservice.BaseOplog),
	}
}

/**********
 * Process Oplog
 **********/

func (pm *ProtocolManager) processFriendLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessFriendInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case FriendOpTypeDeleteFriend:
	case FriendOpTypeCreateMessage:
		origLogs, err = pm.handleCreateMessageLogs(oplog, info)

	case FriendOpTypeCreateMedia:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingFriendLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (isToSign types.Bool, origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessFriendInfo)
	if !ok {
		return false, nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case FriendOpTypeDeleteFriend:

	case FriendOpTypeCreateMessage:
		isToSign, origLogs, err = pm.handlePendingCreateMessageLogs(oplog, info)

	case FriendOpTypeCreateMedia:
	}

	return
}

/**********
 * Postprocess Oplog
 **********/

func (pm *ProtocolManager) postprocessFriendOplogs(processInfo pkgservice.ProcessInfo, toBroadcastLogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isPending bool) (err error) {
	info, ok := processInfo.(*ProcessFriendInfo)
	if !ok {
		err = pkgservice.ErrInvalidData
	}

	// message
	createMessageIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateMessageInfo, FriendOpTypeCreateMessage)

	log.Debug("postprocessFriendOplogs: to syncMessage", "createMessageIDs", createMessageIDs)

	pm.SyncMessage(SyncCreateMessageMsg, createMessageIDs, peer)

	// blocks
	blockIDs := pkgservice.ProcessInfoToSyncBlockIDList(info.BlockInfo, FriendOpTypeCreateMessage)

	log.Debug("postprocessFriendOplogs: to syncBlock", "blockIDs", blockIDs)

	pm.SyncBlock(SyncCreateMessageBlockMsg, blockIDs, peer)

	pm.broadcastFriendOplogsCore(toBroadcastLogs)

	return
}

/**********
 * Set Newest Oplog
 **********/

func (pm *ProtocolManager) SetNewestFriendOplog(oplog *pkgservice.BaseOplog) (err error) {
	var isNewer types.Bool

	switch oplog.Op {
	case FriendOpTypeDeleteFriend:
	case FriendOpTypeCreateMessage:
		isNewer, err = pm.setNewestCreateMessageLog(oplog)
	case FriendOpTypeCreateMedia:
	}

	oplog.IsNewer = isNewer

	return
}

/**********
 * Handle Failed Oplog
 **********/

func (pm *ProtocolManager) HandleFailedFriendOplog(oplog *pkgservice.BaseOplog) (err error) {

	switch oplog.Op {
	case FriendOpTypeDeleteFriend:
	case FriendOpTypeCreateMessage:
		err = pm.handleFailedCreateMessageLog(oplog)
	case FriendOpTypeCreateMedia:
	}

	return
}

/**********
 * Postsync Oplog
 **********/

func (pm *ProtocolManager) postsyncFriendOplogs(peer *pkgservice.PttPeer) (err error) {
	err = pm.SyncPendingFriendOplog(peer)

	return
}
