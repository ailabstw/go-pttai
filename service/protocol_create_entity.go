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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (spm *BaseServiceProtocolManager) CreateEntity(
	data CreateData,
	createOp OpType,

	newEntity func(data CreateData, ptt Ptt, service Service) (Entity, OpData, error),
	newOplogWithTS func(objID *types.PttID, ts types.Timestamp, op OpType, opData OpData) (Oplog, error),

	increate func(entity Entity, oplog *BaseOplog, opData OpData) error,

	postcreateEntity func(entity Entity, oplog *BaseOplog) error,
) (Entity, error) {

	myID := spm.Ptt().GetMyEntity().GetID()

	entity, opData, err := newEntity(data, spm.Ptt(), spm.Service())
	if err != nil {
		return nil, err
	}
	err = entity.Lock()
	if err != nil {
		return nil, err
	}
	defer entity.Unlock()

	pm := entity.PM()
	entityID := entity.GetID()
	ts := entity.GetUpdateTS()

	// master
	_, _, err = pm.AddMaster(myID, true)
	log.Debug("CreateEntity: after AddMaster", "e", err)
	if err != nil {
		return nil, err
	}

	// member
	_, _, err = pm.AddMember(myID, true)
	log.Debug("CreateEntity: after AddMember", "e", err)
	if err != nil {
		return nil, err
	}

	// oplog
	theOplog, err := newOplogWithTS(entityID, ts, createOp, opData)
	if err != nil {
		return nil, err
	}
	oplog := theOplog.GetBaseOplog()
	oplog.dbPrefixID = entityID

	// in-create
	if increate != nil {
		err = increate(entity, oplog, opData)
		if err != nil {
			return nil, err
		}
	}

	// sign oplog
	err = pm.SignOplog(oplog)
	if err != nil {
		return nil, err
	}

	// save entity
	origStatus := entity.GetStatus()

	err = pm.SaveNewEntityWithOplog(oplog, true, true)
	if err != nil {
		return nil, err
	}

	// add to entities
	err = spm.RegisterEntity(entityID, entity)
	if err != nil {
		return nil, err
	}

	// oplog save
	err = oplog.Save(false)
	if err != nil {
		return nil, err
	}

	// op-key, required entity to be alive to generate op-key
	err = pm.CreateOpKey()
	log.Debug("CreateEntity: after CreateOpKeyInfo", "e", err)
	if err != nil {
		return nil, err
	}

	// entity start
	err = entity.PrestartAndStart()
	if err != nil {
		return nil, err
	}

	err = pm.MaybePostcreateEntity(oplog, origStatus, true, postcreateEntity)

	return entity, nil
}

/**********
 * PM functions. Requiring public funcions to let SPM able to access.
 **********/

func (pm *BaseProtocolManager) SaveNewEntityWithOplog(oplog *BaseOplog, isLocked bool, isForce bool) error {

	entity := pm.Entity()

	origStatus := entity.GetStatus()
	status := oplog.ToStatus()

	if !isForce && origStatus >= status && !(origStatus == types.StatusFailed && status == types.StatusAlive) {
		return nil
	}

	entity.SetStatus(status)
	entity.SetUpdateTS(oplog.UpdateTS)
	err := entity.Save(isLocked)
	if err == pttdb.ErrInvalidUpdateTS {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) MaybePostcreateEntity(
	oplog *BaseOplog,
	origStatus types.Status,
	isForce bool,
	postcreateEntity func(entity Entity, oplog *BaseOplog) error,
) error {
	if postcreateEntity == nil {
		return nil
	}

	entity := pm.Entity()

	status := oplog.ToStatus()

	if !isForce && origStatus >= types.StatusAlive && origStatus != types.StatusFailed {
		// orig-status is already alive
		return nil
	}

	if status != types.StatusAlive {
		return nil
	}

	return postcreateEntity(entity, oplog)
}
