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
 * Set Newest CreateObjectLog
 **********/

func (pm *BaseProtocolManager) SetNewestCreateObjectLog(
	oplog *BaseOplog,
	obj Object,
) (types.Bool, error) {

	objID := oplog.ObjID
	obj.SetID(objID)

	// 1. lock
	err := obj.RLock()
	if err != nil {
		return false, err
	}
	defer obj.RUnlock()

	// 2. get data
	err = obj.GetByID(true)
	if err != nil {
		// possibly already deleted
		log.Debug("SetNewestCreateObjectLog: unable to get obj", "oplog", oplog.ID)
		return true, nil
	}

	// 3. cmp

	updateLogID := obj.GetUpdateLogID()
	if updateLogID != nil {
		log.Debug("SetNewestCreateObjectLog: with updateLogID", "oplog", oplog.ID, "updateLogID", updateLogID)
		return true, nil
	}

	if !reflect.DeepEqual(oplog.ID, obj.GetLogID()) {
		log.Debug("SetNewestCreateObjectLog: logID not same", "oplog", oplog.ID, "obj.LogID", obj.GetLogID())
		return true, nil
	}

	log.Debug("SetNewestCreateObjectLog: no newerID", "oplog", oplog.ID)

	return false, nil
}
