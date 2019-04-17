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

func (pm *BaseProtocolManager) handleMigrateMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandleTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MemberMerkle(),

		types.StatusMigrated,

		pm.SetMemberDB,
		pm.postmigrateMember,
	)
}

/*
PosttransferMember deals with the situation after transferring the member:

1. do add member first.
2. if fromID is me:
    => if toID is one of my ids: Doing migration
    => no: transfer to others.
*/
func (pm *BaseProtocolManager) postmigrateMember(fromID *types.PttID, toID *types.PttID, theMember Object, oplog *BaseOplog, opData OpData) error {
	_, ok := theMember.(*Member)
	if !ok {
		return ErrInvalidData
	}

	// 1. posttransfer
	origPerson := NewEmptyMember()
	pm.SetMemberObjDB(origPerson)

	_, err := pm.posttransferPerson(
		toID,
		oplog,
		origPerson,

		pm.NewMember,
		pm.postaddMember,
	)
	if err != nil {
		return err
	}

	// 2. check myID and fromID
	myID := pm.Ptt().GetMyEntity().GetID()
	if !reflect.DeepEqual(myID, fromID) {
		return nil
	}

	// 4. entity owner
	log.Debug("postmigrateMember: to check owner", "fromID", fromID, "toID", toID, "myID", myID)
	entity := pm.Entity()
	if entity.IsOwner(fromID) {
		entity.RemoveOwnerID(fromID)
	}

	if !entity.IsOwner(toID) {
		entity.AddOwnerID(toID)
	}
	err = entity.Save(false)
	if err != nil {
		return err
	}

	return err
}

func (pm *BaseProtocolManager) handlePendingMigrateMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandlePendingTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MemberMerkle(),

		types.StatusInternalMigrate,
		types.StatusPendingMigrate,
		types.StatusMigrated,

		pm.SetMemberDB,
	)
}

func (pm *BaseProtocolManager) setNewestMigrateMemberLog(oplog *BaseOplog) (types.Bool, error) {
	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	return pm.SetNewestTransferPersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) handleFailedMigrateMemberLog(oplog *BaseOplog) error {

	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	pm.HandleFailedTransferPersonLog(oplog, obj)

	return nil
}

func (pm *BaseProtocolManager) handleFailedValidMigrateMemberLog(oplog *BaseOplog) error {

	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	pm.HandleFailedValidTransferPersonLog(oplog, obj)

	return nil
}
