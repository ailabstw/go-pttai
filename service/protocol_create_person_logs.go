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

/*
HandleCreatePersonLog handles valid create-person log.

	1. lock person
	2. check whether the person exists. (should always be ErrNotFound)
	3. new person.
	4. set is-all-sync
	5. save object with oplog.
*/
func (pm *BaseProtocolManager) HandleCreatePersonLog(
	oplog *BaseOplog,
	person Object,
	opData OpData,

	postcreate func(obj Object, oplog *BaseOplog) error,
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
	if err == nil {
		return nil, ErrNewerOplog
	}
	if err != leveldb.ErrNotFound {
		return nil, err
	}

	// 3. new person
	NewObjectWithOplog(person, oplog)

	// 4. set is good
	person.SetIsGood(true)
	person.SetIsAllGood(true)

	// 5. save object
	err = pm.saveNewObjectWithOplog(person, oplog, true, false, postcreate)
	if err != nil {
		return nil, err
	}

	// 6. set oplog is-sync
	oplog.IsSync = true

	return nil, nil
}

/*
HandlePendingCreateObjectLog handles pending create-person log (should never happen)
*/
func (pm *BaseProtocolManager) HandlePendingCreatePersonLog(
	oplog *BaseOplog,
	person Object,

	opData OpData,

) (types.Bool, []*BaseOplog, error) {
	return false, nil, types.ErrNotImplemented
}

/*
HandleFailedCreatePersonLog handles failed create-person log (should never happen)
*/
func (pm *BaseProtocolManager) HandleFailedCreatePersonLog(
	oplog *BaseOplog,
	person Object,

	prefailed func(obj Object, oplog *BaseOplog) error,
) error {
	return types.ErrNotImplemented
}
