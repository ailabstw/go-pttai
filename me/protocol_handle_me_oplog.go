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

package me

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProcessMeInfo struct {
	DeleteMeInfo map[types.PttID]*pkgservice.BaseOplog
	MetaInfo     map[types.PttID]*pkgservice.BaseOplog
	BoardInfo    map[types.PttID]*pkgservice.BaseOplog
	FriendInfo   map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessMeInfo() *ProcessMeInfo {
	return &ProcessMeInfo{
		DeleteMeInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		MetaInfo:     make(map[types.PttID]*pkgservice.BaseOplog),
		BoardInfo:    make(map[types.PttID]*pkgservice.BaseOplog),
		FriendInfo:   make(map[types.PttID]*pkgservice.BaseOplog),
	}
}

/**********
 * Process Oplog
 **********/

func (pm *ProtocolManager) processMeLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {

	info, ok := processInfo.(*ProcessMeInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case MeOpTypeMigrateMe:
		origLogs, err = pm.handleMigrateMeLog(oplog, info)
	case MeOpTypeDeleteMe:
		origLogs, err = pm.handleDeleteMeLog(oplog, info)

	case MeOpTypeCreateBoard:
		origLogs, err = pm.handleBoardLog(oplog, info)
	case MeOpTypeJoinBoard:
		origLogs, err = pm.handleBoardLog(oplog, info)

	case MeOpTypeCreateFriend:
		origLogs, err = pm.handleFriendLog(oplog, info)
	case MeOpTypeJoinFriend:
		origLogs, err = pm.handleFriendLog(oplog, info)

	case MeOpTypeSetNodeName:
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingMeLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (isToSign types.Bool, origLogs []*pkgservice.BaseOplog, err error) {

	info, ok := processInfo.(*ProcessMeInfo)
	if !ok {
		return false, nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case MeOpTypeMigrateMe:
		isToSign, origLogs, err = pm.handlePendingMigrateMeLog(oplog, info)
	case MeOpTypeDeleteMe:
		isToSign, origLogs, err = pm.handlePendingDeleteMeLog(oplog, info)

	case MeOpTypeSetNodeName:
	}
	return
}

/**********
 * Postprocess Oplog
 **********/

func (pm *ProtocolManager) postprocessMeOplogs(processInfo pkgservice.ProcessInfo, toBroadcastLogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isPending bool) (err error) {
	info, ok := processInfo.(*ProcessMeInfo)
	if !ok {
		err = pkgservice.ErrInvalidData
	}

	// board
	for _, oplog := range info.BoardInfo {
		pm.InternalSyncBoard(oplog, peer)
	}

	// friend
	for _, oplog := range info.FriendInfo {
		pm.InternalSyncFriend(oplog, peer)
	}

	// delete-me

	log.Debug("postprocessMeOplogs: to check delete-me", "isPending", isPending, "DeleteMeInfo", info.DeleteMeInfo)

	if isPending {
		toBroadcastLogs = pkgservice.ProcessInfoToBroadcastLogs(info.DeleteMeInfo, toBroadcastLogs)
	}

	pm.broadcastMeOplogsCore(toBroadcastLogs)

	return
}

/**********
 * Set Newest Oplog
 **********/

func (pm *ProtocolManager) SetNewestMeOplog(oplog *pkgservice.BaseOplog) (err error) {
	var isNewer types.Bool

	switch oplog.Op {
	case MeOpTypeMigrateMe:
		isNewer, err = pm.setNewestMigrateMeLog(oplog)
	case MeOpTypeDeleteMe:
		isNewer, err = pm.setNewestDeleteMeLog(oplog)

	case MeOpTypeCreateBoard:
		isNewer, err = pm.setNewestBoardLog(oplog)
	case MeOpTypeJoinBoard:
		isNewer, err = pm.setNewestBoardLog(oplog)

	case MeOpTypeCreateFriend:
		isNewer, err = pm.setNewestFriendLog(oplog)
	case MeOpTypeJoinFriend:
		isNewer, err = pm.setNewestFriendLog(oplog)

	case MeOpTypeSetNodeName:
	}

	oplog.IsNewer = isNewer

	return
}

/**********
 * Handle Failed Oplog
 **********/

func (pm *ProtocolManager) HandleFailedMeOplog(oplog *pkgservice.BaseOplog) (err error) {
	switch oplog.Op {
	case MeOpTypeMigrateMe:
		err = pm.handleFailedMigrateMeLog(oplog)
	case MeOpTypeDeleteMe:
		err = pm.handleFailedDeleteMeLog(oplog)

	case MeOpTypeSetNodeName:
	}

	return
}

/**********
 * Handle Failed Valid Oplog
 **********/

func (pm *ProtocolManager) HandleFailedValidMeOplog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (err error) {

	info, ok := processInfo.(*ProcessMeInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case MeOpTypeMigrateMe:
		err = pm.handleFailedValidMigrateMeLog(oplog, info)
	case MeOpTypeDeleteMe:
		err = pm.handleFailedValidDeleteMeLog(oplog, info)

	case MeOpTypeSetNodeName:
	}

	return
}

func (pm *ProtocolManager) postprocessFailedValidMeOplogs(processInfo pkgservice.ProcessInfo, peer *pkgservice.PttPeer) error {

	return nil
}

/**********
 * Postsync Oplog
 **********/

func (pm *ProtocolManager) postsyncMeOplogs(peer *pkgservice.PttPeer) (err error) {
	err = pm.SyncPendingMeOplog(peer)

	return
}
