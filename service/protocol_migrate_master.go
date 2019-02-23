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

func (pm *BaseProtocolManager) MigrateMaster(id *types.PttID) error {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	// 1. validate
	if !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	// 2. do transfer-person
	origPerson := NewEmptyMaster()
	pm.SetMasterObjDB(origPerson)
	data := &PersonOpTransferPerson{ToID: id}

	err := pm.TransferPerson(
		myID,
		id,

		MasterOpTypeMigrateMaster,
		origPerson,
		data,

		pm.MasterMerkle(),

		types.StatusInternalMigrate,
		types.StatusPendingMigrate,
		types.StatusMigrated,

		pm.SetMasterDB,
		pm.NewMasterOplog,
		pm.signMigrateMasterOplog,
		pm.setTransferMasterWithOplog,
		pm.broadcastMasterOplogCore,
		pm.posttransferMaster,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) signMigrateMasterOplog(oplog *BaseOplog, fromID *types.PttID, toID *types.PttID) error {
	return pm.ForceSignOplog(oplog)
}

func (pm *BaseProtocolManager) setMigrateMasterWithOplog(theMaster Object, oplog *BaseOplog) error {
	master, ok := theMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	SetDeleteObjectWithOplog(theMaster, types.StatusMigrated, oplog)

	master.TransferToID = oplog.ObjID

	return nil
}
