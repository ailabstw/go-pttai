// Copyright 2019 The go-pttai Authors
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

package service

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type ProcessOpKeyInfo struct {
	CreateOpKeyInfo map[types.PttID]*BaseOplog
	DeleteOpKeyInfo map[types.PttID]*BaseOplog
}

func NewProcessOpKeyInfo() *ProcessOpKeyInfo {
	return &ProcessOpKeyInfo{
		CreateOpKeyInfo: make(map[types.PttID]*BaseOplog),
		DeleteOpKeyInfo: make(map[types.PttID]*BaseOplog),
	}
}

/**********
 * Process Oplog
 **********/

func (pm *BaseProtocolManager) processOpKeyLog(oplog *BaseOplog, processInfo ProcessInfo) (origLogs []*BaseOplog, err error) {

	info, ok := processInfo.(*ProcessOpKeyInfo)
	if !ok {
		return nil, ErrInvalidData
	}

	switch oplog.Op {
	case OpKeyOpTypeCreateOpKey:
		origLogs, err = pm.handleCreateOpKeyLog(oplog, info)
	case OpKeyOpTypeRevokeOpKey:
		origLogs, err = pm.handleRevokeOpKeyLog(oplog, info)
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *BaseProtocolManager) processPendingOpKeyLog(oplog *BaseOplog, processInfo ProcessInfo) (types.Bool, []*BaseOplog, error) {

	info, ok := processInfo.(*ProcessOpKeyInfo)
	if !ok {
		return false, nil, ErrInvalidData
	}

	var isToSign types.Bool
	var origLogs []*BaseOplog
	var err error
	switch oplog.Op {
	case OpKeyOpTypeCreateOpKey:
		isToSign, origLogs, err = pm.handlePendingCreateOpKeyLog(oplog, info)
	case OpKeyOpTypeRevokeOpKey:
		isToSign, origLogs, err = pm.handlePendingRevokeOpKeyLog(oplog, info)
	}
	return isToSign, origLogs, err
}

/**********
 * Postprocess Oplog
 **********/

func (pm *BaseProtocolManager) postprocessOpKeyOplogs(processInfo ProcessInfo, toBroadcastLogs []*BaseOplog, peer *PttPeer, isPending bool) error {
	info, ok := processInfo.(*ProcessOpKeyInfo)
	if !ok {
		return ErrInvalidData
	}

	createOpKeyInfos, deleteOpKeyInfos := info.CreateOpKeyInfo, info.DeleteOpKeyInfo

	createOpKeyIDList := ProcessInfoToSyncIDList(createOpKeyInfos, OpKeyOpTypeCreateOpKey)

	log.Debug("postprocessOpKeyOplogs: to process create-op-key-ids", "createOpKeyIDList", createOpKeyIDList, "entity", pm.Entity().IDString())

	if isPending {
		toBroadcastLogs = ProcessInfoToBroadcastLogs(deleteOpKeyInfos, toBroadcastLogs)
	}

	pm.SyncCreateOpKey(createOpKeyIDList, peer)
	pm.broadcastOpKeyOplogsCore(toBroadcastLogs)

	return nil
}

/**********
 * Set Newest Oplog
 **********/

func (pm *BaseProtocolManager) SetNewestOpKeyOplog(oplog *BaseOplog) error {
	var err error
	var isNewer types.Bool

	switch oplog.Op {
	case OpKeyOpTypeCreateOpKey:
		isNewer, err = pm.setNewestCreateOpKeyLog(oplog)
	case OpKeyOpTypeRevokeOpKey:
		isNewer, err = pm.setNewestRevokeOpKeyLog(oplog)
	}

	if err != nil {
		return err
	}

	oplog.IsNewer = isNewer

	return nil
}

/**********
 * Handle Failed Oplog
 **********/

func (pm *BaseProtocolManager) HandleFailedOpKeyOplog(oplog *BaseOplog) error {
	var err error

	switch oplog.Op {
	case OpKeyOpTypeCreateOpKey:
		err = pm.handleFailedCreateOpKeyLog(oplog)
	case OpKeyOpTypeRevokeOpKey:
		err = pm.handleFailedRevokeOpKeyLog(oplog)
	}

	return err
}

/**********
 * Postsync Oplog
 **********/

func (pm *BaseProtocolManager) postsyncOpKeyOplogs(peer *PttPeer) error {
	return pm.SyncPendingOpKeyOplog(peer)
}
