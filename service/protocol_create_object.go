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
	"github.com/ailabstw/go-pttai/pttdb"
)

type CreateData interface{}

func (pm *BaseProtocolManager) CreateObject(
	data CreateData,
	createOp OpType,

	newObj func(data CreateData) (Object, OpData, error),
	newOplogWithTS func(objID *types.PttID, ts types.Timestamp, op OpType, opData OpData) (Oplog, error),
	increate func(obj Object, oplog *BaseOplog, opData OpData) error,
	broadcastLog func(oplog *BaseOplog) error,
	postcreateObject func(obj Object, oplog *BaseOplog) error,
) (Object, error) {

	entity := pm.Entity()

	//validate
	if entity.GetStatus() != types.StatusAlive {
		return nil, types.ErrInvalidStatus
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

	// in-create
	if increate != nil {
		err = increate(obj, oplog, opData)
		if err != nil {
			log.Warn("CreateObject: unable to increate", "e", err)
			return nil, err
		}
	}

	// sign oplog
	err = pm.SignOplog(oplog)
	if err != nil {
		log.Warn("CreateObject: unable to sign", "e", err)
		return nil, err
	}

	// save object
	err = pm.saveNewObjectWithOplog(obj, oplog, false, false, postcreateObject)
	if err != nil {
		log.Warn("CreateObject: unable to saveNewObjectWithOplog", "e", err)
		return nil, err
	}

	// oplog-save
	if oplog.ToStatus() == types.StatusAlive {
		oplog.IsSync = true
	}
	err = oplog.Save(false)
	if err != nil {
		log.Warn("CreateObject: unable to save oplog", "e", err)
		return nil, err
	}

	broadcastLog(oplog)

	return obj, nil
}

/**********
 * save New Object with Oplog
 **********/

func (pm *BaseProtocolManager) saveNewObjectWithOplog(
	origObj Object,
	oplog *BaseOplog,

	isLocked bool,
	isForce bool,

	postcreateObject func(obj Object, oplog *BaseOplog) error,
) error {

	origStatus := origObj.GetStatus()
	status := oplog.ToStatus()

	log.Debug("saveNewObjectWithOplog: start", "origStatus", origStatus, "status", status, "isForce", isForce)

	if !isForce && origStatus >= status && !(origStatus == types.StatusFailed && status == types.StatusAlive) {
		oplog.IsSync = true
		return nil
	}

	origObj.SetLogID(oplog.ID)
	origObj.SetStatus(status)
	origObj.SetUpdateTS(oplog.UpdateTS)
	err := origObj.Save(isLocked)
	log.Debug("saveNewObjectWithOplog: after Save", "e", err, "obj", origObj.GetID())
	if err == pttdb.ErrInvalidUpdateTS {
		return nil
	}
	if err != nil {
		return err
	}

	// set oplog is sync
	oplog.IsSync = true

	// postcreateObject
	if postcreateObject == nil {
		return nil
	}

	if !isForce && origStatus >= types.StatusAlive && origStatus != types.StatusFailed {
		// orig-status is already alive
		return nil
	}

	if status != types.StatusAlive {
		return nil
	}

	return postcreateObject(origObj, oplog)
}
