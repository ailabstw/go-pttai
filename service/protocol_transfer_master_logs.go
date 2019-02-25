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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) handleTransferMasterLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandleTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MasterMerkle(),

		types.StatusTransferred,

		pm.SetMasterDB,
		pm.posttransferMaster,
	)
}

func (pm *BaseProtocolManager) posttransferMaster(fromID *types.PttID, toID *types.PttID, theMaster Object, oplog *BaseOplog, opData OpData) error {
	_, ok := theMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	origPerson := NewEmptyMaster()
	pm.SetMasterObjDB(origPerson)

	log.Debug("posttransferMaster: start", "fromID", fromID, "toID", toID)

	newMaster, err := pm.posttransferPerson(
		toID,
		oplog,
		origPerson,

		pm.NewMaster,
		nil,
	)
	log.Debug("posttransferMaster: after posttransferPerson", "e", err)
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

func (pm *BaseProtocolManager) postposttransferMaster(theMaster Object, theNewMaster Object, oplog *BaseOplog) error {

	master, ok := theMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	newMaster, ok := theNewMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	err := pm.SetNewestMasterLogID(oplog.ID)
	log.Debug("postposttransferMaster: after SetNewestMasterLogID", "e", err)
	if err != nil {
		return err
	}

	pm.lockMaster.Lock()
	defer pm.lockMaster.Unlock()

	err = pm.UnregisterMaster(master, true)
	if err != nil {
		return err
	}
	err = pm.RegisterMaster(newMaster, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) handlePendingTransferMasterLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandlePendingTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MasterMerkle(),

		types.StatusInternalTransfer,
		types.StatusPendingTransfer,
		types.StatusTransferred,

		pm.SetMasterDB,
	)
}

func (pm *BaseProtocolManager) setNewestTransferMasterLog(oplog *BaseOplog) (types.Bool, error) {
	obj := NewEmptyMaster()
	pm.SetMasterObjDB(obj)

	return pm.SetNewestTransferPersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) handleFailedTransferMasterLog(oplog *BaseOplog) error {

	obj := NewEmptyMaster()
	pm.SetMasterObjDB(obj)

	pm.HandleFailedTransferPersonLog(oplog, obj)

	return nil
}

func (pm *BaseProtocolManager) handleFailedValidTransferMasterLog(oplog *BaseOplog) error {

	obj := NewEmptyMaster()
	pm.SetMasterObjDB(obj)

	pm.HandleFailedValidTransferPersonLog(oplog, obj)

	return nil
}
