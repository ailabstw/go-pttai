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
	CreateUserNameInfo map[types.PttID]*pkgservice.BaseOplog
	UserNameInfo       map[types.PttID]*pkgservice.BaseOplog
	CreateUserImgInfo  map[types.PttID]*pkgservice.BaseOplog
	UserImgInfo        map[types.PttID]*pkgservice.BaseOplog
	CreateUserNodeInfo map[types.PttID]*pkgservice.BaseOplog
	UserNodeInfo       map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessUserInfo() *ProcessUserInfo {
	return &ProcessUserInfo{
		CreateUserNameInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		UserNameInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		CreateUserImgInfo:  make(map[types.PttID]*pkgservice.BaseOplog),
		UserImgInfo:        make(map[types.PttID]*pkgservice.BaseOplog),
		CreateUserNodeInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		UserNodeInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
	}
}

/**********
 * Process Oplog
 **********/

func (pm *ProtocolManager) processUserLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case UserOpTypeCreateUserName:
		origLogs, err = pm.handleCreateUserNameLogs(oplog, info)
	case UserOpTypeUpdateUserName:
		origLogs, err = pm.handleUpdateUserNameLogs(oplog, info)

	case UserOpTypeCreateUserImg:
		origLogs, err = pm.handleCreateUserImgLogs(oplog, info)
	case UserOpTypeUpdateUserImg:
		origLogs, err = pm.handleUpdateUserImgLogs(oplog, info)

	case UserOpTypeAddUserNode:
		origLogs, err = pm.handleAddUserNodeLog(oplog, info)
	case UserOpTypeRemoveUserNode:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingUserLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	log.Debug("processPendingUserLog: to process", "op", oplog.Op)

	switch oplog.Op {
	case UserOpTypeCreateUserName:
		origLogs, err = pm.handlePendingCreateUserNameLogs(oplog, info)
	case UserOpTypeUpdateUserName:
		origLogs, err = pm.handlePendingUpdateUserNameLogs(oplog, info)

	case UserOpTypeCreateUserImg:
		origLogs, err = pm.handlePendingCreateUserImgLogs(oplog, info)
	case UserOpTypeUpdateUserImg:
		origLogs, err = pm.handlePendingUpdateUserImgLogs(oplog, info)

	case UserOpTypeAddUserNode:
		origLogs, err = pm.handlePendingAddUserNodeLog(oplog, info)
	case UserOpTypeRemoveUserNode:
	}
	return
}

/**********
 * Postprocess Oplog
 **********/

func (pm *ProtocolManager) postprocessUserOplogs(processInfo pkgservice.ProcessInfo, toBroadcastLogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isPending bool) (err error) {
	info, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		err = pkgservice.ErrInvalidData
	}

	// user node
	addUserNodeList := pkgservice.ProcessInfoToLogs(info.CreateUserNodeInfo, UserOpTypeAddUserNode)

	pm.SyncAddUserNode(addUserNodeList, peer)

	// user name
	createUserNameIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateUserNameInfo, UserOpTypeCreateUserName)

	updateUserNameIDs := pkgservice.ProcessInfoToSyncIDList(info.UserNameInfo, UserOpTypeUpdateUserName)

	pm.SyncUserName(SyncCreateUserNameMsg, createUserNameIDs, peer)
	pm.SyncUserName(SyncUpdateUserNameMsg, updateUserNameIDs, peer)

	// user img
	createUserImgIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateUserImgInfo, UserOpTypeCreateUserImg)

	updateUserImgIDs := pkgservice.ProcessInfoToSyncIDList(info.UserImgInfo, UserOpTypeUpdateUserImg)

	pm.SyncUserImg(SyncCreateUserImgMsg, createUserImgIDs, peer)
	pm.SyncUserImg(SyncUpdateUserImgMsg, updateUserImgIDs, peer)

	// broadcast
	pm.broadcastUserOplogsCore(toBroadcastLogs)

	return
}

/**********
 * Set Newest Oplog
 **********/

func (pm *ProtocolManager) SetNewestUserOplog(oplog *pkgservice.BaseOplog) (err error) {
	var isNewer types.Bool

	switch oplog.Op {
	case UserOpTypeCreateUserName:
		isNewer, err = pm.setNewestCreateUserNameLog(oplog)
	case UserOpTypeUpdateUserName:
		isNewer, err = pm.setNewestUpdateUserNameLog(oplog)

	case UserOpTypeCreateUserImg:
		isNewer, err = pm.setNewestCreateUserImgLog(oplog)
	case UserOpTypeUpdateUserImg:
		isNewer, err = pm.setNewestUpdateUserImgLog(oplog)

	case UserOpTypeAddUserNode:
		isNewer, err = pm.setNewestAddUserNodeLog(oplog)
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
	case UserOpTypeCreateUserName:
		err = pm.handleFailedCreateUserNameLog(oplog)
	case UserOpTypeUpdateUserName:
		err = pm.handleFailedUpdateUserNameLog(oplog)

	case UserOpTypeCreateUserImg:
		err = pm.handleFailedCreateUserImgLog(oplog)
	case UserOpTypeUpdateUserImg:
		err = pm.handleFailedUpdateUserImgLog(oplog)

	case UserOpTypeAddUserNode:
		err = pm.handleFailedAddUserNodeLog(oplog)
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
