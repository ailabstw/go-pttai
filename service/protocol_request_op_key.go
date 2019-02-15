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
	"github.com/ailabstw/go-pttai/pttdb"
)

type RequestOpKey struct {
	EntityID  *types.PttID `json:"ID"`
	OpKeys    []*KeyInfo   `json:"o"`
	OpKeyLogs []*BaseOplog `json:"O"`
}

func (p *BasePtt) RequestOpKey(hash *common.Address, peer *PttPeer) error {

	if peer.UserID == nil {
		return types.ErrInvalidID
	}

	entity, err := p.getEntityFromHash(hash, &p.lockOps, p.ops)
	if err != nil {
		log.Error("RequestOpKey: getEntityFromHash", "e", err)
		return err
	}

	opKeys := entity.PM().OpKeyList()

	opKeyOplogs, err := entity.PM().GetOpKeyOplogList(nil, 0, pttdb.ListOrderNext, types.StatusAlive)
	if err != nil {
		log.Error("RequestOpKey: unable to get OpKeyOplogList", "e", err)
		return err
	}
	oplogs := OpKeyOplogsToOplogs(opKeyOplogs)

	data := &RequestOpKey{
		EntityID:  entity.GetID(),
		OpKeyLogs: oplogs,
		OpKeys:    opKeys,
	}
	log.Debug("RequestOpKey: to SendDataToPeer", "entity", entity.GetID(), "service", entity.Service().Name(), "opKeys", opKeys)

	return p.SendDataToPeer(CodeTypeRequestOpKey, data, peer)
}

func (p *BasePtt) HandleRequestOpKey(dataBytes []byte, peer *PttPeer) error {

	if peer.UserID == nil {
		return types.ErrInvalidID
	}

	data := &RequestOpKey{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	entity, ok := p.entities[*data.EntityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	if entity.GetStatus() >= types.StatusMigrated {
		return p.EntityTerminal(entity, pm, peer)
	}

	err = pm.HandleOpKeyOplogs(data.OpKeyLogs, peer, false)

	log.Debug("HandleRequestOpKey: after HandleOpKeyOplogs", "e", err, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	for _, opKey := range data.OpKeys {
		pm.HandleSyncCreateOpKeyAckObj(opKey, peer)
	}

	return p.RequestOpKeyAck(data.EntityID, peer)
}
