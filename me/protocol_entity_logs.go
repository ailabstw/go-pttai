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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleEntityLog(
	oplog *pkgservice.BaseOplog,

	spm pkgservice.ServiceProtocolManager,
	opData *MeOpEntity,

	info *ProcessMeInfo,

	updateCreateEntityInfo func(oplog *pkgservice.BaseOplog, info *ProcessMeInfo),
) ([]*pkgservice.BaseOplog, error) {

	entity := spm.Entity(oplog.ObjID)
	log.Debug("HandleEntityLog: after get Entity", "entityID", oplog.ObjID, "entity", entity, "isNewer", oplog.IsNewer)
	if entity == nil {
		if oplog.IsNewer {
			return nil, nil
		}
		updateCreateEntityInfo(oplog, info)
		return nil, nil
	}

	// 1. lock
	err := entity.Lock()
	if err != nil {
		return nil, err
	}
	defer entity.Unlock()

	// already same log-id.
	if reflect.DeepEqual(oplog.ID, entity.GetMeLogID()) {
		oplog.IsSync = true
		return nil, nil
	}

	// newer oplog
	if oplog.UpdateTS.IsLess(entity.GetMeLogTS()) {
		oplog.IsSync = true
		return nil, nil
	}

	// already alive
	if entity.GetStatus() == types.StatusAlive {
		entity.SetMeLogTS(oplog.UpdateTS)
		entity.SetMeLogID(oplog.ID)
		entity.Save(true)

		oplog.IsSync = true

		return nil, nil
	}

	// update create entity info
	if oplog.IsNewer {
		return nil, nil
	}
	updateCreateEntityInfo(oplog, info)

	return nil, nil
}

func (pm *ProtocolManager) SetNewestEntityLog(
	oplog *pkgservice.BaseOplog,

	spm pkgservice.ServiceProtocolManager,
) (types.Bool, error) {

	// possibly already deleted.
	entity := spm.Entity(oplog.ObjID)
	if entity == nil {
		return true, nil
	}

	return types.Bool(reflect.DeepEqual(oplog.ID, entity.GetMeLogID())), nil
}
