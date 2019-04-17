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

/**********
 * BroadcastMemberOplog
 **********/

func (pm *BaseProtocolManager) BroadcastMemberOplog(oplog *MemberOplog) error {
	return pm.broadcastMemberOplogCore(oplog.BaseOplog)
}

func (pm *BaseProtocolManager) broadcastMemberOplogCore(oplog *BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddMemberOplogMsg, AddPendingMemberOplogMsg)
}

/**********
 * BroadcastMemberOplogs
 **********/

func (pm *BaseProtocolManager) BroadcastMemberOplogs(opKeyLogs []*MemberOplog) error {
	oplogs := MemberOplogsToOplogs(opKeyLogs)
	return pm.broadcastMemberOplogsCore(oplogs)
}

func (pm *BaseProtocolManager) broadcastMemberOplogsCore(oplogs []*BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddMemberOplogsMsg, AddPendingMemberOplogsMsg)
}

/**********
 * SetMemberOplogIsSync
 **********/

func (pm *BaseProtocolManager) SetMemberOplogIsSync(oplog *MemberOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastMemberOplogCore)
}

/**********
 * CleanMemberOplog
 **********/

func (pm *BaseProtocolManager) CleanMemberOplog(isRetainLog bool) {
	oplog := &BaseOplog{}
	pm.SetMemberDB(oplog)

	pm.CleanOplog(oplog, pm.MemberMerkle())

	// retain my-log
	if isRetainLog {
		myLog := pm.myMemberLog
		myLog.Save(false, nil)
	}
}
