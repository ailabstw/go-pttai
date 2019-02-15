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
	"bytes"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
)

/**********
 * Handle UpdatePersonLog
 **********/

func (pm *BaseProtocolManager) HandleUpdatePersonLog(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	postupdate func(obj Object, oplog *BaseOplog) error,

) ([]*BaseOplog, error) {

	// 1. lock person
	personID := oplog.ObjID
	origPerson.SetID(personID)

	err := origPerson.Lock()
	if err != nil {
		return nil, err
	}
	defer origPerson.Unlock()

	// 2. get person (should never delete once stored)
	err = origPerson.GetByID(true)
	if err != nil {
		return nil, err
	}

	if oplog.UpdateTS.IsLess(origPerson.GetUpdateTS()) {
		return nil, ErrNewerOplog
	}

	/*
		if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
			return nil, ErrInvalidPreLog
		}
	*/

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus == types.StatusAlive {
		if oplog.UpdateTS.IsLess(origPerson.GetUpdateTS()) {
			err = pm.saveUpdateObjectWithOplog(origPerson, oplog, true)
			if err != nil {
				return nil, err
			}
		}
		return nil, ErrNewerOplog
	}
	if origStatus == types.StatusTransferred {
		return nil, ErrNewerOplog
	}

	// 4. core
	err = pm.handleUpdatePersonLogCore(
		oplog,
		origPerson,
		opData,

		merkle,

		setLogDB,
		postupdate,
	)

	if err != nil {
		return nil, err
	}

	return nil, nil

}

func (pm *BaseProtocolManager) handleUpdatePersonLogCore(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	postupdate func(obj Object, oplog *BaseOplog) error,

) error {

	var err error

	// 1. check sync-info
	oplogStatus := types.StatusAlive

	var isReplaceSyncInfo bool
	origSyncInfo := origPerson.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncPersonInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
	}

	// 1.1. replace sync-info
	if isReplaceSyncInfo {
		err = pm.removeBlockAndMediaInfoBySyncInfo(
			origSyncInfo,

			nil,
			oplog,
			true,

			merkle,

			nil,
			setLogDB,
		)
		if err != nil {
			return err
		}
		origPerson.SetSyncInfo(nil)
	}

	// 4. saveUpdateObj
	err = pm.saveUpdateObjectWithOplog(origPerson, oplog, true)
	if err != nil {
		return err
	}

	if postupdate != nil {
		postupdate(origPerson, oplog)
	}

	oplog.IsSync = true

	return nil
}

/**********
 * Handle UpdatePendingPersonLog
 **********/

func (pm *BaseProtocolManager) HandlePendingUpdatePersonLog(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

) (types.Bool, []*BaseOplog, error) {

	// 1. lock person
	personID := oplog.ObjID
	origPerson.SetID(personID)

	err := origPerson.Lock()
	if err != nil {
		return false, nil, err
	}
	defer origPerson.Unlock()

	// 2. get person
	err = origPerson.GetByID(true)
	if err != nil {
		return false, nil, err
	}

	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		return false, nil, ErrInvalidPreLog
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus == types.StatusAlive {
		return false, nil, ErrNewerOplog
	}
	if origStatus == types.StatusTransferred {
		return false, nil, ErrNewerOplog
	}

	// 4. core
	err = pm.handlePendingUpdatePersonLogCore(
		oplog,

		origPerson,
		opData,

		merkle,

		setLogDB,
	)

	if err != nil {
		return false, nil, err
	}

	return oplog.IsSync, nil, nil
}

func (pm *BaseProtocolManager) handlePendingUpdatePersonLogCore(
	oplog *BaseOplog,

	origObj Object,
	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

) error {

	var err error

	// 1. sync info
	oplogStatus := oplog.ToStatus()

	var isReplaceSyncInfo bool
	origSyncInfo := origObj.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncPersonInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
		if !isReplaceSyncInfo {
			return types.ErrAlreadyPending
		}

		// 1.1 replace sync-info
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			pm.removeBlockAndMediaInfoBySyncInfo(
				origSyncInfo,

				nil,
				oplog,
				false,

				merkle,

				nil,
				setLogDB,
			)
		}
		origObj.SetSyncInfo(nil)
	}

	// 4. save obj
	SetPendingPersonSyncInfo(origObj, oplogStatus, oplog)
	err = origObj.Save(true)
	if err != nil {
		return err
	}

	// 5. oplog.is-sync

	oplog.IsSync = true

	return nil
}

func SetPendingPersonSyncInfo(person Object, pendingStatus types.Status, oplog *BaseOplog) error {

	syncInfo := NewEmptySyncPersonInfo()
	syncInfo.InitWithOplog(pendingStatus, oplog)

	person.SetSyncInfo(syncInfo)

	return nil
}

func isReplaceOrigSyncPersonInfo(syncInfo SyncInfo, status types.Status, ts types.Timestamp, newLogID *types.PttID) bool {

	if syncInfo == nil {
		return true
	}

	statusClass := types.StatusToStatusClass(status)
	syncStatusClass := types.StatusToStatusClass(syncInfo.GetStatus())

	switch syncStatusClass {
	case types.StatusClassInternalDelete:
		syncStatusClass = types.StatusClassInternalPendingAlive
	case types.StatusClassPendingDelete:
		syncStatusClass = types.StatusClassPendingAlive
	case types.StatusClassDeleted:
		syncStatusClass = types.StatusClassAlive
	}

	if statusClass < syncStatusClass {
		return false
	}
	if statusClass > syncStatusClass {
		return true
	}

	syncTS := syncInfo.GetUpdateTS()
	if syncTS.IsLess(ts) {
		return false
	}
	if ts.IsLess(syncTS) {
		return true
	}

	origLogID := syncInfo.GetLogID()
	return bytes.Compare(origLogID[:], newLogID[:]) > 0
}
