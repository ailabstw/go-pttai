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
)

func (pm *BaseProtocolManager) MigrateMember(fromID *types.PttID, toID *types.PttID) error {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	ownerIDs := pm.Entity().GetOwnerIDs()

	// 1. validate
	isValid := false

	if pm.IsMaster(fromID, false) {
		return types.ErrInvalidID
	}

	if pm.IsMaster(myID, false) {
		isValid = true
	}

	for _, ownerID := range ownerIDs {
		if reflect.DeepEqual(fromID, ownerID) {
			isValid = true
		}
	}
	if !isValid {
		return types.ErrInvalidID
	}

	// 2. do transfer-person
	origPerson := NewEmptyMember()
	pm.SetMemberObjDB(origPerson)
	data := &PersonOpTransferPerson{ToID: toID}

	err := pm.TransferPerson(
		fromID,
		toID,

		MemberOpTypeMigrateMember,
		origPerson,
		data,

		pm.MemberMerkle(),

		types.StatusInternalMigrate,
		types.StatusPendingMigrate,
		types.StatusMigrated,

		pm.SetMemberDB,
		pm.NewMemberOplog,
		pm.signMigrateMemberOplog,
		pm.setMigrateMemberWithOplog,
		pm.broadcastMemberOplogCore,
		pm.postmigrateMember,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) signMigrateMemberOplog(oplog *BaseOplog, fromID *types.PttID, toID *types.PttID) error {
	return pm.ForceSignOplog(oplog)
}

func (pm *BaseProtocolManager) setMigrateMemberWithOplog(theMember Object, oplog *BaseOplog) error {

	member, ok := theMember.(*Member)
	if !ok {
		return ErrInvalidData
	}

	SetDeleteObjectWithOplog(theMember, types.StatusMigrated, oplog)

	member.TransferToID = oplog.ObjID

	return nil
}
