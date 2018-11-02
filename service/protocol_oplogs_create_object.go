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
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
)

/**********
 * Handle CreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandleCreateObjectLog(
	oplog *BaseOplog, obj Object,
	opData OpData,
	info ProcessInfo,
	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
) ([]*BaseOplog, error) {

	err := pm.syncCreateObjectLog(oplog, obj, opData, info, existsInInfo, postprocessCreateObject)
	return nil, err
}

/**********
 * Handle PendingCreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingCreateObjectLog(
	oplog *BaseOplog, obj Object,
	opData OpData,
	info ProcessInfo,
	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
) ([]*BaseOplog, error) {

	err := pm.syncCreateObjectLog(oplog, obj, opData, info, existsInInfo, postprocessCreateObject)
	return nil, err
}

func (pm *BaseProtocolManager) syncCreateObjectLog(
	oplog *BaseOplog, obj Object,
	opData OpData,
	info ProcessInfo,
	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
) error {

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
		return pm.syncCreateObjectNewLog(oplog, obj, opData, info, existsInInfo)
	}
	if err != nil {
		return err
	}

	// orig-obj exists

	// newest-orig-log-id
	newestLogID := obj.GetUpdateLogID()
	if newestLogID == nil {
		newestLogID = obj.GetLogID()
	}

	// same log
	if reflect.DeepEqual(newestLogID, oplog.ID) {
		return pm.syncCreateObjectSameLog(oplog, obj, opData, info, postprocessCreateObject)
	}

	return pm.syncCreateObjectDiffLog(oplog, obj, info)
}

func (pm *BaseProtocolManager) syncCreateObjectNewLog(
	oplog *BaseOplog, obj Object,
	opData OpData,
	info ProcessInfo,
	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
) error {

	isExists, err := existsInInfo(oplog, info)
	if err != nil {
		return err
	}
	if isExists {
		return nil
	}

	err = obj.NewObjWithOplog(oplog, opData)
	if err != nil {
		return err
	}

	err = obj.Save(true)
	if err != nil {
		return err
	}

	if oplog.IsNewer {
		return nil
	}

	return obj.UpdateCreateInfo(oplog, opData, info)
}

func (pm *BaseProtocolManager) syncCreateObjectSameLog(
	oplog *BaseOplog, origObj Object,
	opData OpData,
	info ProcessInfo,
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
) error {

	var err error

	if origObj.GetStatus() == types.StatusInternalSync {
		if !oplog.IsNewer {
			err = origObj.UpdateCreateInfo(oplog, opData, info)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// it's supposed to be already synced. should not be here.
	return pm.saveNewObjectWithOplog(origObj, oplog, true, postprocessCreateObject)
}

func (pm *BaseProtocolManager) syncCreateObjectDiffLog(
	oplog *BaseOplog, origObj Object,
	info ProcessInfo,
) error {

	oplog.IsSync = true
	return nil
}

/**********
 * Set Newest CreateObjectLog
 **********/

func (pm *BaseProtocolManager) SetNewestCreateObjectLog(
	oplog *BaseOplog, obj Object,
) (types.Bool, error) {

	objID := oplog.ObjID
	obj.SetID(objID)

	err := obj.GetByID(false)
	if err != nil {
		// possibly already deleted
		return false, nil
	}

	updateLogID := obj.GetUpdateLogID()
	if updateLogID != nil {
		return false, nil
	}

	return types.Bool(reflect.DeepEqual(oplog.ID, obj.GetLogID())), nil
}

/**********
 * Handle Failed CreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandleFailedCreateObjectLog(
	oplog *BaseOplog, obj Object,
	postprocessFailedCreateObject func(obj Object, oplog *BaseOplog) error,
) error {

	objID := oplog.ObjID
	obj.SetID(objID)

	// lock-obj
	err := obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	err = obj.GetByID(true)
	if err != nil {
		// already deleted
		return nil
	}

	// check validity
	objLogID := obj.GetLogID()
	if obj.GetUpdateLogID() != nil || !reflect.DeepEqual(objLogID, oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(obj.GetUpdateTS()) {
		return nil
	}

	// handle fail
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	obj.SetUpdateTS(ts)
	obj.SetLogID(nil)
	obj.SetStatus(types.StatusFailed)

	err = obj.Save(true)
	if err != nil {
		return err
	}

	return postprocessFailedCreateObject(obj, oplog)
}

/**********
 * save New Object with Oplog
 **********/

func (pm *BaseProtocolManager) saveNewObjectWithOplog(
	origObj Object,
	oplog *BaseOplog, isLocked bool,
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
) error {

	origStatus := origObj.GetStatus()
	status := oplog.ToStatus()

	if origStatus >= status && !(origStatus == types.StatusFailed && status == types.StatusAlive) {
		return nil
	}

	origObj.SetStatus(status)
	origObj.SetUpdateTS(oplog.UpdateTS)
	err := origObj.Save(isLocked)
	if err == pttdb.ErrInvalidUpdateTS {
		return nil
	}
	if err != nil {
		return err
	}

	if origStatus >= types.StatusAlive && origStatus != types.StatusFailed {
		// orig-status is already alive
		return nil
	}

	if status != types.StatusAlive {
		return nil
	}

	return postprocessCreateObject(origObj, oplog)
}
