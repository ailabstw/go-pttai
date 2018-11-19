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
)

func (pm *BaseProtocolManager) CreatePerson(
	id *types.PttID,
	createOp OpType,
	isForce bool,

	newPerson func(id *types.PttID) (Object, OpData, error),
	newOplogWithTS func(objID *types.PttID, ts types.Timestamp, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
	postcreatePerson func(obj Object, oplog *BaseOplog) error,

) (Object, *BaseOplog, error) {

	// 2. new person
	person, opData, err := newPerson(id)
	if err != nil {
		return nil, nil, err
	}

	// 3. oplog
	theOplog, err := newOplogWithTS(id, person.GetUpdateTS(), createOp, opData)
	if err != nil {
		return nil, nil, err
	}
	oplog := theOplog.GetBaseOplog()

	// 4.1. set is good
	person.SetIsGood(true)
	person.SetIsAllGood(true)

	// 5. sign oplog
	masterLogID := pm.GetNewestMasterLogID()
	if masterLogID == nil {
		masterLogID = oplog.ID
	}

	myEntity := pm.Ptt().GetMyEntity()
	err = myEntity.Sign(oplog)
	if err != nil {
		return nil, nil, err
	}
	err = myEntity.MasterSign(oplog)
	if err != nil {
		return nil, nil, err
	}
	oplog.SetMasterLogID(masterLogID, 1)
	oplog.Hash, err = oplog.SignsHash()
	if err != nil {
		return nil, nil, err
	}

	// 6. save object
	err = pm.saveNewObjectWithOplog(person, oplog, true, false, postcreatePerson)
	if err != nil {
		return nil, nil, err
	}

	// 7. oplog-save
	err = oplog.Save(false)
	if err != nil {
		return nil, nil, err
	}

	broadcastLog(oplog)

	return person, oplog, nil
}
