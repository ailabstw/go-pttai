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
	"github.com/syndtr/goleveldb/leveldb"
)

/**********
 * Handle CreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandleCreateObjectLog(
	oplog *BaseOplog,
	obj Object,
	opData OpData,
	info ProcessInfo,

	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	newObjWithOplog func(oplog *BaseOplog, opData OpData) Object,
	postcreateObject func(obj Object, oplog *BaseOplog) error,
	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,

) ([]*BaseOplog, error) {

	err := pm.handleCreateObjectLogCore(oplog, obj, opData, info, existsInInfo, newObjWithOplog, postcreateObject, updateCreateInfo)
	return nil, err
}

/**********
 * Handle PendingCreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingCreateObjectLog(
	oplog *BaseOplog,
	obj Object,
	opData OpData,
	info ProcessInfo,

	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	newObjWithOplog func(oplog *BaseOplog, opData OpData) Object,
	postcreateObject func(obj Object, oplog *BaseOplog) error,
	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,
) ([]*BaseOplog, error) {

	err := pm.handleCreateObjectLogCore(oplog, obj, opData, info, existsInInfo, newObjWithOplog, postcreateObject, updateCreateInfo)
	return nil, err
}

/*
handleCreateObjectLogCore deals with create-object logs

We need existsInInfo because the object may be already deleted by parent objects (comment vs article)
*/
func (pm *BaseProtocolManager) handleCreateObjectLogCore(
	oplog *BaseOplog,
	obj Object,
	opData OpData,
	info ProcessInfo,

	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	newObjWithOplog func(oplog *BaseOplog, opData OpData) Object,
	postcreate func(obj Object, oplog *BaseOplog) error,

	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,
) error {

	objID := oplog.ObjID
	obj.SetID(objID)

	err := oplog.GetData(opData)
	if err != nil {
		return err
	}

	// lock obj
	err = obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	// get obj
	err = obj.GetByID(true)
	if err == leveldb.ErrNotFound {
		log.Debug("handleCreateObjectLogCore: to handleCreateObjectNewLog", "oplog", oplog.ID, "obj", objID)
		return pm.handleCreateObjectNewLog(oplog, opData, info, existsInInfo, newObjWithOplog, updateCreateInfo)
	}
	if err != nil {
		return err
	}

	// orig-obj exists
	origObj := obj

	// newest-orig-log-id
	newestLogID := origObj.GetUpdateLogID()
	if newestLogID == nil {
		newestLogID = origObj.GetLogID()
	}

	// same log
	if reflect.DeepEqual(newestLogID, oplog.ID) {
		log.Debug("handleCreateObjectLogCore: to handleCreateObjectSameLog", "oplog", oplog.ID, "obj", objID)

		return pm.handleCreateObjectSameLog(oplog, origObj, opData, info, postcreate, updateCreateInfo)
	}

	log.Debug("handleCreateObjectLogCore: to handleCreateObjectDiffLog", "oplog", oplog.ID, "obj", objID)

	return pm.handleCreateObjectDiffLog(oplog, origObj, info)
}

func (pm *BaseProtocolManager) handleCreateObjectNewLog(
	oplog *BaseOplog,
	opData OpData,
	info ProcessInfo,

	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	newObjWithOplog func(oplog *BaseOplog, opData OpData) Object,

	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,
) error {

	isExists, err := existsInInfo(oplog, info)
	log.Debug("handleCreateObjectNewLog: after existsInInfo", "oplog", oplog.ID, "isExists", isExists, "e", err)
	if err != nil {
		return err
	}
	if isExists {
		return nil
	}

	obj := newObjWithOplog(oplog, opData)
	if obj == nil {
		return ErrInvalidOplog
	}
	err = obj.Save(true)
	log.Debug("handleCreateObjectNewLog: after newObjWithOplog", "oplog", oplog.ID, "obj", obj.GetID(), "obj.Status", obj.GetStatus(), "oplog.IsNewer", oplog.IsNewer)
	if err != nil {
		return err
	}

	if oplog.IsNewer {
		return nil
	}

	return updateCreateInfo(obj, oplog, opData, info)
}

func (pm *BaseProtocolManager) handleCreateObjectSameLog(
	oplog *BaseOplog,
	origObj Object,
	opData OpData,
	info ProcessInfo,

	postcreate func(obj Object, oplog *BaseOplog) error,

	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,
) error {

	var err error

	origStatus := origObj.GetStatus()

	if origStatus == types.StatusInternalSync {
		// still in sync, requesting again.
		if !oplog.IsNewer {
			err = updateCreateInfo(origObj, oplog, opData, info)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// although we got the content synced:
	// 1. the oplog-status may change.
	// 2. we may get older date if the oplog-status is not changed.
	// => check status and do saveNewObjectWithOplog.
	oplogStatus := oplog.ToStatus()

	origTS := origObj.GetUpdateTS()
	if oplogStatus < origStatus || oplogStatus == origStatus && origTS.IsLessEqual(oplog.UpdateTS) {
		return ErrNewerOplog
	}

	log.Debug("handleCreateObjectSameLog: to saveNewObjectWithOplog", "oplogStatus", oplogStatus, "origStatus", origStatus)

	if oplogStatus == origStatus {
		// the status is already the same, we just update the object without postcreate.
		return pm.saveNewObjectWithOplog(origObj, oplog, true, true, postcreate)
	}

	// We got higher oplogStatus. Do saveNewObjectWithOplog with postcreate
	return pm.saveNewObjectWithOplog(origObj, oplog, true, false, postcreate)
}

func (pm *BaseProtocolManager) handleCreateObjectDiffLog(
	oplog *BaseOplog,
	origObj Object,
	info ProcessInfo,
) error {

	return ErrNewerOplog
}
