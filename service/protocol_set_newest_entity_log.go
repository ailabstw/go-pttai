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

/*
SetNewestEntityLog sets the newest entity log. Because of the nature of entity log, we just need to compare whether the oplog.ID is the same as entity.LogID
*/
func (pm *BaseProtocolManager) SetNewestEntityLog(
	oplog *BaseOplog,
) (types.Bool, error) {

	entity := pm.Entity()

	return !types.Bool(reflect.DeepEqual(oplog.ID, entity.GetLogID())), nil
}

/*
SetNewestDeleteEntityLog sets the newest delete entity log. Utilizing SetNewestEntityLog as the underlying mechanism.
*/
func (pm *BaseProtocolManager) SetNewestDeleteEntityLog(
	oplog *BaseOplog,
) (types.Bool, error) {
	return pm.SetNewestEntityLog(oplog)
}
