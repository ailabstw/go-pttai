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

func (pm *BaseProtocolManager) DeleteMember(
	id *types.PttID,
) (bool, error) {

	person := NewEmptyMember()
	pm.SetMemberObjDB(person)

	opData := &MemberOpDeleteMember{}

	err := pm.DeletePerson(
		id,
		MemberOpTypeDeleteMember,
		person,
		opData,

		types.StatusInternalDeleted,
		types.StatusPendingDeleted,
		types.StatusDeleted,

		pm.MemberMerkle(),

		pm.SetMemberDB,
		pm.NewMemberOplog,
		pm.broadcastMemberOplogCore,
		pm.postdeleteMember,
	)
	log.Debug("DeleteMember: after DeletePerson", "e", err)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) postdeleteMember(
	id *types.PttID,
	oplog *BaseOplog,
	origObj Object,
	opData OpData,
) error {

	var err error

	myID := pm.Ptt().GetMyEntity().GetID()
	entity := pm.Entity()

	log.Debug("postdeleteMember: start", "id", id, "entity", pm.Entity().IDString(), "myID", myID)

	if pm.inpostdeleteMember != nil {
		err = pm.inpostdeleteMember(id, oplog, origObj, opData)
		if err != nil {
			return err
		}
	}

	if reflect.DeepEqual(myID, oplog.ObjID) {
		pm.myMemberLog = OplogToMemberOplog(oplog)
		entity.SetStatus(types.StatusDeleted)
		entity.SetUpdateTS(pm.myMemberLog.UpdateTS)
		entity.Save(false)

		if pm.postdelete != nil {
			pm.postdelete(opData, true)
		}
	} else {
		pm.UnregisterPeerByOtherUserID(oplog.ObjID, true, false)
	}

	return nil
}
