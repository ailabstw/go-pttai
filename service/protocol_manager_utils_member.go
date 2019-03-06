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
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *BaseProtocolManager) IsMember(id *types.PttID, isLocked bool) bool {
	return pm.isMember(id, isLocked)
}
func (pm *BaseProtocolManager) defaultIsMember(id *types.PttID, isLocked bool) bool {
	member, err := pm.GetMember(id, isLocked)
	if err != nil {
		return false
	}
	return member.Status == types.StatusAlive
}

func (pm *BaseProtocolManager) IsPendingMember(id *types.PttID, isLocked bool) bool {
	peer, err := pm.GetPendingPeerByUserID(id, false)
	if err != nil {
		return false
	}
	if peer == nil {
		return false
	}

	return true
}

func (pm *BaseProtocolManager) GetMember(id *types.PttID, isLocked bool) (*Member, error) {
	member := NewEmptyMember()
	pm.SetMemberObjDB(member)
	member.SetID(id)

	err := member.GetByID(isLocked)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (pm *BaseProtocolManager) RegisterMember(member *Member, isLocked bool) error {

	return pm.Ptt().RegisterEntityPeerWithOtherUserID(pm.Entity(), member.ID, PeerTypeMember, false)
}

/*
UnregisterMember unregisters member. Possibly already did postdeleteEntity.
*/
func (pm *BaseProtocolManager) UnregisterMember(member *Member, isLock bool, isPostdeleteEntity bool) error {
	myID := pm.Ptt().GetMyEntity().GetID()

	if reflect.DeepEqual(myID, member.ID) {
		if isPostdeleteEntity {
			pm.PostdeleteEntity(nil, false)
		}
		return nil
	}

	return pm.UnregisterPeerByOtherUserID(member.ID, false, false)
}

func (pm *BaseProtocolManager) SetMemberSyncTime(ts types.Timestamp) error {
	return pm.MemberMerkle().SaveSyncTime(ts)
}

func (pm *BaseProtocolManager) GetMemberLogByMemberID(id *types.PttID, isLocked bool) (*MemberOplog, error) {

	member, err := pm.GetMember(id, isLocked)
	if err != nil {
		return nil, err
	}
	if member.LogID == nil {
		return nil, err
	}
	memberLog := &MemberOplog{BaseOplog: &BaseOplog{}}
	pm.SetMemberDB(memberLog.BaseOplog)
	logID := member.GetNewestLogID()
	log.Debug("GetMemberLogByMemberID: after GetNewestLogID", "id", id, "logID", logID, "entity", pm.Entity().GetID(), "service", pm.entity.Service().Name())
	err = memberLog.Get(logID, false)
	if err != nil {
		return nil, err
	}
	return memberLog, nil
}

func (pm *BaseProtocolManager) loadMyMemberLog() error {
	myID := pm.Ptt().GetMyEntity().GetID()
	if !pm.Entity().IsOwner(myID) {
		return nil
	}
	memberLog, err := pm.GetMemberLogByMemberID(myID, false)
	if err != nil {
		return err
	}
	pm.myMemberLog = memberLog

	return nil
}

func (pm *BaseProtocolManager) CleanMember(isRetainMe bool) error {
	members, err := pm.GetMemberList(nil, 0, pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}

	myID := pm.Ptt().GetMyEntity().GetID()
	for _, member := range members {
		if isRetainMe && reflect.DeepEqual(myID, member.ID) {
			continue
		}
		member.Delete(false)
	}

	return nil
}
