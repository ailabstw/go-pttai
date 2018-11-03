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

func (pm *BaseProtocolManager) HandleSyncCreateObjectAck(
	obj Object,
	peer *PttPeer,
	setLogDB func(oplog *BaseOplog),
	origObj Object,
	postprocessCreateObject func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
) error {

	// oplog
	oplog := &BaseOplog{ID: obj.GetID()}
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

	if oplog.IsSync { // already synced
		return nil
	}

	// orig-obj
	err = obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	err = origObj.GetByID(true)
	if err != nil {
		return err
	}

	if origObj.GetUpdateLogID() == nil && reflect.DeepEqual(origObj.GetLogID(), oplog.ID) {
		if origObj.GetStatus() == types.StatusInternalSync {
			origObj.UpdateCreateObject(obj)
		}
		err = pm.saveNewObjectWithOplog(origObj, oplog, true, postprocessCreateObject)
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

	// obj becomes alive after set-oplog
	if origStatus >= types.StatusAlive && origStatus != types.StatusFailed {
		return nil
	}

	if oplog.ToStatus() < types.StatusAlive {
		return nil
	}

	err = pm.saveNewObjectWithOplog(origObj, oplog, true, postprocessCreateObject)
	if err != nil {
		return err
	}

	return nil
}
