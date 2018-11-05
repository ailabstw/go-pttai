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

func (pm *BaseProtocolManager) CreateObject(
	data interface{},
	createOp OpType,

	newObj func(data interface{}) (Object, OpData, error),
	newOplogWithTS func(objID *types.PttID, ts types.Timestamp, op OpType, opData OpData) (Oplog, error),
	increate func(obj Object, oplog *BaseOplog, opData OpData) error,
	broadcastLog func(oplog *BaseOplog) error,
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
) (Object, error) {

	myID := pm.Ptt().GetMyEntity().GetID()
	entity := pm.Entity()

	//validate
	if entity.GetStatus() != types.StatusAlive {
		return nil, types.ErrInvalidStatus
	}

	if entity.GetEntityType() == EntityTypePersonal && !reflect.DeepEqual(entity.GetCreatorID(), myID) {
		return nil, types.ErrInvalidID
	}

	// new-obj
	obj, opData, err := newObj(data)
	if err != nil {
		log.Warn("CreateObject: unable to newObj", "e", err)
		return nil, err
	}

	// oplog
	theOplog, err := newOplogWithTS(obj.GetID(), obj.GetUpdateTS(), createOp, opData)
	if err != nil {
		log.Warn("CreateObject: unable to newOplogWithTS", "e", err)
		return nil, err
	}
	oplog := theOplog.GetBaseOplog()

	if increate != nil {
		err = increate(obj, oplog, opData)
		if err != nil {
			log.Warn("CreateObject: unable to increate", "e", err)
			return nil, err
		}
	}

	err = pm.SignOplog(oplog)
	if err != nil {
		log.Warn("CreateObject: unable to sign", "e", err)
		return nil, err
	}

	err = pm.saveNewObjectWithOplog(obj, oplog, true, postprocessCreateObject)
	if err != nil {
		log.Warn("CreateObject: unable to saveNewObjectWithOplog", "e", err)
		return nil, err
	}

	// oplog-save
	err = oplog.Save(false)
	if err != nil {
		log.Warn("CreateObject: unable to save oplog", "e", err)
		return nil, err
	}

	broadcastLog(oplog)

	return obj, nil
}
