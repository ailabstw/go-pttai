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
 * Handle DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) HandleDeleteObjectLog(
	oplog *BaseOplog,
	info ProcessInfo,

	obj Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
	removeInfoByBlockInfo func(blockInfo BlockInfo, info ProcessInfo, oplog *BaseOplog),
	postdelete func(id *types.PttID, oplog *BaseOplog, opData OpData, origObj Object, blockInfo BlockInfo) error,
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,
) ([]*BaseOplog, error) {

	objID := oplog.ObjID
	obj.SetID(objID)

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

	// 3. check validity
	origStatus := origObj.GetStatus()
	if origStatus >= types.StatusDeleted {
		if oplog.UpdateTS.IsLess(origObj.GetUpdateTS()) {
			err = pm.saveDeleteObjectWithOplog(origObj, oplog, types.StatusDeleted, true)
			if err != nil {
				return nil, err
			}
		}
		return nil, ErrNewerOplog
	}

	err = pm.handleDeleteObjectLogCore(oplog, info, obj, opData, setLogDB, removeInfoByBlockInfo, postdelete, updateDeleteInfo)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handleDeleteObjectLogCore(
	oplog *BaseOplog,
	info ProcessInfo,

	origObj Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
	removeInfoByBlockInfo func(blockInfo BlockInfo, info ProcessInfo, oplog *BaseOplog),
	postdelete func(id *types.PttID, oplog *BaseOplog, opData OpData, origObj Object, blockInfo BlockInfo) error,
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,
) error {

	var err error

	objID := origObj.GetID()

	// 1. check sync-info
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), types.StatusInternalDeleted, types.StatusPendingDeleted, types.StatusDeleted)

	var isReplaceSyncInfo bool
	origSyncInfo := origObj.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
	}

	// 1.1. replace sync-info
	if isReplaceSyncInfo {
		err = pm.removeBlockAndInfoBySyncInfo(origSyncInfo, info, oplog, true, removeInfoByBlockInfo, setLogDB)
		if err != nil {
			return err
		}
		origObj.SetSyncInfo(nil)
	}

	// 2. block info
	blockInfo := origObj.GetBlockInfo()
	err = pm.removeBlockAndInfoByBlock(blockInfo, info, oplog, true, removeInfoByBlockInfo)
	if err != nil {
		return err
	}

	// 3. deal with create-log (delete obj completely and return skip log if not me, else do not set sync)
	origStatus := origObj.GetStatus()
	myID := pm.Ptt().GetMyEntity().GetID()
	if origStatus < types.StatusAlive {
		if !reflect.DeepEqual(myID, oplog.CreatorID) {
			origObj.Delete(true)
			return ErrSkipOplog
		}
	}

	// 4. saveDeleteObj
	err = pm.saveDeleteObjectWithOplog(origObj, oplog, types.StatusDeleted, true)
	if err != nil {
		return err
	}

	log.Debug("handleDeleteObjectLogCore: to postdelete")

	// 4.1
	if postdelete != nil {
		postdelete(objID, oplog, opData, origObj, blockInfo)
	}

	log.Debug("handleDeleteObjectLogCore: after postdelete")

	// 5. set oplog is-sync (do not set sync if orig-status is alive)
	if origStatus == types.StatusAlive {
		oplog.IsSync = true
	}

	// 6. updateDeleteInfo
	if info == nil {
		return nil
	}
	updateDeleteInfo(origObj, oplog, info)

	return nil
}

/**********
 * Handle PendingDeleteObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingDeleteObjectLog(
	oplog *BaseOplog,
	info ProcessInfo,

	obj Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
	removeInfoByBlockInfo func(blockInfo BlockInfo, info ProcessInfo, oplog *BaseOplog),
	setPendingDeleteSyncInfo func(origObj Object, status types.Status, oplog *BaseOplog) error,

	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,
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

	// 3. already deleted
	origObj := obj

	origStatus := origObj.GetStatus()
	if origStatus == types.StatusDeleted {
		return nil, ErrNewerOplog
	}

	err = pm.handlePendingDeleteObjectLogCore(oplog, info, origObj, opData, setLogDB, removeInfoByBlockInfo, setPendingDeleteSyncInfo, updateDeleteInfo)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handlePendingDeleteObjectLogCore(
	oplog *BaseOplog,
	info ProcessInfo,

	origObj Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
	removeInfoByBlockInfo func(blockInfo BlockInfo, info ProcessInfo, oplog *BaseOplog),
	setPendingDeleteSyncInfo func(origObj Object, status types.Status, oplog *BaseOplog) error,

	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,
) error {

	var err error

	// 1. sync info
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), types.StatusInternalDeleted, types.StatusPendingDeleted, types.StatusDeleted)

	var isReplaceSyncInfo bool
	origSyncInfo := origObj.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
		if !isReplaceSyncInfo {
			return types.ErrAlreadyPending
		}

		// 1.1 replace sync-info
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			pm.removeBlockAndInfoBySyncInfo(origSyncInfo, info, oplog, false, removeInfoByBlockInfo, setLogDB)
		}
		origObj.SetSyncInfo(nil)
	}

	// 4. save obj
	setPendingDeleteSyncInfo(origObj, oplogStatus, oplog)
	err = origObj.Save(true)
	if err != nil {
		return err
	}

	// 5. oplog.is-sync

	oplog.IsSync = true

	// 6. update delete info
	if info == nil {
		return nil
	}

	updateDeleteInfo(origObj, oplog, info)

	return nil
}

/**********
 * Set Newest DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) SetNewestDeleteObjectLog(
	oplog *BaseOplog,
	obj Object,
) (types.Bool, error) {
	return false, nil
}

/**********
 * Handle Failed DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) HandleFailedDeleteObjectLog(
	oplog *BaseOplog,
	obj Object,
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

	// 5. obj-save
	err = origObj.Save(true)
	if err != nil {
		return err
	}

	return nil
}

/*
SaveDeleteObjectWithOplog saves Delete Object with Oplog.

We can't integrate with postdelete because there are situations that we want to save without postdelete. (already deleted but we have older ts).
*/
func (pm *BaseProtocolManager) saveDeleteObjectWithOplog(
	obj Object,
	oplog *BaseOplog,
	status types.Status,
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

	SetDeleteObjectWithOplog(obj, status, oplog)

	err = obj.Save(true)
	if err != nil {
		return err
	}

	return nil
}
