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

func (pm *BaseProtocolManager) AddMember(id *types.PttID, isForce bool) (*Member, *MemberOplog, error) {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()
	origMember := NewEmptyMember()
	pm.SetMemberObjDB(origMember)

	// 1. validate
	if !isForce && !pm.IsMaster(myID, false) {
		return nil, nil, types.ErrInvalidID
	}

	data := &MemberOpAddMember{}
	person, oplog, err := pm.AddPerson(
		id,
		MemberOpTypeAddMember,
		isForce,

		origMember,
		data,

		pm.MemberMerkle(),

		pm.NewMember,
		pm.NewMemberOplogWithTS,
		pm.broadcastMemberOplogCore,
		pm.postaddMember,

		pm.SetMemberDB,
		pm.NewMemberOplog,
	)

	if err != nil {
		return nil, nil, err
	}
	member, ok := person.(*Member)
	if !ok {
		return nil, nil, ErrInvalidObject
	}

	memberOplog := &MemberOplog{BaseOplog: oplog}

	return member, memberOplog, nil
}

func (pm *BaseProtocolManager) NewMember(id *types.PttID) (Object, OpData, error) {
	entity := pm.Entity()
	myEntity := pm.Ptt().GetMyEntity()
	myID := myEntity.GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	member := NewMember(id, ts, myID, entity.GetID(), nil, types.StatusInit)
	pm.SetMemberObjDB(member)

	return member, &MemberOpAddMember{}, nil
}

func (pm *BaseProtocolManager) postaddMember(theMember Object, oplog *BaseOplog) error {
	member, ok := theMember.(*Member)
	if !ok {
		return ErrInvalidData
	}

	log.Debug("postaddMember: start", "entity", pm.Entity().GetID(), "member", theMember.GetID(), "oplog", oplog.ID)

	myID := pm.Ptt().GetMyEntity().GetID()
	if reflect.DeepEqual(myID, member.GetID()) {
		log.Debug("postaddMember: to myMemberLog", "entity", pm.Entity().GetID(), "member", theMember.GetID(), "myID", myID, "oplog", oplog.ID)

		pm.myMemberLog = OplogToMemberOplog(oplog)

		return nil
	}

	log.Debug("postaddMember: to register-member", "entity", pm.Entity().GetID(), "member", theMember.GetID(), "oplog", oplog.ID)

	pm.RegisterMember(member, false)

	return nil
}
