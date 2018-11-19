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
)

func (spm *BaseServiceProtocolManager) CreateJoinEntity(
	approveJoin *ApproveJoinEntity,
	peer *PttPeer,

	meLogID *types.PttID,
	isStart bool,
	isNew bool,
) (Entity, error) {

	entity, oplog0, masterLogs, memberLogs, opKey, opKeyLog := approveJoin.Entity, approveJoin.Oplog0, approveJoin.MasterLogs, approveJoin.MemberLogs, approveJoin.OpKey, approveJoin.OpKeyLog

	ptt := spm.Ptt()
	service := spm.Service()
	sspm := service.SPM()

	// entity
	entity.SetSyncInfo(nil)
	err := entity.Save(true)
	if err != nil {
		return nil, err
	}

	if isNew {
		err = entity.Init(ptt, service, sspm)
		if err != nil {
			return nil, err
		}
	}

	// master-logs
	pm := entity.PM()
	log.Debug("CreateJoinEntity: to HandleMasterOplogs", "entity", entity.GetID(), "masterLogs", len(masterLogs))
	for _, masterLog := range masterLogs {
		log.Debug("CreateJoinEntity: to HandleMasterOplogs", "entity", entity.GetID(), "masterLog", masterLog.ID, "master", masterLog.ObjID)
	}

	pm.HandleMasterOplogs(masterLogs, peer, false)

	// member-logs
	log.Debug("CreateJoinEntity: to HandleMemberOplogs", "entity", entity.GetID(), "memberLogs", len(memberLogs))
	for _, memberLog := range memberLogs {
		log.Debug("CreateJoinEntity: to HandleMemberOplogs", "entity", entity.GetID(), "memberLog", memberLog.ObjID)
	}
	pm.HandleMemberOplogs(memberLogs, peer, false)
	pm.SetMemberSyncTime(types.ZeroTimestamp)

	// register-master-peer again from
	masters, err := pm.GetMasterListFromCache(false)
	if err != nil {
		return nil, err
	}
	for _, master := range masters {
		pm.RegisterMaster(master, false, false)
	}

	// oplog0
	log.Debug("CreateJoinEntity: to SetLog0DB", "oplog0", oplog0)
	pm.SetLog0DB(oplog0)
	err = oplog0.Save(false)
	if err != nil {
		return nil, err
	}

	// op-key
	pm.SetOpKeyObjDB(opKey)
	err = opKey.Save(false)
	if err != nil {
		return nil, err
	}
	pm.SetOpKeyDB(opKeyLog)
	err = opKeyLog.Save(false)
	if err != nil {
		return nil, err
	}
	err = pm.RegisterOpKey(opKey, false)
	log.Debug("CreateJoinEntity: after register op key", "e", err, "entity", pm.Entity().GetID(), "opKey", opKey.Hash)
	if err != nil {
		return nil, err
	}

	// spm-register
	spm.RegisterEntity(entity.GetID(), entity)

	if isStart {
		entity.PrestartAndStart()
	}

	// me-oplog
	if meLogID != nil {
		return entity, nil
	}

	if entity.GetEntityType() == EntityTypePersonal {
		return entity, nil
	}

	err = ptt.GetMyEntity().CreateJoinEntityOplog(entity)
	log.Debug("CreateJoinEntity: after CreateJoinEntityOplog", "e", err)
	if err != nil {
		return nil, err
	}

	return entity, nil
}
