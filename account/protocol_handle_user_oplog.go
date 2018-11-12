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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProcessUserInfo struct {
	UserNameInfo map[types.PttID]*pkgservice.BaseOplog
	UserImgInfo  map[types.PttID]*pkgservice.BaseOplog
	UserNodeInfo map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessUserInfo() *ProcessUserInfo {
	return &ProcessUserInfo{
		UserNameInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		UserImgInfo:  make(map[types.PttID]*pkgservice.BaseOplog),
		UserNodeInfo: make(map[types.PttID]*pkgservice.BaseOplog),
	}
}

/**********
 * Process Oplog
 **********/

func (pm *ProtocolManager) processUserLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	_, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case UserOpTypeSetUserName:
	case UserOpTypeSetUserImg:
	case UserOpTypeAddUserNode:
	case UserOpTypeRemoveUserNode:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingUserLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	_, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case UserOpTypeSetUserName:
	case UserOpTypeSetUserImg:
	case UserOpTypeAddUserNode:
	case UserOpTypeRemoveUserNode:
	}
	return
}

/**********
 * Postprocess Oplog
 **********/

func (pm *ProtocolManager) postprocessUserOplogs(processInfo pkgservice.ProcessInfo, toBroadcastLogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isPending bool) (err error) {
	_, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		err = pkgservice.ErrInvalidData
	}

	pm.broadcastUserOplogsCore(toBroadcastLogs)

	return
}

/**********
 * Set Newest Oplog
 **********/

func (pm *ProtocolManager) SetNewestUserOplog(oplog *pkgservice.BaseOplog) (err error) {
	var isNewer types.Bool

	switch oplog.Op {
	case UserOpTypeSetUserName:
	case UserOpTypeSetUserImg:
	case UserOpTypeAddUserNode:
	case UserOpTypeRemoveUserNode:
	}

	oplog.IsNewer = isNewer

	return
}

/**********
 * Handle Failed Oplog
 **********/

func (pm *ProtocolManager) HandleFailedUserOplog(oplog *pkgservice.BaseOplog) (err error) {
	switch oplog.Op {
	case UserOpTypeSetUserName:
	case UserOpTypeSetUserImg:
	case UserOpTypeAddUserNode:
	case UserOpTypeRemoveUserNode:
	}

	return
}

/**********
 * Postsync Oplog
 **********/

func (pm *ProtocolManager) postsyncUserOplogs(peer *pkgservice.PttPeer) (err error) {
	err = pm.SyncPendingUserOplog(peer)
	log.Debug("postsyncUserOplogs: after SyncPendingUserOplog", "e", err)

	return
}
