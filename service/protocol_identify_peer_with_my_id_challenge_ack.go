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
)

type IdentifyPeerWithMyIDChallengeAck struct {
	Challenge *types.Salt      `json:"C"`
	AckData   *IdentifyPeerAck `json:"A"`
}

/*
IdentifyPeerWithMyIDChallengeAck acks IdentifyPeerWithMyIDChallenge (requester)
*/
func (p *BasePtt) IdentifyPeerWithMyIDChallengeAck(data *IdentifyPeer, peer *PttPeer) error {
	if p.myEntity == nil {
		return ErrInvalidEntity
	}

	myID := p.myEntity.GetID()
	if !reflect.DeepEqual(peer.IDEntityID, myID) {
		return nil
	}

	peerAckData, err := p.IdentifyPeerAck(data.Challenge, peer)
	if err != nil {
		return err
	}

	ackData := &IdentifyPeerWithMyIDChallengeAck{
		Challenge: peer.IDChallenge,
		AckData:   peerAckData,
	}

	return p.SendDataToPeer(CodeTypeIdentifyPeerWithMyIDChallengeAck, ackData, peer)
}

/*
HandleIdentifyPeerWithMyIDChallengeAck handles IdentifyPeerWithMyIDChallengeAck (acker)
*/
func (p *BasePtt) HandleIdentifyPeerWithMyIDChallengeAck(dataBytes []byte, peer *PttPeer) error {
	if p.myEntity == nil {
		return ErrInvalidEntity
	}

	myID := p.myEntity.GetID()

	data := &IdentifyPeerWithMyIDChallengeAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	err = p.HandleIdentifyPeerAck(myID, data.AckData, peer)
	if err != nil {
		return err
	}

	if peer.PeerType == PeerTypeRandom {
		return nil
	}

	return p.IdentifyPeerWithMyIDAck(data.Challenge, peer)
}
