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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) HandleSyncUpdateObjectAck(
	obj Object,
	peer *PttPeer,

	origObj Object,

	merkle *Merkle,

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
	if err != nil {
		return err
	}

	// validate
	if oplog.IsSync { // already synced
		return nil
	}

	// obj already all synced
	origUpdateLogID := origObj.GetUpdateLogID()
	if reflect.DeepEqual(origUpdateLogID, logID) {
		return ErrNewerOplog
	}

	syncInfo := origObj.GetSyncInfo()
	if syncInfo == nil || !reflect.DeepEqual(syncInfo.GetLogID(), logID) {
		return ErrNewerOplog
	}

	if syncInfo.GetIsGood() {
		return nil
	}

	if updateSyncInfo != nil {
		err = updateSyncInfo(syncInfo, obj, oplog)
		if err != nil {
			return err
		}
	}
	syncInfo.SetIsGood(true)

	err = pm.handleUpdateObjectSameLog(origObj, syncInfo, oplog, postupdate)
	if err != nil {
		return err
	}

	err = pm.syncUpdateAckSaveOplog(
		oplog,
		syncInfo,
		origObj,

		merkle,

		broadcastLog,
		postupdate,
	)

	return err
}

func (pm *BaseProtocolManager) syncUpdateAckSaveOplog(
	oplog *BaseOplog,
	syncInfo SyncInfo,
	obj Object,

	merkle *Merkle,

	broadcastLog func(oplog *BaseOplog) error,
	postupdate func(obj Object, oplog *BaseOplog) error,
) error {

	// oplog-save
	if syncInfo == nil {
		return nil
	}

	if syncInfo.GetIsAllGood() {
		pm.SetOplogIsSync(oplog, true, broadcastLog)
	}

	err := oplog.Save(true, merkle)
	if err != nil {
		return err
	}

	if oplog.ToStatus() < types.StatusAlive {
		return nil
	}

	err = pm.handleUpdateObjectSameLog(obj, syncInfo, oplog, postupdate)

	return err
}
