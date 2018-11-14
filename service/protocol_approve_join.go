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

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type ApproveJoin struct {
	ID         *types.PttID
	Name       []byte          `json:"N"`
	OpKeyBytes []byte          `json:"K"`
	Data       ApproveJoinData `json:"D"`
}

type ApproveJoinData interface{}

/*
ApproveJoin approves join (invitor)
*/
func (p *BasePtt) ApproveJoin(confirmKey []byte) error {
	p.lockConfirmJoin.Lock()
	defer p.lockConfirmJoin.Unlock()

	confirmKeyStr := string(confirmKey)

	confirmJoin, ok := p.confirmJoins[confirmKeyStr]
	if !ok {
		return ErrInvalidKey
	}

	entity, joinEntity, keyInfo, peer := confirmJoin.Entity, confirmJoin.JoinEntity, confirmJoin.KeyInfo, confirmJoin.Peer

	pm := entity.PM()
	opKeyInfo, approvedData, err := pm.ApproveJoin(joinEntity, keyInfo, peer)
	log.Debug("ApproveJoin: after pm.ApproveJoin", "e", err)
	if err != nil {
		return err
	}

	id := entity.GetID()
	name := entity.Name()
	opKeyBytes := opKeyInfo.KeyBytes

	approveJoin := &ApproveJoin{
		ID:         id,
		Name:       []byte(name),
		OpKeyBytes: opKeyBytes,
		Data:       approvedData,
	}

	data, err := json.Marshal(approveJoin)
	if err != nil {
		return err
	}

	encData, err := p.EncryptData(ApproveJoinMsg, data, keyInfo)
	if err != nil {
		return err
	}

	pttData, err := p.MarshalData(CodeTypeJoinAck, keyInfo.Hash, encData)
	if err != nil {
		return err
	}

	log.Debug("ApproveJoin: to send to peer", "peer", peer)

	pttData.Node = peer.GetID()[:]
	err = peer.SendData(pttData)

	delete(p.confirmJoins, confirmKeyStr)

	return nil
}

func (p *BasePtt) HandleApproveJoin(dataBytes []byte, hash *common.Address, joinRequest *JoinRequest, peer *PttPeer) error {
	if joinRequest.Status != JoinStatusWaitAccepted {
		return ErrInvalidData
	}

	return p.myEntity.HandleApproveJoin(dataBytes, hash, joinRequest, peer)
}
