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

func (pm *BaseProtocolManager) handleAddMasterLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	opData := &MasterOpCreateMaster{}

	if oplog.PreLogID == nil {
		return pm.HandleCreatePersonLog(oplog, person, opData, pm.postaddMaster)
	} else {
		return pm.HandleUpdatePersonLog(oplog, person, opData, pm.SetMasterDB, pm.postaddMaster)

	}
}

func (pm *BaseProtocolManager) handlePendingAddMasterLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	opData := &MasterOpCreateMaster{}

	if oplog.PreLogID == nil {
		return pm.HandlePendingCreatePersonLog(oplog, person, opData)
	} else {
		return pm.HandlePendingUpdatePersonLog(oplog, person, opData, pm.SetMasterDB)
	}
}

func (pm *BaseProtocolManager) setNewestAddMasterLog(oplog *BaseOplog) (types.Bool, error) {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	return pm.SetNewestPersonLog(oplog, person)
}

func (pm *BaseProtocolManager) handleFailedAddMasterLog(oplog *BaseOplog) error {

	person := NewEmptyMaster()
	pm.SetMasterObjDB(person)

	if oplog.PreLogID == nil {
		return pm.HandleFailedCreatePersonLog(oplog, person, nil)
	} else {
		return pm.HandleFailedUpdatePersonLog(oplog, person)
	}
}

/**********
 * Customize
 **********/

func (pm *BaseProtocolManager) updateCreateMasterInfo(master Object, oplog *BaseOplog, theOpData OpData, theInfo ProcessInfo) error {

	info, ok := theInfo.(*ProcessPersonInfo)
	if !ok {
		return ErrInvalidData
	}

	personID := oplog.ObjID
	delete(info.DeleteInfo, *personID)
	info.CreateInfo[*personID] = oplog

	return nil
}
