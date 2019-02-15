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
)

func (pm *BaseProtocolManager) DeleteEntity(
	deleteOp OpType,
	opData OpData,
	internalPendingStatus types.Status,
	pendingStatus types.Status,
	status types.Status,

	merkle *Merkle,

	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),

	setPendingDeleteSyncInfo func(entity Entity, status types.Status, oplog *BaseOplog) error,

	broadcastLog func(oplog *BaseOplog) error,
	postdelete func(opData OpData, isForce bool) error,
) error {

	myEntity := pm.Ptt().GetMyEntity()
	myID := myEntity.GetID()
	entity := pm.Entity()

	// validate
	if entity.GetStatus() > types.StatusFailed {
		return types.ErrInvalidStatus
	}
	/*
		if myEntity.GetStatus() != types.StatusAlive {
			return types.ErrInvalidStatus
		}
	*/

	// 1. lock object
	err := entity.Lock()
	if err != nil {
		return err
	}
	defer entity.Unlock()

	// 3. check validity
	origStatus := entity.GetStatus()
	if origStatus >= types.StatusMigrated {
		return nil
	}

	if !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	// 4. oplog
	entityID := entity.GetID()
	theOplog, err := newOplog(entityID, deleteOp, opData)
	if err != nil {
		return err
	}
	oplog := theOplog.GetBaseOplog()

	origLogID := entity.GetLogID()
	if origStatus == types.StatusAlive {
		oplog.SetPreLogID(origLogID)
	}

	err = pm.SignOplog(oplog)
	if err != nil {
		return err
	}

	// 5. update obj
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), internalPendingStatus, pendingStatus, status)
	if oplogStatus >= types.StatusDeleted {
		SetNewEntityWithOplog(entity, status, oplog)
	} else {
		if !isReplaceOrigSyncInfo(entity.GetSyncInfo(), oplogStatus, oplog.UpdateTS, oplog.ID) {
			return types.ErrAlreadyPendingDelete
		}

		setPendingDeleteSyncInfo(entity, oplogStatus, oplog)
	}

	log.Debug("DeleteEntity: entity to save", "status", entity.GetStatus(), "syncInfo", entity.GetSyncInfo())

	err = entity.Save(true)
	if err != nil {
		return err
	}

	// 6. oplog
	err = oplog.Save(true, merkle)
	if err != nil {
		return err
	}

	broadcastLog(oplog)

	// 6.1. postdelete
	if oplogStatus < types.StatusDeleted {
		return nil
	}

	if postdelete != nil {
		postdelete(opData, false)
	}

	return nil
}
