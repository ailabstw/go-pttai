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

import "github.com/ailabstw/go-pttai/common/types"

/**********
 * Process Oplog
 **********/

func (pm *BaseProtocolManager) processMemberLog(oplog *BaseOplog, processInfo ProcessInfo) (origLogs []*BaseOplog, err error) {

	info, ok := processInfo.(*ProcessPersonInfo)
	if !ok {
		return nil, ErrInvalidData
	}

	switch oplog.Op {
	case MemberOpTypeAddMember:
		origLogs, err = pm.handleAddMemberLog(oplog, info)
	case MemberOpTypeDeleteMember:
		origLogs, err = pm.handleDeleteMemberLog(oplog, info)
	case MemberOpTypeMigrateMember:
		origLogs, err = pm.handleMigrateMemberLog(oplog, info)
	}
	return
}

/**********
 * Process Pending Oplog
 **********/

func (pm *BaseProtocolManager) processPendingMemberLog(oplog *BaseOplog, processInfo ProcessInfo) (types.Bool, []*BaseOplog, error) {

	info, ok := processInfo.(*ProcessPersonInfo)
	if !ok {
		return false, nil, ErrInvalidData
	}

	var isToSign types.Bool
	var origLogs []*BaseOplog
	var err error
	switch oplog.Op {
	case MemberOpTypeAddMember:
		isToSign, origLogs, err = pm.handlePendingAddMemberLog(oplog, info)
	case MemberOpTypeDeleteMember:
		isToSign, origLogs, err = pm.handlePendingDeleteMemberLog(oplog, info)
	case MemberOpTypeMigrateMember:
		isToSign, origLogs, err = pm.handlePendingMigrateMemberLog(oplog, info)
	}
	return isToSign, origLogs, err
}

/**********
 * Postprocess Oplog
 **********/

func (pm *BaseProtocolManager) postprocessMemberOplogs(processInfo ProcessInfo, toBroadcastLogs []*BaseOplog, peer *PttPeer, isPending bool) error {
	info, ok := processInfo.(*ProcessPersonInfo)
	if !ok {
		return ErrInvalidData
	}

	deleteInfos := info.DeleteInfo

	if isPending {
		for _, eachLog := range deleteInfos {
			toBroadcastLogs = pm.PostprocessPendingDeleteOplog(eachLog, toBroadcastLogs)
		}
	}

	pm.broadcastMemberOplogsCore(toBroadcastLogs)

	if !isPending {
		for _, eachLog := range deleteInfos {
			pm.postdeleteMember(eachLog.ObjID, eachLog, nil, nil)
		}
	}

	return nil
}

/**********
 * Set Newest Oplog
 **********/

func (pm *BaseProtocolManager) SetNewestMemberOplog(oplog *BaseOplog) error {
	var err error
	var isNewer types.Bool

	switch oplog.Op {
	case MemberOpTypeAddMember:
		isNewer, err = pm.setNewestAddMemberLog(oplog)
	case MemberOpTypeDeleteMember:
		isNewer, err = pm.setNewestDeleteMemberLog(oplog)
	case MemberOpTypeMigrateMember:
		isNewer, err = pm.setNewestMigrateMemberLog(oplog)
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

func (pm *BaseProtocolManager) HandleFailedMemberOplog(oplog *BaseOplog) error {
	var err error

	switch oplog.Op {
	case MemberOpTypeAddMember:
		err = pm.handleFailedAddMemberLog(oplog)
	case MemberOpTypeDeleteMember:
		err = pm.handleFailedDeleteMemberLog(oplog)
	case MemberOpTypeMigrateMember:
		err = pm.handleFailedMigrateMemberLog(oplog)
	}

	return err
}

/**********
 * Handle Failed Valid Oplog
 **********/

func (pm *BaseProtocolManager) HandleFailedValidMemberOplog(oplog *BaseOplog, processInfo ProcessInfo) error {
	var err error

	switch oplog.Op {
	case MemberOpTypeAddMember:
		err = pm.handleFailedValidAddMemberLog(oplog)
	case MemberOpTypeDeleteMember:
		err = pm.handleFailedValidDeleteMemberLog(oplog)
	case MemberOpTypeMigrateMember:
		err = pm.handleFailedValidMigrateMemberLog(oplog)
	}

	return err
}

/**********
 * Postprocess Failed Valid Oplog
 **********/

func (pm *BaseProtocolManager) postprocessFailedValidMemberOplogs(processInfo ProcessInfo, peer *PttPeer) error {

	return nil
}

/**********
 * Postsync Oplog
 **********/

func (pm *BaseProtocolManager) postsyncMemberOplogs(peer *PttPeer) error {
	pm.SyncPendingMemberOplog(peer)

	if pm.postsyncMemberOplog != nil {
		pm.postsyncMemberOplog(peer)
	}

	return nil
}
