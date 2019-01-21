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

func (pm *BaseProtocolManager) DeleteObject(
	id *types.PttID,
	deleteOp OpType,

	origObj Object,
	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	indelete func(origObj Object, opData OpData, oplog *BaseOplog) (*BlockInfo, error),
	setPendingDeleteSyncInfo func(origObj Object, status types.Status, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
	postdelete func(id *types.PttID, oplog *BaseOplog, opData OpData, origObj Object, blockInfo *BlockInfo) error,
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
	log.Debug("DeleteObject: after GetByID", "e", err)
	if err != nil {
		return err
	}

	// 3. check validity
	origStatus := origObj.GetStatus()
	if origStatus >= types.StatusDeleted {
		return nil
	}

	creatorID := origObj.GetCreatorID()
	if !reflect.DeepEqual(myID, creatorID) && !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	// 4. oplog
	log.Debug("DeleteObject: to newOplog")
	theOplog, err := newOplog(id, deleteOp, opData)
	if err != nil {
		return err
	}
	oplog := theOplog.GetBaseOplog()

	origLogID := origObj.GetLogID()
	if origStatus == types.StatusAlive {
		oplog.SetPreLogID(origLogID)
	}

	log.Debug("DeleteObject: to SignOplog", "oplog", oplog, "origLogID", origLogID, "origStatus", origStatus)
	err = pm.SignOplog(oplog)
	log.Debug("DeleteObject: after SignOplog", "e", err)
	if err != nil {
		return err
	}

	// 5. core
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), types.StatusInternalDeleted, types.StatusPendingDeleted, types.StatusDeleted)

	log.Debug("DeleteObject: to core", "oplogStatus", oplogStatus, "statusDelete", types.StatusDeleted)

	if oplogStatus >= types.StatusDeleted {
		err = pm.handleDeleteObjectLogCore(
			oplog,
			nil,

			origObj,
			opData,

			merkle,

			setLogDB,
			nil,
			postdelete,
			nil,
		)
	} else {
		err = pm.handlePendingDeleteObjectLogCore(
			oplog,
			nil,

			origObj,
			opData,

			merkle,

			setLogDB,

			nil,
			setPendingDeleteSyncInfo,
			nil,
		)
	}
	if err != nil {
		return err
	}

	// 6. oplog save
	err = oplog.Save(false, merkle)
	if err != nil {
		return err
	}

	log.Debug("DeleteObject: to broadcastLog", "oplog", oplog)

	broadcastLog(oplog)

	log.Debug("DeleteObject: after broadcastLog")

	return nil
}
