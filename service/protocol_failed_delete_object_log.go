// Copyright 2019 The go-pttai Authors
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

func (pm *BaseProtocolManager) HandleFailedValidDeleteObjectLog(
	oplog *BaseOplog,
	obj Object,

	info ProcessInfo,
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,

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

	// 3. check validity
	objLogID := obj.GetLogID()
	if obj.GetUpdateLogID() != nil || !reflect.DeepEqual(objLogID, oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(obj.GetUpdateTS()) {
		return nil
	}

	// 6. obj-save
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	SetFailedObjectWithOplog(obj, oplog, ts)

	err = obj.Save(true)
	if err != nil {
		return err
	}

	// 7. update delete info
	updateDeleteInfo(obj, oplog, info)

	return nil
}
