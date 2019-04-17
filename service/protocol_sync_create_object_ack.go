// Copyright 2019 The go-pttai Authors
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
)

type SyncObjectAck struct {
	Objs []Object `json:"o"`
}

func (pm *BaseProtocolManager) SyncObjectAck(objs []Object, syncAckMsg OpType, peer *PttPeer) error {
	if len(objs) == 0 {
		return nil
	}

	pObjs := objs
	var eachObjs []Object
	lenEachObjs := 0
	var data *SyncObjectAck
	for len(pObjs) > 0 {
		lenEachObjs = MaxSyncObjectAck
		if lenEachObjs > len(pObjs) {
			lenEachObjs = len(pObjs)
		}

		eachObjs, pObjs = pObjs[:lenEachObjs], pObjs[lenEachObjs:]

		data = &SyncObjectAck{
			Objs: eachObjs,
		}

		err := pm.SendDataToPeer(syncAckMsg, data, peer)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
HandleSyncCreateObjectAck

We don't need to have updaateCreateObject as member-function of obj because we copy only the public members.
*/
func (pm *BaseProtocolManager) HandleSyncCreateObjectAck(
	obj Object,
	peer *PttPeer,

	origObj Object,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	updateCreateObject func(toObj Object, fromObj Object) error,
	postcreate func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
) error {

	// oplog
	objID := obj.GetID()
	logID := obj.GetLogID()

	oplog := &BaseOplog{ID: logID}
	setLogDB(oplog)

	err := oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	// the temporal-oplog may be already deleted.
	err = oplog.Get(logID, true)
	if err != nil {
		return nil
	}

	// orig-obj
	err = obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	origObj.SetID(objID)
	err = origObj.GetByID(true)
	if err != nil {
		return err
	}

	// validate
	if oplog.IsSync { // already synced
		return nil
	}

	if origObj.GetIsGood() {
		if origObj.GetIsAllGood() {
			return pm.syncCreateAckSaveOplog(
				oplog,
				origObj,

				merkle,

				broadcastLog,
				postcreate,
			)
		}
		return nil
	}

	if origObj.GetUpdateLogID() == nil && reflect.DeepEqual(origObj.GetLogID(), oplog.ID) {
		// still in sync-block
		if updateCreateObject != nil {
			err = updateCreateObject(origObj, obj)
			if err != nil {
				return err
			}
		}
		origObj.SetIsGood(true)
		isAllGood := origObj.CheckIsAllGood()
		if !isAllGood {
			return origObj.Save(true)
		}

		// The oplog may be synced after saveNewObjectWithOplog.
		err = pm.saveNewObjectWithOplog(origObj, oplog, true, false, postcreate)
		if err != nil {
			return err
		}

	} else {
		oplog.IsSync = true
	}

	return pm.syncCreateAckSaveOplog(
		oplog,
		origObj,

		merkle,

		broadcastLog,
		postcreate,
	)

}

func (pm *BaseProtocolManager) syncCreateAckSaveOplog(
	oplog *BaseOplog,
	obj Object,

	merkle *Merkle,

	broadcastLog func(oplog *BaseOplog) error,
	postcreate func(obj Object, oplog *BaseOplog) error,
) error {
	// oplog-save
	if oplog.IsSync {
		pm.SetOplogIsSync(oplog, true, broadcastLog)
	}
	err := oplog.Save(true, merkle)
	if err != nil {
		return err
	}

	if !oplog.IsSync {
		return nil
	}

	if obj.GetStatus() == types.StatusAlive {
		return nil
	}

	if oplog.ToStatus() < types.StatusAlive {
		return nil
	}

	return pm.saveNewObjectWithOplog(obj, oplog, true, false, postcreate)
}
