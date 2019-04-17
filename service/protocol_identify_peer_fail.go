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

	"github.com/ethereum/go-ethereum/common"
)

type IdentifyPeerFail struct {
	Hash *common.Address `json:"H"`
}

/*
IdentifyPeerFail acks PMIdentifyPeer as failed
*/
func (p *BasePtt) IdentifyPeerFail(hash *common.Address, peer *PttPeer) error {
	data := &IdentifyPeerFail{
		Hash: hash,
	}

	return p.SendDataToPeer(CodeTypeIdentifyPeerFail, data, peer)
}

func (p *BasePtt) HandleIdentifyPeerFail(dataBytes []byte, peer *PttPeer) error {
	data := &IdentifyPeerFail{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	//hash := data.Hash
	//p.RemoveOpHash(hash)

	return p.IdentifyPeerWithMyID(peer)
}
