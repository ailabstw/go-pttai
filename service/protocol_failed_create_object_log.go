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

/*
HandleFailedCreateObjectLog handles failed create-object log.

	1. lock-obj.
	2. get obj, and return if unable to get the obj (already deleted)
	3. check validity.
	4. prefailed
	5. if not my object: remove blocks and the object.
	6. if my object: only set the status as failed.
*/
func (pm *BaseProtocolManager) HandleFailedCreateObjectLog(
	oplog *BaseOplog,
	obj Object,

	prefailed func(obj Object, oplog *BaseOplog) error,
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

	// 4. prefailed
	if prefailed != nil {
		err = prefailed(obj, oplog)
		if err != nil {
			return err
		}
	}

	// 5. not my object.
	myID := pm.Ptt().GetMyEntity().GetID()
	if !reflect.DeepEqual(myID, obj.GetCreatorID()) {
		blockInfo := obj.GetBlockInfo()
		err = pm.removeBlockAndMediaInfoByBlockInfo(blockInfo, nil, oplog, true, nil)
		if err != nil {
			return err
		}
		obj.Delete(true)

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

	return nil
}

/*
HandleFailedCreateObjectLog handles failed create-object log.

	1. lock-obj.
	2. get obj, and return if unable to get the obj (already deleted)
	3. check validity.
	4. prefailed
	5. if not my object: remove blocks and the object.
	6. if my object: only set the status as failed.
*/
func (pm *BaseProtocolManager) HandleFailedValidCreateObjectLog(
	oplog *BaseOplog,
	obj Object,

	prefailed func(obj Object, oplog *BaseOplog) error,
) error {

	return pm.HandleFailedCreateObjectLog(oplog, obj, prefailed)
}
