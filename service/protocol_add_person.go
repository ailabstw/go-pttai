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
	"github.com/ailabstw/go-pttai/log"
	"github.com/syndtr/goleveldb/leveldb"
)

/*
type CreatePerson struct {
	ID *types.PttID
}
*/

func (pm *BaseProtocolManager) AddPerson(
	id *types.PttID,
	addOp OpType,
	isForce bool,

	origPerson Object, // for update
	opData OpData, // for update

	merkle *Merkle,

	newPerson func(id *types.PttID) (Object, OpData, error),
	newOplogWithTS func(objID *types.PttID, ts types.Timestamp, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
	postcreate func(obj Object, oplog *BaseOplog) error,

	setLogDB func(oplog *BaseOplog), // for update
	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error), // for update
) (Object, *BaseOplog, error) {

	myID := pm.Ptt().GetMyEntity().GetID()
	entity := pm.Entity()

	// 1. validate
	if !isForce {
		if entity.GetStatus() != types.StatusAlive {
			return nil, nil, types.ErrInvalidStatus
		}

		if !pm.IsMaster(myID, false) {
			return nil, nil, types.ErrInvalidID
		}
	}

	// 1.5. lock orig-person
	origPerson.SetID(id)
	err := origPerson.Lock()
	if err != nil {
		return nil, nil, err
	}
	defer origPerson.Unlock()

	// get orig-person
	err = origPerson.GetByID(true)
	log.Debug("AddPerson: after get origPerson", "entity", entity.GetID(), "creatorID", entity.GetCreatorID(), "id", id, "e", err, "origPerson", origPerson)
	if err == nil {
		return pm.UpdatePerson(
			id,
			addOp,
			isForce,

			origPerson,
			opData,

			merkle,

			setLogDB,
			newOplog,
			broadcastLog,
			postcreate,
		)
	}
	if err != leveldb.ErrNotFound {
		return nil, nil, err
	}

	return pm.CreatePerson(
		id,
		addOp,
		isForce,

		merkle,

		newPerson,
		newOplogWithTS,
		broadcastLog,
		postcreate,
	)
}
