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

package service

import (
	"github.com/ailabstw/go-pttai/common/types"
)

func (pm *BaseProtocolManager) getOpKeyOplogsFromKeys(keys [][]byte) ([]*OpKeyOplog, error) {
	logs, err := pm.GetOplogsFromKeys(pm.setOpKeyDB, keys)
	if err != nil {
		return nil, err
	}

	opKeyLogs := OplogsToOpKeyOplogs(logs)

	return opKeyLogs, nil
}

func (pm *BaseProtocolManager) IntegrateOpKeyOplog(log *OpKeyOplog, isLocked bool) (bool, error) {
	return pm.IntegrateOplog(log.Oplog, isLocked)
}

func (pm *BaseProtocolManager) GetPendingOpKeyOplogs() ([]*OpKeyOplog, []*OpKeyOplog, error) {
	logs, failedLogs, err := pm.GetPendingOplogs(pm.setOpKeyDB)
	if err != nil {
		return nil, nil, err
	}

	opKeyLogs := OplogsToOpKeyOplogs(logs)

	failedOpKeyLogs := OplogsToOpKeyOplogs(failedLogs)

	return opKeyLogs, failedOpKeyLogs, nil
}

func (pm *BaseProtocolManager) BroadcastOpKeyOplog(log *OpKeyOplog) error {
	return pm.BroadcastOplog(log.Oplog, AddOpKeyOplogMsg, AddPendingOpKeyOplogMsg)
}

func (pm *BaseProtocolManager) BroadcastOpKeyOplogs(opKeyLogs []*OpKeyOplog) error {
	logs := OpKeyOplogsToOplogs(opKeyLogs)
	return pm.BroadcastOplogs(logs, AddOpKeyOplogsMsg, AddPendingOpKeyOplogsMsg)
}

func (pm *BaseProtocolManager) SetOpKeyOplogIsSync(log *OpKeyOplog, isBroadcast bool) (bool, error) {
	isNewSign, err := pm.SetOplogIsSync(log.Oplog)
	if err != nil {
		return false, err
	}
	if isNewSign && isBroadcast {
		pm.BroadcastOpKeyOplog(log)
	}

	return isNewSign, nil
}

func (pm *BaseProtocolManager) RemoveNonSyncOpKeyOplog(logID *types.PttID, isRetainValid bool, isLocked bool) (*OpKeyOplog, error) {
	log, err := pm.RemoveNonSyncOplog(pm.setOpKeyDB, logID, isRetainValid, isLocked)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, nil
	}

	return &OpKeyOplog{Oplog: log}, nil
}
