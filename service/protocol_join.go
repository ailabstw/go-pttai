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
	"crypto/ecdsa"
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ethereum/go-ethereum/common"
)

/*
TryJoin trys to do join (joiner)
	1. If the nodeID is my node: return err because we dont join entity from our devices

	2. If the nodeID is my peer: do join.
	3. Else: do add peer.
*/
func (p *BasePtt) TryJoin(challenge []byte, hash *common.Address, key *ecdsa.PrivateKey, request *JoinRequest) error {
	nodeID := request.NodeID

	p.peerLock.RLock()
	defer p.peerLock.RUnlock()

	log.Debug("TryJoin start", "challenge", challenge, "hash", hash)

	// 1. my node
	_, ok := p.myPeers[*nodeID]
	if ok {
		return ErrAlreadyMyNode
	}

	// 2. my peer
	var err error
	peer := p.hubPeers[*nodeID]
	if peer != nil {
		err = p.join(challenge, hash, key, peer)
		if err != nil {
			return err
		}
		request.Status = JoinStatusRequested
		return nil
	}

	peer = p.importantPeers[*nodeID]
	if peer != nil {
		err = p.join(challenge, hash, key, peer)
		if err != nil {
			return err
		}
		request.Status = JoinStatusRequested
		return nil

	}

	peer = p.memberPeers[*nodeID]
	if peer != nil {
		err = p.join(challenge, hash, key, peer)
		if err != nil {
			return err
		}
		request.Status = JoinStatusRequested
		return nil
	}

	peer = p.pendingPeers[*nodeID]
	if peer != nil {
		err = p.join(challenge, hash, key, peer)
		if err != nil {
			return err
		}
		request.Status = JoinStatusRequested
		return nil
	}

	peer = p.randomPeers[*nodeID]
	if peer != nil {
		err = p.join(challenge, hash, key, peer)
		if err != nil {
			return err
		}
		request.Status = JoinStatusRequested
		return nil
	}

	// 3. add peer
	node := discover.NewWebrtcNode(*nodeID)
	p.Server().AddPeer(node)

	return nil
}

/*
Join initiates joining a specific entity with the peer.
*/
func (p *BasePtt) join(challenge []byte, hash *common.Address, joinKey *ecdsa.PrivateKey, peer *PttPeer) error {
	join := &Join{
		Hash:      hash[:],
		Challenge: challenge,
	}

	data, err := json.Marshal(join)
	if err != nil {
		return err
	}

	keyInfo := joinKeyToKeyInfo(joinKey)

	encData, err := p.EncryptData(JoinMsg, data, keyInfo)
	if err != nil {
		return err
	}

	pttData, err := p.MarshalData(CodeTypeJoin, hash, encData)
	if err != nil {
		return err
	}

	pttData.Node = peer.GetID()[:]
	log.Debug("join: to SendData")
	err = peer.SendData(pttData)
	if err != nil {
		return err
	}

	return nil
}

func (p *BasePtt) HandleJoin(dataBytes []byte, hash *common.Address, entity Entity, pm ProtocolManager, keyInfo *KeyInfo, peer *PttPeer) error {
	join := &Join{}
	err := json.Unmarshal(dataBytes, join)
	if err != nil {
		return err
	}

	log.Debug("HandleJoin: start")

	if !reflect.DeepEqual(join.Hash, hash[:]) {
		log.Error("handleJoinCore: Hash not the same", "hash", hash[:], "join.Hash", join.Hash)
		return ErrInvalidData
	}

	return p.JoinAckChallenge(keyInfo, join, peer, entity)
}
