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

func (pm *BaseProtocolManager) handleAddMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &MemberOpAddMember{}

	log.Debug("handleAddMemberLog: start", "oplog.PreLogID", oplog.PreLogID)

	if oplog.PreLogID == nil {
		return pm.HandleCreatePersonLog(oplog, person, opData, pm.postaddMember)
	} else {
		return pm.HandleUpdatePersonLog(oplog, person, opData, pm.SetMasterDB, pm.postaddMember)
	}
}

func (pm *BaseProtocolManager) handlePendingAddMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &MemberOpAddMember{}

	if oplog.PreLogID == nil {
		return pm.HandlePendingCreatePersonLog(oplog, person, opData)
	} else {
		return pm.HandlePendingUpdatePersonLog(oplog, person, opData, pm.SetMemberDB)
	}
}

func (pm *BaseProtocolManager) setNewestAddMemberLog(oplog *BaseOplog) (types.Bool, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	return pm.SetNewestPersonLog(oplog, person)
}

func (pm *BaseProtocolManager) handleFailedAddMemberLog(oplog *BaseOplog) error {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	if oplog.PreLogID == nil {
		return pm.HandleFailedCreatePersonLog(oplog, person, nil)
	} else {
		return pm.HandleFailedUpdatePersonLog(oplog, person)
	}
}

/**********
 * Customize
 **********/

func (pm *BaseProtocolManager) updateCreateMemberInfo(member Object, oplog *BaseOplog, theOpData OpData, theInfo ProcessInfo) error {

	info, ok := theInfo.(*ProcessPersonInfo)
	if !ok {
		return ErrInvalidData
	}

	personID := oplog.ObjID
	delete(info.DeleteInfo, *personID)
	info.CreateInfo[*personID] = oplog

	return nil
}
