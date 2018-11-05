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

func (pm *BaseProtocolManager) DeleteEntity(
	deleteOp OpType,
	opData OpData,
	status types.Status,

	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
	postdelete func(opData OpData) error,
) error {

	entity := pm.Entity()

	// 1. lock object
	err := entity.Lock()
	if err != nil {
		return err
	}
	defer entity.Unlock()

	// 3. check validity
	origStatus := entity.GetStatus()
	origStatusClass := types.StatusToStatusClass(origStatus)
	if origStatusClass == types.StatusClassDeleted {
		return nil
	}

	myID := pm.Ptt().GetMyEntity().GetID()
	if !pm.IsMaster(myID) {
		return types.ErrInvalidID
	}

	// 4. oplog
	entityID := entity.GetID()
	theOplog, err := newOplog(entityID, deleteOp, opData)
	if err != nil {
		return err
	}
	oplog := theOplog.GetBaseOplog()

	err = pm.SignOplog(oplog)
	if err != nil {
		return err
	}

	// 5. update obj
	oplogStatus := oplog.ToStatus()
	if oplogStatus == types.StatusAlive {
		EntitySetStatusWithOplog(entity, status, oplog)
	} else {
		entity.SetPendingDeleteSyncInfo(status, oplog)
	}

	err = entity.Save(true)
	if err != nil {
		return err
	}

	// 6. oplog
	err = oplog.Save(true)
	if err != nil {
		return err
	}

	broadcastLog(oplog)

	// 6.1. postdelete
	if oplogStatus != types.StatusAlive {
		return nil
	}

	if postdelete != nil {
		postdelete(opData)
	}

	return nil
}
