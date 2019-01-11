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
	"github.com/ailabstw/go-pttai/log"
)

/**********
 * Handle DeleteObjectLog
 **********/

func (pm *BaseProtocolManager) HandleUpdateObjectLog(
	oplog *BaseOplog,
	opData OpData,

	obj Object,

	info ProcessInfo,

	merkle *Merkle,

	syncInfoFromOplog func(oplog *BaseOplog, opData OpData) (SyncInfo, error),

	setLogDB func(oplog *BaseOplog),
	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,

	updateUpdateInfo func(obj Object, oplog *BaseOplog, opData OpData, origSyncInfo SyncInfo, info ProcessInfo) error,

) ([]*BaseOplog, error) {

	err := pm.handleUpdateObjectCore(
		oplog,
		opData,

		obj,

		info,

		true,

		merkle,

		syncInfoFromOplog,
		setLogDB,
		removeMediaInfoByBlockInfo,
		postupdate,
		updateUpdateInfo,
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) HandlePendingUpdateObjectLog(
	oplog *BaseOplog,
	opData OpData,

	obj Object,

	info ProcessInfo,

	merkle *Merkle,

	syncInfoFromOplog func(oplog *BaseOplog, opData OpData) (SyncInfo, error),

	setLogDB func(oplog *BaseOplog),
	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	postUpdate func(obj Object, oplog *BaseOplog) error,

	updateUpdateInfo func(obj Object, oplog *BaseOplog, opData OpData, origSyncInfo SyncInfo, info ProcessInfo) error,

) (types.Bool, []*BaseOplog, error) {

	err := pm.handleUpdateObjectCore(
		oplog,
		opData,

		obj,

		info,

		false,

		merkle,

		syncInfoFromOplog,
		setLogDB,
		removeMediaInfoByBlockInfo,
		postUpdate,
		updateUpdateInfo,
	)
	if err != nil {
		return false, nil, err
	}

	return oplog.IsSync, nil, nil
}

func (pm *BaseProtocolManager) handleUpdateObjectCore(
	oplog *BaseOplog,
	opData OpData,

	obj Object,

	info ProcessInfo,

	isRetainValid bool,

	merkle *Merkle,

	syncInfoFromOplog func(oplog *BaseOplog, opData OpData) (SyncInfo, error),

	setLogDB func(oplog *BaseOplog),
	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,

	updateUpdateInfo func(obj Object, oplog *BaseOplog, opData OpData, origSyncInfo SyncInfo, info ProcessInfo) error,

) error {

	// 1. lock-obj
	objID := oplog.ObjID

	obj.SetID(objID)
	err := obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	// orig-article (possibly never exists or already deleted)
	err = obj.GetByID(true)
	if err != nil {
		log.Warn("handleUpdateObjectCore: unable to get obj", "objID", objID, "oplog", oplog.ID, "e", err)
		return ErrNewerOplog
	}

	origObj := obj

	// already deleted
	objStatus := origObj.GetStatus()

	if objStatus >= types.StatusDeleted {
		return ErrNewerOplog
	}

	// prelog
	if !reflect.DeepEqual(origObj.GetLogID(), oplog.PreLogID) {
		return ErrInvalidPreLog
	}

	// already updated
	// we want to have older ts if we are in the same log.
	if reflect.DeepEqual(oplog.ID, origObj.GetUpdateLogID()) {
		if reflect.DeepEqual(oplog.ID, origObj.GetUpdateLogID()) {
			origObj.SetUpdateTS(oplog.UpdateTS)
			return origObj.Save(true)
		}
		return ErrNewerOplog
	}

	// newer-oplog
	// we want to have newer ts and return newer if we are diff logs.
	if oplog.UpdateTS.IsLess(origObj.GetUpdateTS()) {
		return ErrNewerOplog
	}

	if oplog.IsNewer {
		return ErrNewerOplog
	}

	// set op-data
	err = oplog.GetData(opData)
	if err != nil {
		return err
	}

	// new sync info
	newSyncInfo, err := syncInfoFromOplog(oplog, opData)
	if err != nil {
		return err
	}
	newSyncInfo.SetStatus(types.StatusInternalSync)

	return pm.handleUpdateObjectCoreCore(
		oplog,
		opData,

		origObj,
		newSyncInfo,

		info,

		isRetainValid,

		merkle,

		setLogDB,
		removeMediaInfoByBlockInfo,
		postupdate,
		updateUpdateInfo,
	)
}

/*
handleUpdateObjectCoreCore deals with the core of handleUpdateObjectCore.

Requiring oplog, filled-data op-data, orig-obj
*/
func (pm *BaseProtocolManager) handleUpdateObjectCoreCore(
	oplog *BaseOplog,
	opData OpData,

	origObj Object,
	newSyncInfo SyncInfo,

	info ProcessInfo,

	isRetainValid bool,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,

	updateUpdateInfo func(obj Object, oplog *BaseOplog, opData OpData, origSyncInfo SyncInfo, info ProcessInfo) error,

) error {

	removeSyncInfo, err := pm.handleUpdateObjectWithNewSyncInfo(
		origObj,
		newSyncInfo,

		oplog,

		info,

		isRetainValid,

		merkle,

		setLogDB,
		removeMediaInfoByBlockInfo,
		postupdate,
	)

	if err != nil {
		return err
	}

	// update info
	if info == nil || updateUpdateInfo == nil {
		return nil
	}

	err = updateUpdateInfo(origObj, oplog, opData, removeSyncInfo, info)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) handleUpdateObjectWithNewSyncInfo(
	obj Object,
	newSyncInfo SyncInfo,

	oplog *BaseOplog,

	info ProcessInfo,

	isRetainValid bool,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,

) (SyncInfo, error) {

	// new sync log
	var err error

	syncInfo := obj.GetSyncInfo()
	if syncInfo == nil {
		err = pm.handleUpdateObjectNewLog(obj, newSyncInfo, oplog, postupdate)
		return nil, err
	}

	status := oplog.ToStatus()

	// not replace
	isReplaceSyncInfo := isReplaceOrigSyncInfo(syncInfo, status, oplog.UpdateTS, oplog.ID)
	if !isReplaceSyncInfo {
		return nil, ErrNewerOplog
	}

	syncLogID := syncInfo.GetLogID()
	if reflect.DeepEqual(syncLogID, oplog.ID) {
		err = pm.handleUpdateObjectSameLog(obj, newSyncInfo, oplog, postupdate)
		return nil, err
	}

	origSyncInfo := obj.GetSyncInfo()
	err = pm.handleUpdateObjectDiffLog(
		obj,
		syncInfo,
		newSyncInfo,

		oplog,

		info,

		isRetainValid,

		merkle,

		setLogDB,
		removeMediaInfoByBlockInfo,
		postupdate,
	)

	return origSyncInfo, err
}

func (pm *BaseProtocolManager) handleUpdateObjectSameLog(
	obj Object,
	newSyncInfo SyncInfo,

	oplog *BaseOplog,

	postupdate func(obj Object, oplog *BaseOplog) error,
) error {

	obj.SetSyncInfo(nil)

	return pm.handleUpdateObjectNewLog(obj, newSyncInfo, oplog, postupdate)

}

func (pm *BaseProtocolManager) handleUpdateObjectDiffLog(
	obj Object,
	origSyncInfo SyncInfo,
	newSyncInfo SyncInfo,

	oplog *BaseOplog,

	info ProcessInfo,

	isRetainValid bool,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,

) error {
	err := pm.removeBlockAndMediaInfoBySyncInfo(
		origSyncInfo,

		info,
		oplog,
		isRetainValid,

		merkle,

		removeMediaInfoByBlockInfo,
		setLogDB,
	)
	if err != nil {
		return err
	}

	obj.SetSyncInfo(nil)

	return pm.handleUpdateObjectNewLog(obj, newSyncInfo, oplog, postupdate)
}

func (pm *BaseProtocolManager) handleUpdateObjectNewLog(
	obj Object,
	newSyncInfo SyncInfo,

	oplog *BaseOplog,

	postupdate func(obj Object, oplog *BaseOplog) error,
) error {

	var err error

	isAllGood := newSyncInfo.CheckIsAllGood()

	status := oplog.ToStatus()

	if isAllGood && status == types.StatusAlive {
		err = newSyncInfo.ToObject(obj)
		if err != nil {
			return err
		}

		// save
		err = pm.saveUpdateObjectWithOplog(obj, oplog, true)
		if err != nil {
			return err
		}

		if postupdate == nil {
			return nil
		}

		return postupdate(obj, oplog)
	}

	// not synced yet.
	if isAllGood {
		newSyncInfo.SetStatus(status)
	}
	obj.SetSyncInfo(newSyncInfo)

	err = obj.Save(true)
	if err != nil {
		return err
	}

	return nil
}

func isReplaceOrigSyncInfo(
	syncInfo SyncInfo,
	status types.Status,
	ts types.Timestamp,
	newLogID *types.PttID,
) bool {

	if syncInfo == nil {
		return true
	}

	statusClass := types.StatusToStatusClass(status)
	syncStatusClass := types.StatusToStatusClass(syncInfo.GetStatus())

	switch syncStatusClass {
	case types.StatusClassInternalMigrate:
		syncStatusClass = types.StatusClassInternalDelete
	case types.StatusClassPendingMigrate:
		syncStatusClass = types.StatusClassPendingDelete
	case types.StatusClassMigrated:
		syncStatusClass = types.StatusClassDeleted
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

/*
SaveUpdateObjectWithOplog saves Update Object with Oplog.

We can't integrate with postupdate because there are situations that we want to save without postupdate. (already updated but we have older ts).
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
	oplog.IsSync = true

	err = obj.Save(true)
	if err != nil {
		return err
	}

	return nil
}
