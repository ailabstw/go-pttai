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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type IdentifyPeerWithMyID struct {
	ID        *types.PttID
	Challenge *types.Salt `json:"C"`
}

/*
IdentifyPeerWithMyID ask for identifying peer with providing my-id (requester)
*/
func (p *BasePtt) IdentifyPeerWithMyID(peer *PttPeer) error {
	if p.myEntity == nil {
		return ErrInvalidEntity
	}

	myID := p.myEntity.GetID()

	salt, err := peer.InitID(myID, p.quitSync, true)
	if err != nil {
		return err
	}

	data := &IdentifyPeerWithMyID{
		ID:        myID,
		Challenge: salt,
	}

	peer.IDEntityID = nil

	log.Debug("IdentifyPeerWithMyID: to SendDataToPeer", "peer", peer)

	return p.SendDataToPeer(CodeTypeIdentifyPeerWithMyID, data, peer)
}

func (p *BasePtt) HandleIdentifyPeerWithMyID(dataBytes []byte, peer *PttPeer) error {
	data := &IdentifyPeerWithMyID{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	if peer.UserID != nil {
		if !reflect.DeepEqual(peer.UserID, data.ID) {
			return types.ErrInvalidID
		}

		return p.IdentifyPeerWithMyIDAck(data.Challenge, peer)
	}

	log.Debug("HandleIdentifyPeerWithMyID: to IdentifyPeerWithMyIDChallenge", "peer", peer, "data", data)

	return p.IdentifyPeerWithMyIDChallenge(data.ID, peer)
}
