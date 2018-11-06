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
 * Handle UpdatePersonLog
 **********/

func (pm *BaseProtocolManager) HandleUpdatePersonLog(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

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
	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		return nil, ErrInvalidPreLog
	}

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
	err = pm.handleUpdatePersonLogCore(oplog, origPerson, opData, setLogDB, postupdate)
	if err != nil {
		return nil, err
	}

	return nil, nil

}

func (pm *BaseProtocolManager) handleUpdatePersonLogCore(
	oplog *BaseOplog,
	origPerson Object,
	opData OpData,

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
		err = pm.removeBlockAndInfoBySyncInfo(origSyncInfo, nil, oplog, true, nil, setLogDB)
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
	err = pm.handlePendingUpdatePersonLogCore(oplog, origPerson, opData, setLogDB)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handlePendingUpdatePersonLogCore(
	oplog *BaseOplog,

	origObj Object,
	opData OpData,

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
			pm.removeBlockAndInfoBySyncInfo(origSyncInfo, nil, oplog, false, nil, setLogDB)
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

func (pm *BaseProtocolManager) HandleFailedUpdatePersonLog(
	oplog *BaseOplog,
	origPerson Object,
) error {
	return pm.HandleFailedPersonLog(oplog, origPerson)
}

/*
SaveDeleteObjectWithOplog saves Delete Object with Oplog.

We can't integrate with postdelete because there are situations that we want to save without postdelete. (already deleted but we have older ts).
*/
func (pm *BaseProtocolManager) saveUpdateObjectWithOplog(
	obj Object,
	oplog *BaseOplog,
	isLocked bool,

) error {

	var err error
	if !isLocked {
		err = obj.Lock()
		if err != nil {
			return err
		}
		defer obj.Unlock()
	}

	SetUpdateObjectWithOplog(obj, oplog)

	err = obj.Save(true)
	if err != nil {
		return err
	}

	return nil
}

func SetPendingPersonSyncInfo(person Object, pendingStatus types.Status, oplog *BaseOplog) error {

	syncInfo := NewEmptySyncPersonInfo()
	syncInfo.InitWithOplog(oplog)
	syncInfo.Status = pendingStatus

	person.SetSyncInfo(syncInfo)

	return nil
}

func (pm *BaseProtocolManager) HandleFailedPersonLog(
	oplog *BaseOplog,
	person Object,
) error {

	objID := oplog.ObjID
	person.SetID(objID)

	// 1. lock
	err := person.Lock()
	if err != nil {
		return err
	}
	defer person.Unlock()

	// 2. get obj
	err = person.GetByID(true)
	if err != nil {
		return err
	}

	// 3. check validity
	origSyncInfo := person.GetSyncInfo()
	if origSyncInfo == nil || !reflect.DeepEqual(origSyncInfo.GetLogID(), oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(origSyncInfo.GetUpdateTS()) {
		return nil
	}

	// 4. remove block/oplog
	blockInfo := origSyncInfo.GetBlock()
	err = pm.removeBlockAndInfoByBlock(blockInfo, nil, oplog, true, nil)
	if err != nil {
		return err
	}
	person.SetSyncInfo(nil)

	// 5. obj-save
	err = person.Save(true)
	if err != nil {
		return err
	}

	return nil
}
