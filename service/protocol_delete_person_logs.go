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
	"github.com/ailabstw/go-pttai/log"
)

/**********
 * Handle DeleteRenewableObjectLog
 **********/

/**********
 * Handle DeleteRenewableObjectLog
 **********/

func (pm *BaseProtocolManager) HandleDeletePersonLog(
	oplog *BaseOplog,
	info ProcessInfo,

	origPerson Object,
	opData OpData,

	status types.Status,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

	postdelete func(id *types.PttID, oplog *BaseOplog, origPerson Object, opData OpData) error,
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,

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
	if origStatus == types.StatusDeleted {
		if oplog.UpdateTS.IsLess(origPerson.GetUpdateTS()) {
			err = pm.saveUpdateObjectWithOplog(origPerson, oplog, true)
			if err != nil {
				return nil, err
			}
		}
		return nil, ErrNewerOplog
	}
	if origStatus >= types.StatusMigrated {
		return nil, ErrNewerOplog
	}

	err = pm.handleDeletePersonLogCore(
		oplog,
		info,

		origPerson,
		opData,

		status,

		merkle,

		setLogDB,
		postdelete,
		updateDeleteInfo,
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handleDeletePersonLogCore(
	oplog *BaseOplog,
	info ProcessInfo,

	origPerson Object,
	opData OpData,

	oplogStatus types.Status,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	postdelete func(id *types.PttID, oplog *BaseOplog, origPerson Object, opData OpData) error,
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,

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
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
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
		}
		origPerson.SetSyncInfo(nil)
	}

	// 4. saveDeleteObj
	err = pm.saveDeleteObjectWithOplog(origPerson, oplog, oplogStatus, true)
	if err != nil {
		return err
	}

	if postdelete != nil {
		postdelete(oplog.ObjID, oplog, origPerson, opData)
	}

	oplog.IsSync = true

	// 6. updateDeleteInfo
	if info == nil {
		return nil
	}
	updateDeleteInfo(origPerson, oplog, info)

	return nil
}

/**********
 * Handle PendingDeleteRenewableObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingDeletePersonLog(
	oplog *BaseOplog,
	info ProcessInfo,

	origPerson Object,
	opData OpData,

	internalPendingStatus types.Status,
	pendingStatus types.Status,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	updateDeleteInfo func(person Object, oplog *BaseOplog, info ProcessInfo) error,

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
	log.Debug("HandlePendingDeletePersonLog: after GetByID", "e", err)
	if err != nil {
		return false, nil, err
	}

	if oplog.UpdateTS.IsLess(origPerson.GetUpdateTS()) {
		return false, nil, ErrNewerOplog
	}

	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		log.Warn("HandlePendingDeletePersonLog: pre-log-id", "person", origPerson.GetLogID(), "preLogID", oplog.PreLogID, "entity", pm.Entity().GetID())
		return false, nil, ErrInvalidPreLog
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus >= types.StatusMigrated {
		return false, nil, ErrNewerOplog
	}

	// 4. core
	log.Debug("HandlePendingDeletePersonLog: to handlePendingDeletePersonLogCore", "entity", pm.Entity().GetID())
	err = pm.handlePendingDeletePersonLogCore(
		oplog,
		info,

		origPerson,
		opData,

		internalPendingStatus,
		pendingStatus,

		merkle,

		setLogDB,
		updateDeleteInfo,
	)
	log.Debug("HandlePendingDeletePersonLog: after handlePendingDeletePersonLogCore", "entity", pm.Entity().GetID(), "e", err)
	if err != nil {
		return false, nil, err
	}

	pm.InternalSign(oplog)

	myID := pm.Ptt().GetMyEntity().GetID()
	if reflect.DeepEqual(myID, personID) && len(oplog.MasterSigns) == 1 {
		oplog.SetMasterLogID(pm.GetNewestMasterLogID(), 0)
	}

	return true, nil, nil
}

func (pm *BaseProtocolManager) handlePendingDeletePersonLogCore(
	oplog *BaseOplog,
	info ProcessInfo,

	origPerson Object,
	opData OpData,

	internalPendingStatus types.Status,
	pendingStatus types.Status,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	updateDeleteInfo func(person Object, oplog *BaseOplog, info ProcessInfo) error,

) error {

	var err error

	// 1. sync info
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), internalPendingStatus, pendingStatus, types.StatusDeleted)

	var isReplaceSyncInfo bool
	origSyncInfo := origPerson.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncPersonInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
		if !isReplaceSyncInfo {
			return types.ErrAlreadyPending
		}

		// 1.1 replace sync-info
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			err = pm.removeBlockAndMediaInfoBySyncInfo(
				origSyncInfo,

				nil,
				oplog,
				false,

				merkle,

				nil,
				setLogDB,
			)
			if err != nil {
				return err
			}
		}
		origPerson.SetSyncInfo(nil)
	}

	// 4. save person
	SetPendingPersonSyncInfo(origPerson, oplogStatus, oplog)
	err = origPerson.Save(true)
	if err != nil {
		return err
	}

	// 6. info
	if info == nil {
		return nil
	}

	updateDeleteInfo(origPerson, oplog, info)

	return nil
}
