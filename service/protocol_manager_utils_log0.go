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

func (pm *BaseProtocolManager) GetOplog0() *BaseOplog {
	return pm.oplog0
}

func (pm *BaseProtocolManager) SetOplog0(oplog *BaseOplog) {
	pm.oplog0 = oplog
}

func (pm *BaseProtocolManager) SetLog0DB(oplog *BaseOplog) {
	pm.setLog0DB(oplog)
}

func (pm *BaseProtocolManager) Log0Merkle() *Merkle {
	return pm.log0Merkle
}

func (pm *BaseProtocolManager) HandleLog0s(logs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error {
	return pm.handleLog0s(logs, peer, isUpdateSyncTime)
}

func (pm *BaseProtocolManager) CleanLog0(isRetainLog bool) {

	var entityLog *BaseOplog
	// get entity log
	if isRetainLog {
		entityLog, _ = pm.GetEntityLog()
	}

	// clean-log
	oplog := &BaseOplog{}
	pm.setLog0DB(oplog)
	pm.CleanOplog(oplog, pm.log0Merkle)

	// entity log
	if isRetainLog && entityLog != nil {
		entityLog.Save(false, nil)
	}
}
