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

/**********
 * Process Oplog
 **********/

func (pm *BaseProtocolManager) processMasterLog(oplog *BaseOplog, processInfo ProcessInfo) (origLogs []*BaseOplog, err error) {

	info, ok := processInfo.(*ProcessPersonInfo)
	if !ok {
		return nil, ErrInvalidData
	}

	switch oplog.Op {
	case MasterOpTypeAddMaster:
		origLogs, err = pm.handleAddMasterLog(oplog, info)
	case MasterOpTypeMigrateMaster:
		origLogs, err = pm.handleMigrateMasterLog(oplog, info)
	case MasterOpTypeTransferMaster:
		origLogs, err = pm.handleTransferMasterLog(oplog, info)
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *BaseProtocolManager) processPendingMasterLog(oplog *BaseOplog, processInfo ProcessInfo) (types.Bool, []*BaseOplog, error) {

	info, ok := processInfo.(*ProcessPersonInfo)
	if !ok {
		return false, nil, ErrInvalidData
	}

	var isToSign types.Bool
	var origLogs []*BaseOplog
	var err error
	switch oplog.Op {
	case MasterOpTypeAddMaster:
		isToSign, origLogs, err = pm.handlePendingAddMasterLog(oplog, info)
	case MasterOpTypeMigrateMaster:
		isToSign, origLogs, err = pm.handlePendingMigrateMasterLog(oplog, info)
	case MasterOpTypeTransferMaster:
		isToSign, origLogs, err = pm.handlePendingTransferMasterLog(oplog, info)
	}
	return isToSign, origLogs, err
}

/**********
 * Postprocess Oplog
 **********/

func (pm *BaseProtocolManager) postprocessMasterOplogs(processInfo ProcessInfo, toBroadcastLogs []*BaseOplog, peer *PttPeer, isPending bool) error {

	pm.broadcastMasterOplogsCore(toBroadcastLogs)

	return nil
}

/**********
 * Set Newest Oplog
 **********/

func (pm *BaseProtocolManager) SetNewestMasterOplog(oplog *BaseOplog) error {
	var err error
	var isNewer types.Bool

	switch oplog.Op {
	case MasterOpTypeAddMaster:
		isNewer, err = pm.setNewestAddMasterLog(oplog)
	case MasterOpTypeMigrateMaster:
		isNewer, err = pm.setNewestMigrateMasterLog(oplog)
	case MasterOpTypeTransferMaster:
		isNewer, err = pm.setNewestTransferMasterLog(oplog)
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

func (pm *BaseProtocolManager) HandleFailedMasterOplog(oplog *BaseOplog) error {
	var err error

	switch oplog.Op {
	case MasterOpTypeAddMaster:
		err = pm.handleFailedAddMasterLog(oplog)
	case MasterOpTypeMigrateMaster:
		err = pm.handleFailedMigrateMasterLog(oplog)
	case MasterOpTypeTransferMaster:
		err = pm.handleFailedTransferMasterLog(oplog)
	}

	return err
}

func (pm *BaseProtocolManager) HandleFailedValidMasterOplog(oplog *BaseOplog, processInfo ProcessInfo) error {
	var err error

	switch oplog.Op {
	case MasterOpTypeAddMaster:
		err = pm.handleFailedValidAddMasterLog(oplog)
	case MasterOpTypeMigrateMaster:
		err = pm.handleFailedValidMigrateMasterLog(oplog)
	case MasterOpTypeTransferMaster:
		err = pm.handleFailedValidTransferMasterLog(oplog)
	}

	return err
}

/**********
 * Postprocess Failed Valid Oplog
 **********/

func (pm *BaseProtocolManager) postprocessFailedValidMasterOplogs(processInfo ProcessInfo, peer *PttPeer) error {

	return nil
}

/**********
 * Postsync Oplog
 **********/

func (pm *BaseProtocolManager) postsyncMasterOplogs(peer *PttPeer) error {

	log.Debug("postsyncMasterOplogs: start", "entity", pm.Entity().IDString())
	pm.SyncPendingMasterOplog(peer)
	pm.SyncOplog(peer, pm.memberMerkle, SyncMemberOplogMsg)

	return nil
}
