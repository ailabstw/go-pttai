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
	"github.com/ailabstw/go-pttai/log"
)

type OpCheckMemberAck struct {
	EntityID *types.PttID `json:"ID"`
	Log      *BaseOplog   `json:"l"`
}

func (p *BasePtt) OpCheckMemberAck(
	entityID *types.PttID,
	memberLog *BaseOplog,
	peer *PttPeer,
) error {

	data := &OpCheckMemberAck{
		EntityID: entityID,
		Log:      memberLog,
	}

	return p.SendDataToPeer(CodeTypeOpCheckMemberAck, data, peer)

}

func (p *BasePtt) HandleOpCheckMemberAck(dataBytes []byte, peer *PttPeer) error {

	data := &OpCheckMemberAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	entity, ok := p.entities[*data.EntityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	err = data.Log.Verify()
	log.Debug("HandleOpCheckMemberAck: after Verify", "e", err)
	if err != nil {
		return err
	}

	err = pm.HandleMemberOplogs([]*BaseOplog{data.Log}, peer, false)
	log.Debug("HandleOpCheckMemberAck: after HandleMemberOplogs", "e", err)
	return err
}
