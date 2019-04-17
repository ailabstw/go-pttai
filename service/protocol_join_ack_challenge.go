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

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"
)

/*
JoinAckChallenge acks join-challenge to show that I have the join-key. (invitor)
*/
func (p *BasePtt) JoinAckChallenge(keyInfo *KeyInfo, join *Join, peer *PttPeer, entity Entity) error {

	challenge := join.Challenge
	id := entity.GetID()
	name := entity.Name()
	hash := entity.PM().MasterLog0Hash()

	joinAckCallenge := &JoinAckChallenge{
		Challenge:   challenge,
		ID:          id,
		Name:        []byte(name),
		Master0Hash: hash,
	}

	data, err := json.Marshal(joinAckCallenge)
	if err != nil {
		return err
	}

	encData, err := p.EncryptData(JoinAckChallengeMsg, data, keyInfo)
	if err != nil {
		return err
	}

	pttData, err := p.MarshalData(CodeTypeJoinAck, keyInfo.Hash, encData)
	if err != nil {
		return err
	}

	log.Debug("JoinAckChallenge: to SendData", "entity", entity.IDString(), "peer", peer)
	pttData.Node = peer.GetID()[:]
	err = peer.SendData(pttData)

	return nil
}

/*
HandleJoinAckChallenge
*/
func (p *BasePtt) HandleJoinAckChallenge(dataBytes []byte, hash *common.Address, joinRequest *JoinRequest, peer *PttPeer) error {
	log.Debug("HandleJoinAckChallenge: start")

	if joinRequest.Status != JoinStatusRequested {
		log.Error("HandleJoinAckChallenge: joinRequest status err", "peer", peer, "Status", joinRequest.Status)
		return ErrInvalidData
	}

	joinAckChallenge := &JoinAckChallenge{}
	err := json.Unmarshal(dataBytes, joinAckChallenge)
	if err != nil {
		log.Error("HandleJoinAckChallenge: unable to unmarshal data", "peer", peer, "e", err)
		return ErrInvalidData
	}

	if !reflect.DeepEqual(joinRequest.Challenge, joinAckChallenge.Challenge) {
		log.Error("HandleJoinAckChallenge: challenge not the same", "peer", peer, "send", joinRequest.Challenge, "recv", joinAckChallenge.Challenge)
		return ErrInvalidData
	}

	id := joinAckChallenge.ID

	entity, ok := p.entities[*id]
	if ok && entity.GetStatus() == types.StatusAlive {
		log.Error("HandleJoinAckChallenge: already registered", "entity", entity.IDString())
		return ErrAlreadyRegistered
	}

	// entity needs to meets creator-id
	if !reflect.DeepEqual(id[common.AddressLength:], joinRequest.CreatorID[:common.AddressLength]) && !reflect.DeepEqual(id, joinRequest.CreatorID) {
		log.Error("HandleJoinAckChallenge: creator not meet id", "peer", peer, "id", id, "creator", joinRequest.CreatorID)
		return ErrInvalidData
	}

	log.Debug("HandleJoinAckChallenge: to JoinEntity")

	return p.JoinEntity(joinRequest, joinAckChallenge, peer)
}
