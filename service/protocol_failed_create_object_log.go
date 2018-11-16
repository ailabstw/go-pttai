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
 * Handle Failed CreateObjectLog
 **********/

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

	// check validity
	objLogID := obj.GetLogID()
	if obj.GetUpdateLogID() != nil || !reflect.DeepEqual(objLogID, oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(obj.GetUpdateTS()) {
		return nil
	}

	// handle fail
	err = prefailed(obj, oplog)
	if err != nil {
		return err
	}

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
