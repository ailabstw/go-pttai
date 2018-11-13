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

type SyncCreateObjectAck struct {
	Objs []Object `json:"o"`
}

func (pm *BaseProtocolManager) SyncCreateObjectAck(objs []Object, syncAckMsg OpType, peer *PttPeer) error {
	if len(objs) == 0 {
		return nil
	}

	data := &SyncCreateObjectAck{
		Objs: objs,
	}

	err := pm.SendDataToPeer(syncAckMsg, data, peer)
	if err != nil {
		return err
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

	setLogDB func(oplog *BaseOplog),
	updateCreateObject func(toObj Object, fromObj Object) error,
	postcreateObject func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
) error {

	// oplog
	objID := obj.GetID()

	oplog := &BaseOplog{ID: objID}
	setLogDB(oplog)

	err := oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	// the temporal-oplog may be deleted.
	err = oplog.Get(obj.GetLogID(), true)
	if err != nil {
		return nil
	}

	log.Debug("HandleSyncCreateObjectAck: to check IsSync", "oplog", oplog)
	if oplog.IsSync { // already synced
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

	log.Debug("HandleSyncCreateObjectAck: to check oplog", "objLogID", origObj.GetLogID(), "oplogID", oplog.ID, "status", origObj.GetStatus())
	if origObj.GetUpdateLogID() == nil && reflect.DeepEqual(origObj.GetLogID(), oplog.ID) {
		if origObj.GetStatus() == types.StatusInternalSync {
			err = updateCreateObject(origObj, obj)
			if err != nil {
				return err
			}
		}
		err = pm.saveNewObjectWithOplog(origObj, oplog, true, false, postcreateObject)
		if err != nil {
			return err
		}
	}

	// oplog-save
	origStatus := origObj.GetStatus()

	pm.SetOplogIsSync(oplog, true, broadcastLog)
	err = oplog.Save(true)
	if err != nil {
		return err
	}

	log.Debug("HandleSyncCreateObjectAck: after SetOplogIsSync", "origStatus", origStatus, "oplogStatus", oplog.ToStatus())

	// obj becomes alive after set-oplog
	if origStatus >= types.StatusAlive && origStatus != types.StatusFailed {
		return nil
	}

	if oplog.ToStatus() < types.StatusAlive {
		return nil
	}

	err = pm.saveNewObjectWithOplog(origObj, oplog, true, false, postcreateObject)
	if err != nil {
		return err
	}

	return nil
}
