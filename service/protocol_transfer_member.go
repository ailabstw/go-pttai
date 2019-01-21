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

func (pm *BaseProtocolManager) TransferMember(fromID *types.PttID, toID *types.PttID) error {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	ownerIDs := pm.Entity().GetOwnerIDs()

	// 1. validate
	isValid := false

	masters, _ := pm.GetMasterListFromCache(false)
	for _, master := range masters {
		log.Debug("TransferMember: (in-for-loop)", "master", master.ID, "status", master.Status)
	}

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

		MemberOpTypeTransferMember,
		origPerson,
		data,

		pm.MemberMerkle(),

		pm.SetMemberDB,
		pm.NewMemberOplog,
		pm.signMemberOplog,
		pm.setTransferMemberWithOplog,
		pm.broadcastMemberOplogCore,
		pm.posttransferMember,
	)
	log.Debug("TransferMember: after TransferPerson", "e", err)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) signMemberOplog(oplog *BaseOplog, fromID *types.PttID, toID *types.PttID) error {
	return pm.SignOplog(oplog)
}

func (pm *BaseProtocolManager) setTransferMemberWithOplog(theMember Object, oplog *BaseOplog) error {

	member, ok := theMember.(*Member)
	if !ok {
		return ErrInvalidData
	}

	SetDeleteObjectWithOplog(theMember, types.StatusTransferred, oplog)

	member.TransferToID = oplog.ObjID

	return nil
}
