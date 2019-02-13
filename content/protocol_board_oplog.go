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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) CreateBoardOplog(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, data interface{}) (*BoardOplog, error) {

	myID := pm.Entity().GetID()

	oplog, err := NewBoardOplog(objID, ts, myID, op, data, myID, pm.dbBoardLock)
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
 * BroadcastBoardOplog
 **********/

func (pm *ProtocolManager) BroadcastBoardOplog(oplog *BoardOplog) error {
	return pm.broadcastBoardOplogCore(oplog.BaseOplog)
}

func (pm *ProtocolManager) broadcastBoardOplogCore(oplog *pkgservice.BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddBoardOplogMsg, AddPendingBoardOplogMsg)
}

/**********
 * BroadcastBoardOplogs
 **********/

func (pm *ProtocolManager) BroadcastBoardOplogs(opKeyLogs []*BoardOplog) error {
	oplogs := BoardOplogsToOplogs(opKeyLogs)
	return pm.broadcastBoardOplogsCore(oplogs)
}

func (pm *ProtocolManager) broadcastBoardOplogsCore(oplogs []*pkgservice.BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddBoardOplogsMsg, AddPendingBoardOplogsMsg)
}

/**********
 * SetBoardOplogIsSync
 **********/

func (pm *ProtocolManager) SetBoardOplogIsSync(oplog *BoardOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastBoardOplogCore)
}
