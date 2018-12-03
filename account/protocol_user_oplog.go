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

func (pm *ProtocolManager) CreateUserOplog(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, data interface{}) (*UserOplog, error) {

	myID := pm.Entity().GetID()

	oplog, err := NewUserOplog(objID, ts, myID, op, data, myID, pm.dbUserLock)
	if err != nil {
		return nil, err
	}

	err = pm.SignOplog(oplog.BaseOplog)
	if err != nil {
		return nil, err
	}

	return oplog, nil
}

/**********
 * BroadcastUserOplog
 **********/

func (pm *ProtocolManager) BroadcastUserOplog(oplog *UserOplog) error {
	return pm.broadcastUserOplogCore(oplog.BaseOplog)
}

func (pm *ProtocolManager) broadcastUserOplogCore(oplog *pkgservice.BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddUserOplogMsg, AddPendingUserOplogMsg)
}

/**********
 * BroadcastUserOplogs
 **********/

func (pm *ProtocolManager) BroadcastUserOplogs(opKeyLogs []*UserOplog) error {
	oplogs := UserOplogsToOplogs(opKeyLogs)
	return pm.broadcastUserOplogsCore(oplogs)
}

func (pm *ProtocolManager) broadcastUserOplogsCore(oplogs []*pkgservice.BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddUserOplogsMsg, AddPendingUserOplogsMsg)
}

/**********
 * SetUserOplogIsSync
 **********/

func (pm *ProtocolManager) SetUserOplogIsSync(oplog *UserOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastUserOplogCore)
}

/**********
 * CleanUserOplog
 **********/

func (pm *ProtocolManager) CleanUserOplog() {
	oplog := &pkgservice.BaseOplog{}
	pm.SetUserDB(oplog)

	pm.CleanOplog(oplog, pm.userOplogMerkle)

}
