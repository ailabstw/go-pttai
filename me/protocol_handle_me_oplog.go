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
	DeleteMeInfo    map[types.PttID]*pkgservice.BaseOplog
	MetaInfo        map[types.PttID]*pkgservice.BaseOplog
	CreateBoardInfo map[types.PttID]*pkgservice.BaseOplog
	JoinBoardInfo   map[types.PttID]*pkgservice.BaseOplog
	BoardInfo       map[types.PttID]*pkgservice.BaseOplog
}

func NewProcessMeInfo() *ProcessMeInfo {
	return &ProcessMeInfo{
		DeleteMeInfo:    make(map[types.PttID]*pkgservice.BaseOplog),
		MetaInfo:        make(map[types.PttID]*pkgservice.BaseOplog),
		CreateBoardInfo: make(map[types.PttID]*pkgservice.BaseOplog),
		BoardInfo:       make(map[types.PttID]*pkgservice.BaseOplog),
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
	case MeOpTypeRevokeMe:
		origLogs, err = pm.handleRevokeMeLog(oplog, info)
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *ProtocolManager) processPendingMeLog(oplog *pkgservice.BaseOplog, processInfo pkgservice.ProcessInfo) (origLogs []*pkgservice.BaseOplog, err error) {
	info, ok := processInfo.(*ProcessMeInfo)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	switch oplog.Op {
	case MeOpTypeMigrateMe:
		origLogs, err = pm.handlePendingMigrateMeLog(oplog, info)
	case MeOpTypeRevokeMe:
		origLogs, err = pm.handlePendingRevokeMeLog(oplog, info)
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

	deleteMeInfo := info.DeleteMeInfo

	log.Debug("postprocessMeOplog", "isPending", isPending, "toBroadcastLogs", toBroadcastLogs, "deleteMeInfo", deleteMeInfo)

	if isPending {
		for _, eachLog := range deleteMeInfo {
			toBroadcastLogs = pm.PostprocessPendingDeleteOplog(eachLog, toBroadcastLogs)
		}
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
	case MeOpTypeRevokeMe:
		isNewer, err = pm.setNewestRevokeMeLog(oplog)
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
	case MeOpTypeRevokeMe:
		err = pm.handleFailedRevokeMeLog(oplog)
	}

	return
}

/**********
 * Postsync Oplog
 **********/

func (pm *ProtocolManager) postsyncMeOplogs(peer *pkgservice.PttPeer) (err error) {
	err = pm.SyncPendingMeOplog(peer)
	log.Debug("postsyncMeOplogs: after SyncPendingMeOplog", "e", err)

	return
}
