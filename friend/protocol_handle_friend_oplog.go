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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProcessFriendInfo struct {
}

func NewProcessFriendInfo() *ProcessFriendInfo {
	return &ProcessFriendInfo{}
}

/**********
 * Process Oplog
 **********/

func (pm *ProtocolManager) processFriendLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	_, ok := processInfo.(*ProcessFriendInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case FriendOpTypeDeleteFriend:
	case FriendOpTypeCreateArticle:
	case FriendOpTypeCreateMedia:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingFriendLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	_, ok := processInfo.(*ProcessFriendInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case FriendOpTypeDeleteFriend:
	case FriendOpTypeCreateArticle:
	case FriendOpTypeCreateMedia:
	}

	return
}

/**********
 * Postprocess Oplog
 **********/

func (pm *ProtocolManager) postprocessFriendOplogs(processInfo pkgservice.ProcessInfo, toBroadcastLogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isPending bool) (err error) {
	_, ok := processInfo.(*ProcessFriendInfo)
	if !ok {
		err = pkgservice.ErrInvalidData
	}

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
	case FriendOpTypeCreateArticle:
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
	case FriendOpTypeCreateArticle:
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
