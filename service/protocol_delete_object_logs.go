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
 * Handle DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) HandleDeleteObjectLog(
	oplog *BaseOplog,
	info ProcessInfo,

	obj Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
	postdelete func(id *types.PttID, oplog *BaseOplog, origObj Object, opData OpData) error,
) ([]*BaseOplog, error) {

	objID := oplog.ObjID
	obj.SetID(objID)

	toBroadcastLogs := make([]*BaseOplog, 0, 1) // orig-log-id and sync-log-id

	err := oplog.GetData(opData)
	if err != nil {
		return nil, err
	}

	// 1. lock obj
	err = obj.Lock()
	if err != nil {
		return nil, err
	}
	defer obj.Unlock()

	// 2. get obj
	err = obj.GetByID(true)
	if err != nil {
		return nil, ErrNewerOplog
	}

	origObj := obj

	// 3. already deleted
	origStatus := origObj.GetStatus()
	if origStatus == types.StatusDeleted {
		if oplog.UpdateTS.IsLess(origObj.GetUpdateTS()) {
			err = pm.saveDeleteObjectWithOplog(origObj, oplog, true)
			if err != nil {
				return nil, err
			}
		}
		return nil, ErrNewerOplog
	}

	// 4. sync-info
	origSyncInfo := origObj.GetSyncInfo()
	if origSyncInfo != nil {
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			err = origObj.RemoveSyncInfo(oplog, opData, origSyncInfo, info)
			if err != nil {
				return nil, err
			}

			_, err := pm.RemoveNonSyncOplog(setLogDB, syncLogID, true, false)
			if err != nil {
				return nil, err
			}
		}

		origObj.SetSyncInfo(nil)
	}

	// 5. remove orig info
	err = origObj.RemoveBlock(origObj.GetBlockInfo(), info, true)
	if err != nil {
		return nil, err
	}

	// 6. deal with create-log
	if origStatus < types.StatusAlive {
		createLog, err := pm.SetOplogIDIsSync(origObj.GetLogID())
		if err != nil {
			return nil, err
		}
		if createLog != nil {
			createLog.IsNewer = true
			toBroadcastLogs = append(toBroadcastLogs, createLog)
		}
	}

	// 7. saveDeleteObj
	pm.saveDeleteObjectWithOplog(origObj, oplog, true)

	// 7.1
	if postdelete != nil {
		postdelete(objID, oplog, origObj, opData)
	}

	// 8. set oplog is-sync
	oplog.IsSync = true

	// 8. updateDeleteInfo
	origObj.UpdateDeleteInfo(oplog, info)

	return toBroadcastLogs, nil

}

/**********
 * Handle PendingCreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingDeleteObjectLog(
	oplog *BaseOplog, info ProcessInfo,

	obj Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
) ([]*BaseOplog, error) {

	objID := oplog.ObjID
	obj.SetID(objID)

	// 1. lock obj
	err := obj.Lock()
	if err != nil {
		return nil, err
	}
	defer obj.Unlock()

	// 2. get obj
	err = obj.GetByID(true)
	if err != nil {
		return nil, ErrNewerOplog
	}

	origObj := obj

	// 3. already deleted
	origStatus := origObj.GetStatus()
	if origStatus == types.StatusDeleted {
		return nil, ErrNewerOplog
	}

	// 4. sync info
	origSyncInfo := origObj.GetSyncInfo()
	if origSyncInfo != nil {
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			err = oplog.GetData(opData)
			if err != nil {
				return nil, err
			}

			err = origObj.RemoveSyncInfo(oplog, opData, origSyncInfo, info)
			if err != nil {
				return nil, err
			}

			_, err := pm.RemoveNonSyncOplog(setLogDB, syncLogID, false, false)
			if err != nil {
				return nil, err
			}
		}

		origObj.SetSyncInfo(nil)
	}

	// 5. save obj
	origObj.SetPendingDeleteSyncInfo(oplog)
	err = origObj.Save(true)
	if err != nil {
		return nil, err
	}

	// 6. update delete info
	origObj.UpdateDeleteInfo(oplog, info)

	return nil, nil
}

/**********
 * Set Newest DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) SetNewestDeleteObjectLog(
	oplog *BaseOplog, obj Object,
) (types.Bool, error) {
	return false, nil
}

/**********
 * Handle Failed DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) HandleFailedDeleteObjectLog(
	oplog *BaseOplog, obj Object,
) error {

	objID := oplog.ObjID
	obj.SetID(objID)

	// 1. lock obj
	err := obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	// 2. get orig-obj (possibly already deleted)
	err = obj.GetByID(true)
	if err != nil {
		return nil
	}

	origObj := obj

	// 3. check validity
	syncInfo := origObj.GetSyncInfo()
	if syncInfo == nil || !reflect.DeepEqual(syncInfo.GetLogID(), oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(syncInfo.GetUpdateTS()) {
		return nil
	}

	// 4. handle fails
	origObj.SetSyncInfo(nil)
	err = origObj.Save(true)
	if err != nil {
		return err
	}

	return nil
}

/**********
 * save Delete Object with Oplog
 **********/

func (pm *BaseProtocolManager) saveDeleteObjectWithOplog(obj Object, oplog *BaseOplog, isLocked bool) error {

	var err error
	if !isLocked {
		err = obj.Lock()
		if err != nil {
			return err
		}
		defer obj.Unlock()
	}

	SetDeleteObjectWithOplog(obj, oplog)

	err = obj.Save(true)
	if err != nil {
		return err
	}

	return nil
}
