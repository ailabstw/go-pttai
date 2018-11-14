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
	"github.com/syndtr/goleveldb/leveldb"
)

type SyncPersonInfo struct {
	*BaseSyncInfo `json:"b"`
	TransferToID  *types.PttID `json:"t"`
}

func NewEmptySyncPersonInfo() *SyncPersonInfo {
	return &SyncPersonInfo{BaseSyncInfo: &BaseSyncInfo{}}
}

func (pm *BaseProtocolManager) TransferPerson(
	fromID *types.PttID,
	toID *types.PttID,

	createOp OpType,
	origPerson Object,
	opData *PersonOpTransferPerson,

	setLogDB func(oplog *BaseOplog),
	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	signOplog func(oplog *BaseOplog, fromID *types.PttID, toID *types.PttID) error,
	setTransferPersonWithOplog func(person Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
	posttransfer func(fromID *types.PttID, toID *types.PttID, person Object, oplog *BaseOplog, opData OpData) error,
) error {

	myID := pm.Ptt().GetMyEntity().GetID()
	entity := pm.Entity()

	masterList, _ := pm.GetMasterListFromCache(false)
	for i, eachMaster := range masterList {
		log.Debug("TransferPerson: start", "i", i, "master", eachMaster)
	}

	log.Debug("TransferPerson: start", "myID", myID, "fromID", fromID, "toID", toID, "entity", entity.GetID())

	// validate
	if !pm.IsMaster(myID, false) && !reflect.DeepEqual(myID, fromID) {
		return types.ErrInvalidID
	}

	// lock orig-person
	origPerson.SetID(fromID)
	err := origPerson.Lock()
	if err != nil {
		return err
	}
	defer origPerson.Unlock()

	// get orig-person
	err = origPerson.GetByID(true)
	if err != nil {
		return err
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus >= types.StatusDeleted {
		return types.ErrAlreadyDeleted
	}

	// 4. oplog
	theOplog, err := newOplog(fromID, createOp, opData)
	if err != nil {
		return err
	}
	oplog := theOplog.GetBaseOplog()
	oplog.PreLogID = origPerson.GetLogID()

	err = signOplog(oplog, fromID, toID)
	if err != nil {
		return err
	}

	// 5. update obj
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), types.StatusInternalTransfer, types.StatusPendingTransfer, types.StatusTransferred)

	log.Debug("TransferPerson: to TransferPersonLogCore", "oplogStatus", oplogStatus)

	if oplogStatus == types.StatusTransferred {
		err = pm.handleTransferPersonLogCore(oplog, origPerson, opData, setLogDB, posttransfer)
	} else {
		err = pm.handlePendingTransferPersonLogCore(oplog, origPerson, opData, setLogDB)
	}
	if err != nil {
		return err
	}

	// 6. oplog
	err = oplog.Save(true)
	if err != nil {
		return err
	}

	broadcastLog(oplog)

	return nil
}

func SetPendingTransferPersonSyncInfo(person Object, toID *types.PttID, oplog *BaseOplog) error {

	syncInfo := NewEmptySyncPersonInfo()
	syncInfo.InitWithOplog(oplog)
	syncInfo.Status = types.StatusPendingTransfer
	syncInfo.TransferToID = toID

	person.SetSyncInfo(syncInfo)

	return nil
}

func (pm *BaseProtocolManager) posttransferPerson(
	toID *types.PttID,
	oplog *BaseOplog,
	origPerson Object,

	newPerson func(id *types.PttID) (Object, OpData, error),
	postcreatePerson func(obj Object, oplog *BaseOplog) error,

) (Object, error) {

	log.Debug("posttransferPerson: start", "toID", toID, "entity", pm.Entity().GetID())

	// lock orig-person
	origPerson.SetID(toID)
	err := origPerson.Lock()
	if err != nil {
		return nil, err
	}
	defer origPerson.Unlock()

	// get orig-person
	err = origPerson.GetByID(true)
	switch err {
	case nil:
		err = pm.posttransferUpdatePerson(origPerson, oplog, postcreatePerson)
	case leveldb.ErrNotFound:
		origPerson, err = pm.posttransferCreatePerson(toID, oplog, newPerson, postcreatePerson)
	}
	if err != nil {
		return nil, err
	}
	return origPerson, nil
}

func (pm *BaseProtocolManager) posttransferUpdatePerson(
	origPerson Object,
	oplog *BaseOplog,
	postcreatePerson func(obj Object, oplog *BaseOplog) error,
) error {

	log.Debug("posttransferUpdatePerson: start", "entity", pm.Entity().GetID(), "person", origPerson.GetID(), "status", origPerson.GetStatus())

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus == types.StatusAlive {
		return nil
	}
	if origStatus == types.StatusTransferred {
		return types.ErrAlreadyDeleted
	}

	err := pm.saveUpdateObjectWithOplog(origPerson, oplog, true)
	if err != nil {
		return err
	}

	log.Debug("posttransferUpdatePerson: to postcreatePerson", "entity", pm.Entity().GetID(), "person", origPerson)

	if postcreatePerson == nil {
		return nil
	}

	return postcreatePerson(origPerson, oplog)
}

func (pm *BaseProtocolManager) posttransferCreatePerson(
	id *types.PttID,
	oplog *BaseOplog,
	newPerson func(id *types.PttID) (Object, OpData, error),
	postcreatePerson func(obj Object, oplog *BaseOplog) error,
) (Object, error) {

	log.Debug("posttransferNewPerson: start", "entity", pm.Entity().GetID(), "id", id)

	// new person
	person, _, err := newPerson(id)
	if err != nil {
		return nil, err
	}

	// save object
	err = pm.saveNewObjectWithOplog(person, oplog, true, true, postcreatePerson)
	if err != nil {
		return nil, err
	}

	log.Debug("posttransferNewPerson: done", "entity", pm.Entity().GetID(), "id", id)

	return person, nil
}
