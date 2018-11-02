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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) GetPendingOpKeyOplogs() ([]*OpKeyOplog, []*OpKeyOplog, error) {
	oplogs, failedLogs, err := pm.GetPendingOplogs(pm.SetOpKeyDB)
	if err != nil {
		return nil, nil, err
	}

	opKeyLogs := OplogsToOpKeyOplogs(oplogs)

	failedOpKeyLogs := OplogsToOpKeyOplogs(failedLogs)

	return opKeyLogs, failedOpKeyLogs, nil
}

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

func (pm *BaseProtocolManager) RemoveNonSyncOpKeyOplog(logID *types.PttID, isRetainValid bool, isLocked bool) (*OpKeyOplog, error) {
	oplog, err := pm.RemoveNonSyncOplog(pm.SetOpKeyDB, logID, isRetainValid, isLocked)
	if err != nil {
		return nil, err
	}
	return OplogToOpKeyOplog(oplog), nil
}

/**********
 * Handle Oplogs
 **********/

func (pm *BaseProtocolManager) CreateOpKeyPostprocess(theOpKey Object, oplog *BaseOplog) error {
	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	pm.RegisterOpKeyInfo(opKey, false)

	return nil
}

func (pm *BaseProtocolManager) FailedCreateOpKeyPostprocess(theOpKey Object, oplog *BaseOplog) error {
	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	log.Debug("FailedCreateOpKeyPostprocess: to Remove OpKeyInfoFromHash")

	return pm.RemoveOpKeyInfoFromHash(opKey.Hash, false, true, true)
}

func (pm *BaseProtocolManager) CreateOpKeyExistsInInfo(oplog *BaseOplog, theInfo ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return false, ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.DeleteOpKeyInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *BaseProtocolManager) DeleteOpKeyPostprocess(id *types.PttID, oplog *BaseOplog, origObj Object, opData OpData) error {
	hash := keyInfoIDToHash(id)

	opKey, ok := origObj.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	opKey.CreateLogID = oplog.PreLogID

	err := opKey.Save(true)
	if err != nil {
		return err
	}

	log.Debug("DeleteOpKeyPostprocess: to RemoveOpKeyInfoFromHash")

	return pm.RemoveOpKeyInfoFromHash(hash, false, false, false)
}
