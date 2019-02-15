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

type EntityDeleted struct {
	EntityID  *types.PttID `json:"ID"`
	EntityLog *BaseOplog   `json:"E"`
	Status    types.Status `json:"S"`
}

func (p *BasePtt) EntityTerminal(entity Entity, pm ProtocolManager, peer *PttPeer) error {

	entityLog, _ := pm.GetEntityLog()

	data := &EntityDeleted{
		EntityID:  entity.GetID(),
		EntityLog: entityLog,
		Status:    entity.GetStatus(),
	}

	p.SendDataToPeer(CodeTypeEntityDeleted, data, peer)

	return nil
}

func (p *BasePtt) HandleEntityTerminal(dataBytes []byte, peer *PttPeer) error {
	if peer.UserID == nil {
		return types.ErrInvalidID
	}

	data := &EntityDeleted{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	entity, ok := p.entities[*data.EntityID]
	if !ok {
		return types.ErrInvalidID
	}
	pm := entity.PM()

	pm.HandleEntityTerminal(data.Status, data.EntityLog, peer)

	return nil
}

/**********
 * pm
 **********/

func (pm *BaseProtocolManager) HandleEntityTerminal(status types.Status, entityLog *BaseOplog, peer *PttPeer) error {

	if !pm.IsMember(peer.UserID, false) {
		return nil
	}

	if status < types.StatusMigrated {
		return nil
	}

	pm.HandleLog0s([]*BaseOplog{entityLog}, peer, false)

	return nil
}
