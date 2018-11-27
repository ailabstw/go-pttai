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

type ApproveJoinEntity struct {
	MyID       *types.PttID `json:"ID"`
	Entity     Entity       `json:"e"`
	Oplog0     *BaseOplog   `json:"0"`
	MasterLogs []*BaseOplog `json:"M"`
	MemberLogs []*BaseOplog `json:"m"`
	OpKey      *KeyInfo     `json:"O"`
	OpKeyLog   *BaseOplog   `json:"o"`
}

/*
Required variables in joinEntity: ID
*/
func (pm *BaseProtocolManager) ApproveJoin(
	joinEntity *JoinEntity,
	keyInfo *KeyInfo,
	peer *PttPeer,
) (*KeyInfo, interface{}, error) {
	log.Debug("ApproveJoin: start", "name", pm.Entity().Name(), "service", pm.Entity().Service().Name(), "peer", peer, "peerType", peer.PeerType)

	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) && peer.PeerType != PeerTypeMe {
		return nil, nil, types.ErrInvalidID
	}

	if pm.myMemberLog == nil {
		return nil, nil, types.ErrInvalidStatus
	}

	opKey, err := pm.GetNewestOpKey(false)
	log.Debug("ApproveJoin: after GetNewestOpKey", "err", err)
	if err != nil {
		return nil, nil, err
	}
	opKeyLog := &OpKeyOplog{BaseOplog: &BaseOplog{}}
	pm.SetOpKeyDB(opKeyLog.BaseOplog)
	err = opKeyLog.Get(opKey.LogID, false)
	if err != nil {
		return nil, nil, err
	}

	// entity
	entity := pm.Entity()

	// status
	if entity.GetStatus() != types.StatusAlive {
		return nil, nil, ErrInvalidStatus
	}

	// master oplog
	oplog := &BaseOplog{}
	pm.SetMasterDB(oplog)
	masterLogs, err := GetOplogList(oplog, nil, 0, pttdb.ListOrderNext, types.StatusAlive, false)
	log.Debug("ApproveJoin: after get master oplogs", "err", err)
	if err != nil {
		return nil, nil, err
	}

	// member oplog
	// XXX (force adding member. We are single-master for now.)
	var memberLog *MemberOplog
	memberLogs := make([]*BaseOplog, 0, 2)
	if !reflect.DeepEqual(myID, joinEntity.ID) {
		log.Debug("ApproveJoin: peer not me", "joinEntity", joinEntity.ID, "myID", entity.GetCreatorID())
		_, memberLog, err = pm.AddMember(joinEntity.ID, true)
		log.Debug("ApproveJoin: after AddMember", "e", err)
		if err == types.ErrAlreadyExists {
			memberLog, err = pm.GetMemberLogByMemberID(joinEntity.ID, false)
			if err != nil {
				return nil, nil, err
			}
		}
		if err != nil {
			return nil, nil, err
		}

		memberLogs = append(memberLogs, memberLog.BaseOplog)

	}
	memberLogs = append(memberLogs, pm.myMemberLog.BaseOplog)
	log.Debug("ApproveJoin: after get memberLogs", "memberLogs", memberLogs)

	// register-peer
	if peer.UserID == nil {
		peer.UserID = joinEntity.ID
	}

	switch {
	case peer.PeerType < PeerTypeMember:
		pm.Ptt().SetupPeer(peer, PeerTypeMember, false)
	case peer.PeerType == PeerTypeMe:
		pm.RegisterPeer(peer, PeerTypeMe)
	default:
		pm.RegisterPeer(peer, PeerTypeMember)
	}

	log.Debug("ApproveJoinEntity: done", "entity", entity.GetID(), "name", entity.Name(), "peer", peer, "service", entity.Service().Name())

	// approve-join
	approveJoin := &ApproveJoinEntity{
		MyID:       myID,
		Entity:     entity,
		Oplog0:     pm.oplog0,
		MasterLogs: masterLogs,
		MemberLogs: memberLogs,
		OpKey:      opKey,
		OpKeyLog:   opKeyLog.BaseOplog,
	}

	return opKey, approveJoin, nil
}
