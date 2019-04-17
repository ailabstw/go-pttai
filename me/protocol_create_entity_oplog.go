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

package me

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (m *MyInfo) CreateEntityOpType(entity pkgservice.Entity) (pkgservice.OpType, error) {

	switch entity.(type) {
	case *content.Board:
		return MeOpTypeCreateBoard, nil
	case *friend.Friend:
		return MeOpTypeCreateFriend, nil
	}
	return MeOpTypeInvalid, pkgservice.ErrInvalidEntity
}

func (m *MyInfo) CreateEntityOplog(entity pkgservice.Entity) error {
	if entity.GetEntityType() == pkgservice.EntityTypePersonal {
		return nil
	}

	op, err := m.CreateEntityOpType(entity)
	if err != nil {
		return err
	}

	pm := m.PM().(*ProtocolManager)

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	entityID := entity.GetID()
	oplog, err := pm.CreateMeOplog(entityID, ts, op, &MeOpEntity{LogID: entity.GetLogID()})
	if err != nil {
		return err
	}

	entity.SetMeLogID(oplog.ID)
	entity.SetMeLogTS(oplog.UpdateTS)
	entity.Save(true)

	oplog.IsSync = true

	err = oplog.Save(false, pm.meOplogMerkle)
	if err != nil {
		return err
	}

	pm.BroadcastMeOplog(oplog)

	return nil
}
