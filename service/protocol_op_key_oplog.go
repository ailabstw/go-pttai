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
 * BroadcastOpKeyOplog
 **********/

func (pm *BaseProtocolManager) BroadcastOpKeyOplog(oplog *OpKeyOplog) error {
	return pm.broadcastOpKeyOplogCore(oplog.BaseOplog)
}

func (pm *BaseProtocolManager) broadcastOpKeyOplogCore(oplog *BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddOpKeyOplogMsg, AddPendingOpKeyOplogMsg)
}

/**********
 * BroadcastOpKeyOplogs
 **********/

func (pm *BaseProtocolManager) BroadcastOpKeyOplogs(opKeyLogs []*OpKeyOplog) error {
	oplogs := OpKeyOplogsToOplogs(opKeyLogs)
	return pm.broadcastOpKeyOplogsCore(oplogs)
}

func (pm *BaseProtocolManager) broadcastOpKeyOplogsCore(oplogs []*BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddOpKeyOplogsMsg, AddPendingOpKeyOplogsMsg)
}

/**********
 * SetOpKeyOplogIsSync
 **********/

func (pm *BaseProtocolManager) SetOpKeyOplogIsSync(oplog *OpKeyOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastOpKeyOplogCore)
}

/**********
 * CleanOpKeyOplog
 **********/

func (pm *BaseProtocolManager) CleanOpKeyOplog() {
	oplog := &BaseOplog{}
	pm.SetOpKeyDB(oplog)

	pm.CleanOplog(oplog, nil)
}
