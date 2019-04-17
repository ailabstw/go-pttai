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
 * BroadcastMasterOplog
 **********/

func (pm *BaseProtocolManager) BroadcastMasterOplog(oplog *MasterOplog) error {
	return pm.broadcastMasterOplogCore(oplog.BaseOplog)
}

func (pm *BaseProtocolManager) broadcastMasterOplogCore(oplog *BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddMasterOplogMsg, AddPendingMasterOplogMsg)
}

/**********
 * BroadcastMasterOplogs
 **********/

func (pm *BaseProtocolManager) BroadcastMasterOplogs(opKeyLogs []*MasterOplog) error {
	oplogs := MasterOplogsToOplogs(opKeyLogs)
	return pm.broadcastMasterOplogsCore(oplogs)
}

func (pm *BaseProtocolManager) broadcastMasterOplogsCore(oplogs []*BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddMasterOplogsMsg, AddPendingMasterOplogsMsg)
}

/**********
 * SetMasterOplogIsSync
 **********/

func (pm *BaseProtocolManager) SetMasterOplogIsSync(oplog *MasterOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastMasterOplogCore)
}

/**********
 * CleanMasterOplog
 **********/

func (pm *BaseProtocolManager) CleanMasterOplog() {
	oplog := &BaseOplog{}
	pm.SetMasterDB(oplog)

	pm.CleanOplog(oplog, nil)
}
