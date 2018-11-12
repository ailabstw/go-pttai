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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/syndtr/goleveldb/leveldb"
)

type ProcessPersonInfo struct {
	CreateInfo map[types.PttID]*BaseOplog
	DeleteInfo map[types.PttID]*BaseOplog
}

func NewProcessPersonInfo() *ProcessPersonInfo {
	return &ProcessPersonInfo{
		CreateInfo: make(map[types.PttID]*BaseOplog),
		DeleteInfo: make(map[types.PttID]*BaseOplog),
	}
}

/**********
 * Handle CreatePersonLog
 **********/

func (pm *BaseProtocolManager) HandleCreatePersonLog(
	oplog *BaseOplog,
	person Object,
	opData OpData,

	postcreatePerson func(obj Object, oplog *BaseOplog) error,
) ([]*BaseOplog, error) {

	personID := oplog.ObjID

	// 1. lock person
	person.SetID(personID)
	err := person.Lock()
	if err != nil {
		return nil, err
	}
	defer person.Unlock()

	// 2. get person (should never delete once stored)
	err = person.GetByID(true)
	if err != leveldb.ErrNotFound {
		return nil, err
	}

	// 3. save object
	err = pm.saveNewObjectWithOplog(person, oplog, true, true, postcreatePerson)
	if err != nil {
		return nil, err
	}

	// 4. set oplog is-sync
	oplog.IsSync = true

	return nil, nil
}

/**********
 * Handle PendingCreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingCreatePersonLog(
	oplog *BaseOplog,
	person Object,

	opData OpData,

) ([]*BaseOplog, error) {
	return nil, types.ErrNotImplemented
}

/**********
 * Handle Failed CreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandleFailedCreatePersonLog(
	oplog *BaseOplog,
	person Object,

	postfailedCreatePerson func(obj Object, oplog *BaseOplog) error,
) error {
	return types.ErrNotImplemented
}
