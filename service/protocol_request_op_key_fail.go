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

	"github.com/ailabstw/go-pttai/common/types"
)

type OpKeyFail struct {
	EntityID *types.PttID `json:"ID"`
}

func (p *BasePtt) RequestOpKeyFail(entityID *types.PttID, peer *PttPeer) error {
	data := &OpKeyFail{
		EntityID: entityID,
	}
	return p.SendDataToPeer(CodeTypeRequestOpKeyFail, data, peer)
}

func (p *BasePtt) HandleRequestOpKeyFail(dataBytes []byte, peer *PttPeer) error {
	if peer.UserID == nil {
		return types.ErrInvalidID
	}

	data := &OpKeyFail{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return p.OpCheckMember(data.EntityID, peer)
}
