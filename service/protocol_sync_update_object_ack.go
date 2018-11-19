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

func (pm *BaseProtocolManager) HandleSyncUpdateObjectAck(
	obj Object,
	peer *PttPeer,

	origObj Object,

	setLogDB func(oplog *BaseOplog),
	updateSyncInfo func(toSyncInfo SyncInfo, fromObj Object, oplog *BaseOplog) error,

	postupdate func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,

) error {

	// oplog
	objID := obj.GetID()
	logID := obj.GetUpdateLogID()

	oplog := &BaseOplog{ID: logID}
	setLogDB(oplog)

	log.Debug("HandleSyncUpdateObjectAck: start", "logID", logID, "objID", objID)

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
	log.Debug("HandleSyncUpdateObjectAck: after GetByID", "logID", logID, "objID", objID, "e", err)
	if err != nil {
		return err
	}

	// validate
	if oplog.IsSync { // already synced
		log.Debug("HandleSyncUpdateObjectAck: already synced", "logID", logID, "objID", objID)
		return nil
	}

	log.Debug("HandleSyncUpdateObjectAck: to check SyncInfo", "logID", logID, "objID", objID)

	syncInfo := origObj.GetSyncInfo()
	if syncInfo == nil || !reflect.DeepEqual(syncInfo.GetLogID(), logID) {
		log.Debug("HandleSyncUpdateObjectAck: syncInfo: newerOplog", "logID", logID, "objID", objID, "syncInfo", syncInfo)
		return ErrNewerOplog
	}

	if syncInfo.GetIsGood() {
		log.Debug("HandleSyncUpdateObjectAck: syncInfo is already good", "logID", logID, "objID", objID)
		return nil
	}

	if updateSyncInfo != nil {
		err = updateSyncInfo(syncInfo, obj, oplog)
		log.Debug("HandleSyncUpdateObjectAck: after updateSyncInfo", "logID", logID, "objID", objID, "e", err)
		if err != nil {
			return err
		}
	}
	syncInfo.SetIsGood(true)

	err = pm.handleUpdateObjectSameLog(origObj, syncInfo, oplog, postupdate)
	log.Debug("HandleSyncUpdateObjectAck: after handleUpdateObjectSameLog", "logID", logID, "objID", objID, "e", err)
	if err != nil {
		return err
	}

	log.Debug("HandleSyncUpdateObjectAck: to syncUpdateAckSaveOplog", "logID", logID, "objID", objID)
	err = pm.syncUpdateAckSaveOplog(oplog, origObj, broadcastLog, postupdate)
	log.Debug("HandleSyncUpdateObjectAck: after syncUpdateAckSaveOplog", "obj.SyncInfo", origObj.GetSyncInfo())

	return err
}

func (pm *BaseProtocolManager) syncUpdateAckSaveOplog(
	oplog *BaseOplog,
	obj Object,

	broadcastLog func(oplog *BaseOplog) error,
	postupdate func(obj Object, oplog *BaseOplog) error,
) error {
	// oplog-save
	if obj.GetIsAllGood() {
		pm.SetOplogIsSync(oplog, true, broadcastLog)
	}

	err := oplog.Save(true)
	if err != nil {
		return err
	}

	syncInfo := obj.GetSyncInfo()

	log.Debug("syncUpdateAckSaveOplog: to check", "oplog", oplog.ID, "objID", obj.GetID(), "syncInfo", syncInfo, "oplog.Status", oplog.ToStatus())

	if syncInfo == nil {
		return nil
	}

	if oplog.ToStatus() < types.StatusAlive {
		return nil
	}

	log.Debug("syncUpdateAckSaveOplog: to handleUpdateObjectSameLog", "logID", oplog.ID, "objID", obj.GetID())

	err = pm.handleUpdateObjectSameLog(obj, syncInfo, oplog, postupdate)
	log.Debug("syncUpdateAckSaveOplog: after handleUpdateObjectSameLog", "obj.SyncInfo", obj.GetSyncInfo())

	return err
}
