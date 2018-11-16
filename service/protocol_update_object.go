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

type UpdateData interface{}

func (pm *BaseProtocolManager) UpdateObject(
	// id
	id *types.PttID,

	// data
	data UpdateData,
	updateOp OpType,

	// obj
	origObj Object,

	// oplog
	opData OpData,

	setLogDB func(oplog *BaseOplog),

	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),

	inupdate func(obj Object, data UpdateData, oplog *BaseOplog, opData OpData) (SyncInfo, error),

	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),

	broadcastLog func(oplog *BaseOplog) error,
	postupdate func(obj Object, oplog *BaseOplog) error,

) error {

	myEntity := pm.Ptt().GetMyEntity()
	myID := myEntity.GetID()
	entity := pm.Entity()

	// validate
	if entity.GetStatus() != types.StatusAlive {
		return types.ErrInvalidStatus
	}
	if myEntity.GetStatus() != types.StatusAlive {
		return types.ErrInvalidStatus
	}

	// 1. lock object
	origObj.SetID(id)
	err := origObj.Lock()
	if err != nil {
		return err
	}
	defer origObj.Unlock()

	// 2. get obj
	err = origObj.GetByID(true)
	log.Debug("UpdateObject: after GetByID", "e", err)
	if err != nil {
		return err
	}

	// 3. check validity
	origStatus := origObj.GetStatus()
	if origStatus != types.StatusAlive {
		return ErrNotAlive
	}

	creatorID := origObj.GetCreatorID()
	if !reflect.DeepEqual(myID, creatorID) && !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	origSyncInfo := origObj.GetSyncInfo()
	if origSyncInfo != nil {
		return ErrAlreadyPending
	}

	// 4. oplog
	theOplog, err := newOplog(id, updateOp, opData)
	if err != nil {
		return err
	}
	oplog := theOplog.GetBaseOplog()

	origLogID := origObj.GetLogID()
	oplog.SetPreLogID(origLogID)

	// 4.1. inupdate
	if inupdate == nil {
		return ErrInvalidFunc
	}

	syncInfo, err := inupdate(origObj, data, oplog, opData)
	if err != nil {
		return err
	}

	// 4.2. set is good
	syncInfo.SetIsGood(true)
	syncInfo.SetIsAllGood(true)

	// 5. sign oplog
	err = pm.SignOplog(oplog)
	if err != nil {
		return err
	}

	err = oplog.Verify()
	if err != nil {
		log.Warn("UpdateObject: after inupdate: unable to verify oplog")
	}

	// 6. core
	err = pm.handleUpdateObjectCoreCore(oplog, opData, origObj, syncInfo, nil, true, setLogDB, removeMediaInfoByBlockInfo, postupdate, nil)
	log.Debug("UpdateObject: after handleUpdateObjectCoreCore", "oplog", oplog.ID, "obj", origObj.GetID(), "e", err)
	if err != nil {
		return err
	}

	err = oplog.Verify()
	if err != nil {
		log.Warn("UpdateObject: after handleUpdateObjectCoreCore: unable to verify oplog")
	}

	// 6. oplog save
	err = oplog.Save(false)
	if err != nil {
		return err
	}

	log.Debug("UpdateObject: to broadcastLog", "oplog", oplog.ID, "obj", origObj.GetID(), "oplog.Status", oplog.ToStatus())

	broadcastLog(oplog)

	return nil

}
