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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
)

/**********
 * Handle DeleteRenewableObjectLog
 **********/

func (pm *BaseProtocolManager) HandleDeletePersonLog(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

	status types.Status,

	setLogDB func(oplog *BaseOplog),

	postdelete func(id *types.PttID, oplog *BaseOplog, origPerson Object, opData OpData) error,

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
	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		return nil, ErrInvalidPreLog
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus == types.StatusDeleted {
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

	err = pm.handleDeletePersonLogCore(oplog, origPerson, opData, status, setLogDB, postdelete)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handleDeletePersonLogCore(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

	oplogStatus types.Status,

	setLogDB func(oplog *BaseOplog),
	postdelete func(id *types.PttID, oplog *BaseOplog, origPerson Object, opData OpData) error,

) error {

	var err error

	// 1. check sync-info
	var isReplaceSyncInfo bool
	origSyncInfo := origPerson.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncPersonInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
	}

	// 1.1. replace sync-info
	if isReplaceSyncInfo {
		err = pm.removeBlockAndMediaInfoBySyncInfo(origSyncInfo, nil, oplog, true, nil, setLogDB)
		if err != nil {
			return err
		}
		origPerson.SetSyncInfo(nil)
	}

	// 4. saveUpdateObj
	err = pm.saveDeleteObjectWithOplog(origPerson, oplog, oplogStatus, true)
	if err != nil {
		return err
	}

	if postdelete != nil {
		postdelete(oplog.ObjID, oplog, origPerson, opData)
	}

	oplog.IsSync = true

	return nil
}

/**********
 * Handle PendingDeleteRenewableObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingDeletePersonLog(
	oplog *BaseOplog, info ProcessInfo,
	origPerson Object,
	opData OpData,

	internalPendingStatus types.Status,
	pendingStatus types.Status,

	setLogDB func(oplog *BaseOplog),
) ([]*BaseOplog, error) {

	// 1. lock person
	personID := oplog.ObjID
	origPerson.SetID(personID)

	err := origPerson.Lock()
	if err != nil {
		return nil, err
	}
	defer origPerson.Unlock()

	// 2. get person
	err = origPerson.GetByID(true)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		return nil, ErrInvalidPreLog
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus == types.StatusAlive {
		return nil, ErrNewerOplog
	}
	if origStatus == types.StatusTransferred {
		return nil, ErrNewerOplog
	}

	// 4. core
	err = pm.handlePendingDeletePersonLogCore(oplog, origPerson, opData, internalPendingStatus, pendingStatus, setLogDB)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handlePendingDeletePersonLogCore(
	oplog *BaseOplog,

	origObj Object,
	opData OpData,

	internalPendingStatus types.Status,
	pendingStatus types.Status,

	setLogDB func(oplog *BaseOplog),

) error {

	var err error

	// 1. sync info
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), internalPendingStatus, pendingStatus, types.StatusDeleted)

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
			pm.removeBlockAndMediaInfoBySyncInfo(origSyncInfo, nil, oplog, false, nil, setLogDB)
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
