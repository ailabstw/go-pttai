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
	"github.com/ailabstw/go-pttai/common/types"

	"reflect"
)

/*
SetNewestPersonLog sets the newest person log. Because of the nature of person log, we just need to compare whether the oplog.ID is the same as person.LogID.
*/
func (pm *BaseProtocolManager) SetNewestPersonLog(
	oplog *BaseOplog,
	person Object,
) (types.Bool, error) {

	objID := oplog.ObjID
	person.SetID(objID)

	err := person.GetByID(false)
	if err != nil {
		// possibly already deleted
		return true, nil
	}

	return !types.Bool(reflect.DeepEqual(oplog.ID, person.GetLogID())), nil
}

/*
SetNewestDeletePersonLog sets the newest delete person log. Utilizing SetNewestPersonLog as the underlying mechanism.
*/
func (pm *BaseProtocolManager) SetNewestDeletePersonLog(
	oplog *BaseOplog,
	person Object,
) (types.Bool, error) {

	return pm.SetNewestPersonLog(oplog, person)
}

/*
SetNewestTransferPersonLog sets the newest transfer person log. Utilizing SetNewestPersonLog as the underlying mechanism.
*/
func (pm *BaseProtocolManager) SetNewestTransferPersonLog(
	oplog *BaseOplog,
	obj Object,
) (types.Bool, error) {
	return pm.SetNewestPersonLog(oplog, obj)
}
