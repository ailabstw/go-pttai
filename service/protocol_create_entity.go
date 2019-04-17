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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

/*
CreateEntity creates the entity. CreateEntity is from SPM, while CreateObject and CreateMember is from PM.

	1. new entity.
	2. add master.
	2.1. add member.
	3. oplog.
	4. increate.
	5. force-sign oplog.
	6. entity-save.
	6.1. SPM register entity.
	7. oplog-save.
	8. post-save
	8.1. create op-key.
	8.2. prestart and start.
	8.3. postcreate.
*/
func (spm *BaseServiceProtocolManager) CreateEntity(
	data CreateData,
	createOp OpType,

	newEntity func(data CreateData, ptt Ptt, service Service) (Entity, OpData, error),
	newOplogWithTS func(objID *types.PttID, ts types.Timestamp, op OpType, opData OpData) (Oplog, error),

	increate func(entity Entity, oplog *BaseOplog, opData OpData) error,

	postcreate func(entity Entity) error,
) (Entity, error) {

	// 1. new entity
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

	// 2. master
	_, _, err = pm.AddMaster(myID, true)
	log.Debug("CreateEntity: after AddMaster", "e", err)
	if err != nil {
		return nil, err
	}

	// 2.1. member
	_, _, err = pm.AddMember(myID, true)
	log.Debug("CreateEntity: after AddMember", "e", err)
	if err != nil {
		return nil, err
	}

	// 3. oplog
	theOplog, err := newOplogWithTS(entityID, ts, createOp, opData)
	log.Debug("CreateEntity: after newOplogWithTS", "e", err)
	if err != nil {
		return nil, err
	}
	oplog := theOplog.GetBaseOplog()
	oplog.dbPrefixID = entityID

	// 4. in-create
	if increate != nil {
		err = increate(entity, oplog, opData)
		if err != nil {
			return nil, err
		}
	}

	// 5. sign oplog
	err = pm.ForceSignOplog(oplog)
	if err != nil {
		return nil, err
	}

	// 6. save entity
	err = pm.SaveNewEntityWithOplog(oplog, true, true)
	if err != nil {
		return nil, err
	}

	// 6.1. add to entities
	err = spm.RegisterEntity(entityID, entity)
	log.Debug("CreateEntity: after RegisterEntity", "e", err, "service", spm.Service().Name())
	if err != nil {
		return nil, err
	}

	// 7. oplog save
	oplog.IsSync = true
	err = oplog.Save(false, pm.Log0Merkle())
	if err != nil {
		return nil, err
	}

	// 8. postsave
	// 8.1. op-key, required entity to be alive to generate op-key
	err = pm.ForceCreateOpKey()
	log.Debug("CreateEntity: after CreateOpKeyInfo", "e", err)
	if err != nil {
		return nil, err
	}

	// 8.2. entity start
	err = entity.PrestartAndStart()
	log.Debug("CreateEntity: after entity Prestart and start", "e", err)
	if err != nil {
		return nil, err
	}

	// 8.3. postcreate
	err = pm.MaybePostcreateEntity(oplog, false, postcreate)

	return entity, nil
}

/**********
 * PM functions. Requiring public funcions to let SPM able to access.
 **********/

/*
SaveNewEntityWithOplog sets and saves the status/UT of the newly created entity based on the oplog.
*/
func (pm *BaseProtocolManager) SaveNewEntityWithOplog(oplog *BaseOplog, isLocked bool, isForce bool) error {

	entity := pm.Entity()

	var err error
	if !isLocked {
		err = entity.Lock()
		if err != nil {
			return err
		}
		defer entity.Unlock()
	}

	SetNewEntityWithOplog(entity, oplog.ToStatus(), oplog)

	err = entity.Save(true)
	if err != nil {
		return err
	}

	return nil
}

/*
MaybePostcreateEntity checks the whether to do postcreate and does postcreate of the entity.
*/
func (pm *BaseProtocolManager) MaybePostcreateEntity(
	oplog *BaseOplog,

	isForceNot bool,
	postcreate func(entity Entity) error,
) error {
	if postcreate == nil {
		return nil
	}

	entity := pm.Entity()

	status := oplog.ToStatus()

	if status != types.StatusAlive {
		return nil
	}

	return postcreate(entity)
}
