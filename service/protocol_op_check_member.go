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
)

type OpCheckMember struct {
	EntityID *types.PttID `json:"ID"`
	MyLog    *BaseOplog   `json:"m"`
	TheirLog *BaseOplog   `json:"t"`
}

func (p *BasePtt) OpCheckMember(entityID *types.PttID, peer *PttPeer) error {

	if peer.UserID == nil {
		return types.ErrInvalidID
	}

	entity, ok := p.entities[*entityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	myID := p.GetMyEntity().GetID()
	myMemberLog, err := pm.GetMemberLogByMemberID(myID, false)
	if err != nil {
		return err
	}

	peerMemberLog, err := pm.GetMemberLogByMemberID(peer.UserID, false)
	if err != nil {
		return err
	}

	data := &OpCheckMember{
		EntityID: entity.GetID(),
		MyLog:    myMemberLog.BaseOplog,
		TheirLog: peerMemberLog.BaseOplog,
	}

	return p.SendDataToPeer(CodeTypeOpCheckMember, data, peer)
}

func (p *BasePtt) HandleOpCheckMember(dataBytes []byte, peer *PttPeer) error {

	data := &OpCheckMember{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	entity, ok := p.entities[*data.EntityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	err = data.MyLog.Verify()
	if err != nil {
		return err
	}

	err = data.TheirLog.Verify()
	if err != nil {
		return err
	}

	// theirID

	if entity.GetStatus() <= types.StatusAlive {
		theirMemberLog, err := pm.GetMemberLogByMemberID(peer.UserID, false)
		if err != nil {
			return err
		}

		if reflect.DeepEqual(theirMemberLog.Hash, data.MyLog.Hash) {
			err = nil
		} else if theirMemberLog.UpdateTS.IsLess(data.MyLog.UpdateTS) {
			err = pm.HandleMemberOplogs([]*BaseOplog{data.MyLog}, peer, false)
		} else {
			err = p.OpCheckMemberAck(data.EntityID, theirMemberLog.BaseOplog, peer)
		}
		if err != nil {
			return err
		}
	}

	// myID
	myID := p.GetMyEntity().GetID()
	myMemberLog, err := pm.GetMemberLogByMemberID(myID, false)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(myMemberLog.Hash, data.TheirLog.Hash) {
		err = nil
	} else if myMemberLog.UpdateTS.IsLess(data.TheirLog.UpdateTS) {
		if entity.GetStatus() <= types.StatusAlive {
			err = pm.HandleMemberOplogs([]*BaseOplog{data.TheirLog}, peer, false)
		} else {
			err = types.ErrNotImplemented
		}
	} else {
		err = p.OpCheckMemberAck(data.EntityID, myMemberLog.BaseOplog, peer)
	}
	if err != nil {
		return err
	}

	return nil
}
