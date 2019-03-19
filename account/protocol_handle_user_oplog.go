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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProcessUserInfo struct {
	CreateUserNameInfo map[types.PttID]*pkgservice.BaseOplog
	UserNameInfo       map[types.PttID]*pkgservice.BaseOplog
	CreateUserImgInfo  map[types.PttID]*pkgservice.BaseOplog
	UserImgInfo        map[types.PttID]*pkgservice.BaseOplog
	CreateUserNodeInfo map[types.PttID]*pkgservice.BaseOplog
	UserNodeInfo       map[types.PttID]*pkgservice.BaseOplog
	CreateNameCardInfo map[types.PttID]*pkgservice.BaseOplog
	NameCardInfo       map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessUserInfo() *ProcessUserInfo {
	return &ProcessUserInfo{
		CreateUserNameInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		UserNameInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		CreateUserImgInfo:  make(map[types.PttID]*pkgservice.BaseOplog),
		UserImgInfo:        make(map[types.PttID]*pkgservice.BaseOplog),
		CreateUserNodeInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		UserNodeInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
		CreateNameCardInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		NameCardInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
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
		origLogs, err = pm.handleRemoveUserNodeLog(oplog, info)

	case UserOpTypeCreateNameCard:
		origLogs, err = pm.handleCreateNameCardLogs(oplog, info)
	case UserOpTypeUpdateNameCard:
		origLogs, err = pm.handleUpdateNameCardLogs(oplog, info)
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingUserLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (isToSign types.Bool, origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return false, nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case UserOpTypeCreateUserName:
		isToSign, origLogs, err = pm.handlePendingCreateUserNameLogs(oplog, info)
	case UserOpTypeUpdateUserName:
		isToSign, origLogs, err = pm.handlePendingUpdateUserNameLogs(oplog, info)

	case UserOpTypeCreateUserImg:
		isToSign, origLogs, err = pm.handlePendingCreateUserImgLogs(oplog, info)
	case UserOpTypeUpdateUserImg:
		isToSign, origLogs, err = pm.handlePendingUpdateUserImgLogs(oplog, info)

	case UserOpTypeAddUserNode:
		isToSign, origLogs, err = pm.handlePendingAddUserNodeLog(oplog, info)
	case UserOpTypeRemoveUserNode:
		isToSign, origLogs, err = pm.handlePendingRemoveUserNodeLog(oplog, info)

	case UserOpTypeCreateNameCard:
		isToSign, origLogs, err = pm.handlePendingCreateNameCardLogs(oplog, info)
	case UserOpTypeUpdateNameCard:
		isToSign, origLogs, err = pm.handlePendingUpdateNameCardLogs(oplog, info)
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

	// name card
	createNameCardIDs := pkgservice.ProcessInfoToSyncIDList(info.CreateNameCardInfo, UserOpTypeCreateNameCard)

	updateNameCardIDs := pkgservice.ProcessInfoToSyncIDList(info.NameCardInfo, UserOpTypeUpdateNameCard)

	pm.SyncNameCard(SyncCreateNameCardMsg, createNameCardIDs, peer)
	pm.SyncNameCard(SyncUpdateNameCardMsg, updateNameCardIDs, peer)

	// broadcast
	myID := pm.Ptt().GetMyEntity().GetID()
	if isPending || pm.IsMaster(myID, false) {
		pm.broadcastUserOplogsCore(toBroadcastLogs)
	}

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
		isNewer, err = pm.setNewestRemoveUserNodeLog(oplog)

	case UserOpTypeCreateNameCard:
		isNewer, err = pm.setNewestCreateNameCardLog(oplog)
	case UserOpTypeUpdateNameCard:
		isNewer, err = pm.setNewestUpdateNameCardLog(oplog)
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
		err = pm.handleFailedRemoveUserNodeLog(oplog)

	case UserOpTypeCreateNameCard:
		err = pm.handleFailedCreateNameCardLog(oplog)
	case UserOpTypeUpdateNameCard:
		err = pm.handleFailedUpdateNameCardLog(oplog)
	}

	return
}

/**********
 * Handle Failed Valid Oplog
 **********/

func (pm *ProtocolManager) HandleFailedValidUserOplog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (err error) {

	info, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case UserOpTypeCreateUserName:
		err = pm.handleFailedValidCreateUserNameLog(oplog, info)
	case UserOpTypeUpdateUserName:
		err = pm.handleFailedValidUpdateUserNameLog(oplog, info)

	case UserOpTypeCreateUserImg:
		err = pm.handleFailedValidCreateUserImgLog(oplog, info)
	case UserOpTypeUpdateUserImg:
		err = pm.handleFailedValidUpdateUserImgLog(oplog, info)

	case UserOpTypeAddUserNode:
		err = pm.handleFailedValidAddUserNodeLog(oplog, info)
	case UserOpTypeRemoveUserNode:
		err = pm.handleFailedValidRemoveUserNodeLog(oplog, info)

	case UserOpTypeCreateNameCard:
		err = pm.handleFailedValidCreateNameCardLog(oplog, info)
	case UserOpTypeUpdateNameCard:
		err = pm.handleFailedValidUpdateNameCardLog(oplog, info)
	}

	return
}

func (pm *ProtocolManager) postprocessFailedValidUserOplogs(processInfo pkgservice.ProcessInfo, peer *pkgservice.PttPeer) error {

	info, ok := processInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// user-name
	userNameIDs := pkgservice.ProcessInfoToForceSyncIDList(info.UserNameInfo)

	pm.ForceSyncUserName(userNameIDs, peer)

	// user-img
	userImgIDs := pkgservice.ProcessInfoToForceSyncIDList(info.UserImgInfo)

	pm.ForceSyncUserImg(userImgIDs, peer)

	// user-node
	userNodeIDs := pkgservice.ProcessInfoToForceSyncIDList(info.UserNodeInfo)

	pm.ForceSyncUserNode(userNodeIDs, peer)

	// name-card
	nameCardIDs := pkgservice.ProcessInfoToForceSyncIDList(info.NameCardInfo)

	pm.ForceSyncNameCard(nameCardIDs, peer)

	return nil
}

/**********
 * Postsync Oplog
 **********/

func (pm *ProtocolManager) postsyncUserOplogs(peer *pkgservice.PttPeer) (err error) {
	err = pm.SyncPendingUserOplog(peer)

	return
}
