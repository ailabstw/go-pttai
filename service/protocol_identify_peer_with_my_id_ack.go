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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
)

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
