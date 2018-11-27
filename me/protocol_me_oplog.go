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

package me

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) CreateMeOplog(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, data interface{}) (*MeOplog, error) {

	myID := pm.Entity().GetID()

	oplog, err := NewMeOplog(objID, ts, myID, op, data, myID, pm.dbMeLock)
	if err != nil {
		return nil, err
	}

	err = pm.SignOplog(oplog.BaseOplog)
	if err != nil {
		return nil, err
	}

	if op > OffsetMeOpTypeEntity && oplog.MasterLogID == nil {
		oplog.SetMasterLogID(pm.GetNewestMasterLogID(), 1)
	}

	return oplog, nil
}

func (pm *ProtocolManager) GetPendingMeOplogs() ([]*MeOplog, []*MeOplog, error) {
	oplogs, failedLogs, err := pm.GetPendingOplogs(pm.SetMeDB)
	if err != nil {
		return nil, nil, err
	}

	opKeyLogs := OplogsToMeOplogs(oplogs)

	failedMeLogs := OplogsToMeOplogs(failedLogs)

	return opKeyLogs, failedMeLogs, nil
}

/**********
 * BroadcastMeOplog
 **********/

func (pm *ProtocolManager) BroadcastMeOplog(oplog *MeOplog) error {
	return pm.broadcastMeOplogCore(oplog.BaseOplog)
}

func (pm *ProtocolManager) broadcastMeOplogCore(oplog *pkgservice.BaseOplog) error {
	return pm.BroadcastOplog(oplog, AddMeOplogMsg, AddPendingMeOplogMsg)
}

/**********
 * BroadcastMeOplogs
 **********/

func (pm *ProtocolManager) BroadcastMeOplogs(opKeyLogs []*MeOplog) error {
	oplogs := MeOplogsToOplogs(opKeyLogs)
	return pm.broadcastMeOplogsCore(oplogs)
}

func (pm *ProtocolManager) broadcastMeOplogsCore(oplogs []*pkgservice.BaseOplog) error {
	return pm.BroadcastOplogs(oplogs, AddMeOplogsMsg, AddPendingMeOplogsMsg)
}

/**********
 * SetMeOplogIsSync
 **********/

func (pm *ProtocolManager) SetMeOplogIsSync(oplog *MeOplog, isBroadcast bool) (bool, error) {
	return pm.SetOplogIsSync(oplog.BaseOplog, isBroadcast, pm.broadcastMeOplogCore)
}

/**********
 * CleanMeOplog
 **********/

func (pm *ProtocolManager) CleanMeOplog() {
	oplog := &pkgservice.BaseOplog{}
	pm.SetMeDB(oplog)

	pm.CleanOplog(oplog, pm.meOplogMerkle)

}
