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

func (pm *BaseProtocolManager) handleDeleteMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	opData := &MemberOpDeleteMember{}

	toBroadcastLogs, err := pm.HandleDeletePersonLog(
		oplog,
		info,

		obj,
		opData,

		types.StatusDeleted,

		pm.MemberMerkle(),

		pm.SetMemberDB,

		pm.postdeleteMember,
		pm.updateDeleteMemberInfo,
	)
	if err != nil {
		return nil, err
	}

	return toBroadcastLogs, nil
}

func (pm *BaseProtocolManager) handlePendingDeleteMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	opData := &MemberOpDeleteMember{}

	log.Debug("handlePendingDeleteMemberLog: to HandlePendingDeletePersonLog", "entity", pm.Entity().GetID())

	return pm.HandlePendingDeletePersonLog(
		oplog,
		info,

		obj,
		opData,

		types.StatusInternalDeleted,
		types.StatusPendingDeleted,

		pm.MemberMerkle(),

		pm.SetMemberDB,
		pm.updateDeleteMemberInfo,
	)
}

func (pm *BaseProtocolManager) setNewestDeleteMemberLog(oplog *BaseOplog) (types.Bool, error) {
	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	return pm.SetNewestDeletePersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) handleFailedDeleteMemberLog(oplog *BaseOplog) error {
	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	return pm.HandleFailedDeletePersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) handleFailedValidDeleteMemberLog(oplog *BaseOplog) error {
	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	return pm.HandleFailedValidDeletePersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) updateDeleteMemberInfo(
	member Object,
	oplog *BaseOplog,
	theInfo ProcessInfo,
) error {

	info, ok := theInfo.(*ProcessPersonInfo)
	if !ok {
		return ErrInvalidData
	}

	personID := oplog.ObjID
	delete(info.CreateInfo, *personID)
	info.DeleteInfo[*personID] = oplog

	return nil
}
