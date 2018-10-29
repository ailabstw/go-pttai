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
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

/*
JoinEntity confirmed the invitor from HandleJoinAckChallenge and requests joining the entity (joiner)
*/
func (p *BasePtt) JoinEntity(joinRequest *JoinRequest, joinAckChallenge *JoinAckChallenge, peer *PttPeer) error {

	joinRequest.ID = joinAckChallenge.ID
	joinRequest.Name = joinAckChallenge.Name
	joinRequest.Status = JoinStatusWaitAccepted
	joinRequest.Master0Hash = joinAckChallenge.Master0Hash

	id := p.myEntity.GetID()
	name := p.myEntity.Name()

	joinEntity := &JoinEntity{
		ID:          id,
		Name:        []byte(name),
		Master0Hash: joinRequest.Master0Hash,
	}

	data, err := json.Marshal(joinEntity)
	if err != nil {
		return err
	}

	keyInfo := joinKeyToKeyInfo(joinRequest.Key)

	encData, err := p.EncryptData(JoinEntityMsg, data, keyInfo)
	if err != nil {
		return err
	}

	pttData, err := p.MarshalData(CodeTypeJoin, joinRequest.Hash, encData)
	if err != nil {
		return err
	}

	pttData.Node = peer.GetID()[:]
	err = peer.SendData(pttData)

	return nil
}

/*
HandleJoinEntity

Recevied "join-entity" with revealed ID and Name. (invitor)
    1. if the entity auto-rejects the entity-id and node-id:
        => return err
    2. if the entity auto-approves the entity-id and node-id
        => do approve.
    3. put to confirm-queue.
*/
func (p *BasePtt) HandleJoinEntity(dataBytes []byte, hash *common.Address, entity Entity, pm ProtocolManager, keyInfo *KeyInfo, peer *PttPeer) error {
	log.Debug("HandleJoinEntity: start")
	joinEntity := &JoinEntity{}
	err := json.Unmarshal(dataBytes, joinEntity)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(joinEntity.Master0Hash, entity.PM().Master0Hash()) {
		return ErrInvalidData
	}

	id := joinEntity.ID
	nodeID := peer.GetID()
	if entity.PM().IsSuspiciousID(id, nodeID) {
		return ErrInvalidData
	}

	confirmKey := getConfirmKey(id, entity.GetID())
	joinType, err := entity.PM().GetJoinType(hash)
	if err != nil {
		return err
	}

	err = p.ToConfirmJoin(confirmKey, entity, joinEntity, keyInfo, peer, joinType)
	if err != nil {
		return err
	}

	if entity.PM().IsGoodID(id, nodeID) {
		return p.ApproveJoin(confirmKey)
	}

	return nil
}

func getConfirmKey(id *types.PttID, entityID *types.PttID) []byte {
	confirmKey := make([]byte, types.SizePttID+types.SizePttID)
	copy(confirmKey[:types.SizePttID], id[:])
	copy(confirmKey[types.SizePttID:], entityID[:])

	return confirmKey
}
