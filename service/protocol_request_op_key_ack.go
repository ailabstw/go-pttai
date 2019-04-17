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
	"github.com/ailabstw/go-pttai/pttdb"
)

type RequestOpKeyAck struct {
	EntityID *types.PttID `json:"ID"`
	OpKeys   []*KeyInfo   `json:"k"`
	Logs     []*BaseOplog `json:"l"`
}

func (p *BasePtt) RequestOpKeyAck(entityID *types.PttID, peer *PttPeer) error {
	entity, ok := p.entities[*entityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	peerType := pm.GetPeerType(peer)
	if peerType < PeerTypeMember {
		return p.RequestOpKeyFail(entityID, peer)
	}

	opKeys := pm.OpKeyList()

	opKeyLogs, err := pm.GetOpKeyOplogList(nil, 0, pttdb.ListOrderNext, types.StatusAlive)
	if err != nil {
		return err
	}
	oplogs := OpKeyOplogsToOplogs(opKeyLogs)

	data := &RequestOpKeyAck{
		EntityID: entityID,
		OpKeys:   opKeys,
		Logs:     oplogs,
	}

	return p.SendDataToPeer(CodeTypeRequestOpKeyAck, data, peer)
}

func (p *BasePtt) HandleRequestOpKeyAck(dataBytes []byte, peer *PttPeer) error {

	if peer.UserID == nil {
		return types.ErrInvalidID
	}

	data := &RequestOpKeyAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	entity, ok := p.entities[*data.EntityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	pm.HandleOpKeyOplogs(data.Logs, peer, false)

	for _, opKey := range data.OpKeys {
		pm.HandleSyncCreateOpKeyAckObj(opKey, peer)
	}

	return nil
}
