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

	origObj Object,
	deleteOp OpType,
	opData OpData,

	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
	postdelete func(id *types.PttID, oplog *BaseOplog, origObj Object, opData OpData) error,
) error {

	myEntity := pm.Ptt().GetMyEntity()
	myID := myEntity.GetID()

	// 1. lock object
	origObj.SetID(id)
	err := origObj.Lock()
	if err != nil {
		return err
	}
	defer origObj.Unlock()

	// 2. get obj
	err = origObj.GetByID(true)
	if err != nil {
		return err
	}

	// 3. check validity
	origStatus := origObj.GetStatus()
	if origStatus > types.StatusFailed {
		return nil
	}

	creatorID := origObj.GetCreatorID()
	if !reflect.DeepEqual(myID, creatorID) && !pm.IsMaster(myID) {
		return types.ErrInvalidID
	}

	// 4. oplog
	theOplog, err := newOplog(id, deleteOp, opData)
	if err != nil {
		return err
	}
	oplog := theOplog.GetBaseOplog()

	origLogID := origObj.GetLogID()
	if origStatus <= types.StatusAlive {
		oplog.SetPreLogID(origLogID)
	}

	err = pm.SignOplog(oplog)
	if err != nil {
		return err
	}

	// 4.1 get orig-block-info
	origBlockInfo := origObj.GetBlockInfo()

	// 5. update obj
	oplogStatus := oplog.ToStatus()
	if oplogStatus == types.StatusAlive {
		SetDeleteObjectWithOplog(origObj, oplog)
	} else {
		origObj.SetPendingDeleteSyncInfo(oplog)
	}

	err = origObj.Save(true)
	if err != nil {
		log.Error("DeleteObject: unable to update obj", "e", err, "origObj", origObj)
		return err
	}

	// 5.1 remove block
	if origStatus == types.StatusAlive {
		err = origObj.RemoveBlock(origBlockInfo, nil, true)
		if err != nil {
			return err
		}
	}

	// 6. oplog
	err = oplog.Save(true)
	if err != nil {
		return err
	}

	broadcastLog(oplog)

	// 6.2 postdelete
	if origStatus != types.StatusAlive {
		return nil
	}

	if postdelete != nil {
		postdelete(id, oplog, origObj, opData)
	}

	return nil
}
