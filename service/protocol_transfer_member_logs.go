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

func (pm *BaseProtocolManager) handleTransferMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandleTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MemberMerkle(),

		pm.SetMemberDB,
		pm.posttransferMember,
	)
}

/*
PosttransferMember deals with the situation after transferring the member:

1. do add member first.
2. if fromID is me:
    => if toID is one of my ids: Doing migration
    => no: transfer to others.
*/
func (pm *BaseProtocolManager) posttransferMember(fromID *types.PttID, toID *types.PttID, theMember Object, oplog *BaseOplog, opData OpData) error {
	_, ok := theMember.(*Member)
	if !ok {
		return ErrInvalidData
	}

	// 1. posttransfer
	origPerson := NewEmptyMember()
	pm.SetMemberObjDB(origPerson)

	_, err := pm.posttransferPerson(
		toID, oplog, origPerson,
		pm.NewMember, pm.postaddMember,
	)
	log.Debug("posttransferMember: after posttransferPerson", "e", err)
	if err != nil {
		return err
	}

	// 2. check myID and fromID
	myID := pm.Ptt().GetMyEntity().GetID()
	log.Debug("posttransferMember: to check myID", "myID", myID, "fromID", fromID)
	if !reflect.DeepEqual(myID, fromID) {
		return nil
	}

	// 3. check migrate
	isMigrate := false
	mySPM := pm.Ptt().GetMyService().SPM()
	for _, entity := range mySPM.Entities() {
		if reflect.DeepEqual(entity.GetID(), toID) {
			isMigrate = true
			break
		}
	}

	if !isMigrate {
		log.Debug("posttransferMember: not Migrate: to delete Entity", "entity", pm.Entity().GetID())
		pm.PostdeleteEntity(nil, true)
		return nil
	}

	// 4. entity owner
	log.Debug("posttransferMember: to check owner", "fromID", fromID, "toID", toID, "myID", myID)
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

func (pm *BaseProtocolManager) handlePendingTransferMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) (types.Bool, []*BaseOplog, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &PersonOpTransferPerson{}

	return pm.HandlePendingTransferPersonLog(
		oplog,
		person,
		opData,

		pm.MemberMerkle(),

		pm.SetMemberDB,
	)
}

func (pm *BaseProtocolManager) setNewestTransferMemberLog(oplog *BaseOplog) (types.Bool, error) {
	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	return pm.SetNewestTransferPersonLog(oplog, obj)
}

func (pm *BaseProtocolManager) handleFailedTransferMemberLog(oplog *BaseOplog) error {

	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	pm.HandleFailedTransferPersonLog(oplog, obj)

	return nil
}

func (pm *BaseProtocolManager) handleFailedValidTransferMemberLog(oplog *BaseOplog) error {

	obj := NewEmptyMember()
	pm.SetMemberObjDB(obj)

	pm.HandleFailedValidTransferPersonLog(oplog, obj)

	return nil
}
