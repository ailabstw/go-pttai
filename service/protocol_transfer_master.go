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

func (pm *BaseProtocolManager) TransferMaster(id *types.PttID) error {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	// 1. validate
	if !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	if !pm.IsMember(id, false) {
		return types.ErrInvalidID
	}

	// 2. do transfer-person
	origPerson := NewEmptyMaster()
	pm.SetMasterObjDB(origPerson)
	data := &PersonOpTransferPerson{ToID: id}

	err := pm.TransferPerson(
		myID,
		id,

		MasterOpTypeTransferMaster,
		origPerson,
		data,

		pm.MasterMerkle(),

		types.StatusInternalTransfer,
		types.StatusPendingTransfer,
		types.StatusTransferred,

		pm.SetMasterDB,
		pm.NewMasterOplog,
		pm.signTransferMasterOplog,
		pm.setTransferMasterWithOplog,
		pm.broadcastMasterOplogCore,
		pm.posttransferMaster,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) signTransferMasterOplog(oplog *BaseOplog, fromID *types.PttID, toID *types.PttID) error {

	err := pm.SignOplog(oplog)
	log.Debug("signTransferMasterOplog: after SignOplog", "e", err, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	if err != nil {
		return err
	}

	if oplog.MasterLogID == nil {
		return nil
	}

	return pm.checkTransferMasterSign(oplog, fromID, toID)
}

func (pm *BaseProtocolManager) checkTransferMasterSign(oplog *BaseOplog, fromID *types.PttID, toID *types.PttID) error {
	// do nothing if I am also the toID
	mySPM := pm.Ptt().GetMyService().SPM()
	for id, _ := range mySPM.Entities() {
		log.Debug("signTransferMasterOplog: (in-for-loop)", "id", id, "toID", toID)
		if reflect.DeepEqual(id[:], toID[:]) {
			return nil
		}
	}

	log.Debug("signTransferMasterOplog: after for-loop")

	// check master sign
	masterSigns := oplog.MasterSigns
	if !IDInOplogSigns(fromID, masterSigns) || !IDInOplogSigns(toID, masterSigns) {
		oplog.MasterLogID = nil
	}

	return nil

}

func (pm *BaseProtocolManager) setTransferMasterWithOplog(theMaster Object, oplog *BaseOplog) error {
	master, ok := theMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	SetDeleteObjectWithOplog(theMaster, types.StatusTransferred, oplog)

	master.TransferToID = oplog.ObjID

	return nil
}
