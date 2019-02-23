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

func (pm *BaseProtocolManager) handleMigrateMasterLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandleTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MasterMerkle(),

		types.StatusMigrated,

		pm.SetMasterDB,
		pm.postmigrateMaster,
	)
}

func (pm *BaseProtocolManager) postmigrateMaster(
	fromID *types.PttID,
	toID *types.PttID,
	theMaster Object,
	oplog *BaseOplog,
	opData OpData,
) error {
	_, ok := theMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	origPerson := NewEmptyMaster()
	pm.SetMasterObjDB(origPerson)

	newMaster, err := pm.posttransferPerson(
		toID,
		oplog,
		origPerson,

		pm.NewMaster,
		nil,
	)
	if err != nil {
		return err
	}

	if pm.inposttransferMaster != nil {
		err = pm.inposttransferMaster(theMaster, newMaster, oplog)
		if err != nil {
			return err
		}
	}

	return pm.postposttransferMaster(theMaster, newMaster, oplog)
}

func (pm *BaseProtocolManager) handlePendingMigrateMasterLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandlePendingTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MasterMerkle(),

		types.StatusInternalMigrate,
		types.StatusPendingMigrate,
		types.StatusMigrated,

		pm.SetMasterDB,
	)
}

func (pm *BaseProtocolManager) setNewestMigrateMasterLog(oplog *BaseOplog) (types.Bool, error) {
	obj := NewEmptyMaster()
	pm.SetMasterObjDB(obj)

	return pm.SetNewestTransferPersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) handleFailedMigrateMasterLog(oplog *BaseOplog) error {

	obj := NewEmptyMaster()
	pm.SetMasterObjDB(obj)

	pm.HandleFailedTransferPersonLog(oplog, obj)

	return nil
}

func (pm *BaseProtocolManager) handleFailedValidMigrateMasterLog(oplog *BaseOplog) error {

	obj := NewEmptyMaster()
	pm.SetMasterObjDB(obj)

	pm.HandleFailedValidTransferPersonLog(oplog, obj)

	return nil
}
