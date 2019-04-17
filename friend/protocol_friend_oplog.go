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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) CreateFriendOplog(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, data interface{}) (*FriendOplog, error) {

	myID := pm.Entity().GetID()

	oplog, err := NewFriendOplog(objID, ts, myID, op, data, myID, pm.dbFriendLock)
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
 * BroadcastFriendOplog
 **********/

func (pm *ProtocolManager) BroadcastFriendOplog(oplog *FriendOplog) error {
	return pm.broadcastFriendOplogCore(oplog.BaseOplog)
}

func (pm *ProtocolManager) broadcastFriendOplogCore(oplog *pkgservice.BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddFriendOplogMsg, AddPendingFriendOplogMsg)
}

/**********
 * BroadcastFriendOplogs
 **********/

func (pm *ProtocolManager) BroadcastFriendOplogs(opKeyLogs []*FriendOplog) error {
	oplogs := FriendOplogsToOplogs(opKeyLogs)
	return pm.broadcastFriendOplogsCore(oplogs)
}

func (pm *ProtocolManager) broadcastFriendOplogsCore(oplogs []*pkgservice.BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddFriendOplogsMsg, AddPendingFriendOplogsMsg)
}

/**********
 * SetFriendOplogIsSync
 **********/

func (pm *ProtocolManager) SetFriendOplogIsSync(oplog *FriendOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastFriendOplogCore)
}
