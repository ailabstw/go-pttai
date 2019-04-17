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
SetNewestBoardOplogUpdateArticle set newestLogID in UpdateArticle
    1. get article.
    2. if Status > failed: return LogID
    3. return UpdateLogID

*/
func (pm *BaseProtocolManager) SetNewestUpdateObjectLog(
	oplog *BaseOplog,
	obj Object,
) (types.Bool, error) {

	objID := oplog.ObjID
	obj.SetID(objID)

	// 1. lock
	err := obj.Lock()
	if err != nil {
		return false, err
	}
	defer obj.Unlock()

	// 2. get data
	err = obj.GetByID(true)
	if err != nil {
		// possibly already deleted
		return true, nil
	}

	// 3. cmp
	if obj.GetStatus() >= types.StatusDeleted {
		return true, nil
	}

	return !types.Bool(reflect.DeepEqual(oplog.ID, obj.GetUpdateLogID())), nil
}
