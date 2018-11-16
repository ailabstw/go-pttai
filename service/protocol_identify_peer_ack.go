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
)

type IdentifyPeerAck struct {
	AckChallenge []byte        `json:"A,omitempty"`
	Hash         []byte        `json:"H,omitempty"`
	Sig          []byte        `json:"S,omitempty"`
	PubBytes     []byte        `json:"P,omitempty"`
	Extra        *KeyExtraInfo `json:"E,omitempty"`
	MyID         *types.PttID  `json:"M,omitempty"`
}

/*
IdentifyPeerAck acks IdentifyPeer
	1. return my data
	2. if we do not know the peer, do identify peer process.
*/
func (pm *BaseProtocolManager) IdentifyPeerAck(data *IdentifyPeer, peer *PttPeer) error {

	ptt := pm.Ptt()

	ackData, err := ptt.IdentifyPeerAck(data.Challenge, peer)
	if err != nil {
		return err
	}

	pm.SendDataToPeer(IdentifyPeerAckMsg, ackData, peer)

	if peer.UserID != nil {
		return nil
	}

	pm.IdentifyPeer(peer)

	return nil
}

/*
HandleIdentifyPeerAck handles IdentifyPeerAck
	1. if we've already know the user-id: return.
	2. try to cancel peer waiting for identification.
	3. have ptt to do FinishIdentifyPeer
*/
func (pm *BaseProtocolManager) HandleIdentifyPeerAck(dataBytes []byte, peer *PttPeer) error {
	data := &IdentifyPeerAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	log.Debug("HandleIdentifyPeerAck: to ptt.HandleIdentifyPeerAck")

	ptt := pm.Ptt()

	return ptt.HandleIdentifyPeerAck(pm.Entity().GetID(), data, peer)
}

/**********
 * Ptt
 **********/

/*
IdentifyPeerWithMyIDAck
*/
func (p *BasePtt) IdentifyPeerWithMyIDAck(challenge *types.Salt, peer *PttPeer) error {
	ackData, err := p.IdentifyPeerAck(challenge, peer)
	if err != nil {
		return err
	}

	return p.SendDataToPeer(CodeTypeIdentifyPeerWithMyIDAck, ackData, peer)
}

/*
HandleIdentifyPeerWithMyIDAck
*/
func (p *BasePtt) HandleIdentifyPeerWithMyIDAck(dataBytes []byte, peer *PttPeer) error {
	if p.myEntity == nil {
		return ErrInvalidEntity
	}

	myID := p.myEntity.GetID()

	data := &IdentifyPeerAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return p.HandleIdentifyPeerAck(myID, data, peer)
}

/*
IdentifyPeerAck
*/
func (p *BasePtt) IdentifyPeerAck(challenge *types.Salt, peer *PttPeer) (*IdentifyPeerAck, error) {
	if p.myEntity == nil {
		return nil, ErrInvalidEntity
	}

	signKey := p.myEntity.SignKey()
	if signKey == nil {
		return nil, ErrInvalidKey
	}

	bytesWithSalt, hash, sig, pubBytes, err := SignData(challenge[:], signKey)
	if err != nil {
		return nil, err
	}

	myID := p.myEntity.GetID()

	ackData := &IdentifyPeerAck{
		AckChallenge: bytesWithSalt,
		Hash:         hash,
		Sig:          sig,
		PubBytes:     pubBytes,
		MyID:         myID,
		Extra:        signKey.Extra,
	}

	return ackData, nil
}

/*
HandleIdentifyPeerAck
*/

func (p *BasePtt) HandleIdentifyPeerAck(entityID *types.PttID, data *IdentifyPeerAck, peer *PttPeer) error {

	if !reflect.DeepEqual(peer.IDChallenge[:], data.AckChallenge[:types.SizeSalt]) {
		log.Warn("HandleIdentifyPeerAck: unable to match challenge")
		return ErrInvalidData
	}

	err := VerifyData(data.AckChallenge, data.Hash, data.Sig, data.PubBytes, data.MyID, data.Extra)
	if err != nil {
		log.Warn("HandleIdentifyPeerAck: unable to verify data", "peer", peer)
		return err
	}

	if peer.UserID != nil {
		log.Debug("HandleIdentifyPeerAck: already known user-id", "peer", peer)

		return nil
	}

	peer.UserID = data.MyID

	peer.FinishID(entityID)

	log.Debug("HandleIdentifyPeerAck: to FinishIdentifyPeer", "peer", peer, "userID", peer.UserID)

	return p.FinishIdentifyPeer(peer, false, false)
}
