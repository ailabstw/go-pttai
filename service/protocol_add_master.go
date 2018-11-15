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

func (pm *BaseProtocolManager) AddMaster(id *types.PttID, isForce bool) (*Master, *MasterOplog, error) {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()
	origMaster := NewEmptyMaster()
	pm.SetMasterObjDB(origMaster)

	// 1. validate
	if !isForce && !pm.IsMaster(myID, false) {
		return nil, nil, types.ErrInvalidID
	}

	if len(pm.masters) >= pm.maxMasters {
		return nil, nil, ErrTooManyMasters
	}

	data := &MasterOpCreateMaster{}
	person, oplog, err := pm.AddPerson(
		id, MasterOpTypeAddMaster, isForce,
		origMaster, data,
		pm.NewMaster, pm.NewMasterOplogWithTS, pm.broadcastMasterOplogCore, pm.postaddMaster,
		pm.SetMasterDB, pm.NewMasterOplog,
	)
	if err != nil {
		return nil, nil, err
	}
	master, ok := person.(*Master)
	if !ok {
		return nil, nil, ErrInvalidObject
	}

	masterOplog := &MasterOplog{BaseOplog: oplog}

	return master, masterOplog, nil
}

func (pm *BaseProtocolManager) NewMaster(id *types.PttID) (Object, OpData, error) {
	entity := pm.Entity()
	myEntity := pm.Ptt().GetMyEntity()
	myID := myEntity.GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	log.Debug("NewMaster: to NewMaster", "ts", ts)
	master := NewMaster(id, ts, myID, entity.GetID(), nil, types.StatusInternalPending, pm.DB(), pm.DBObjLock(), pm.dbMasterPrefix, pm.dbMasterIdxPrefix)
	log.Debug("NewMaster: after NewMaster", "master", master)

	return master, &MasterOpCreateMaster{}, nil
}

func (pm *BaseProtocolManager) postaddMaster(theMaster Object, oplog *BaseOplog) error {
	master, ok := theMaster.(*Master)
	if !ok {
		return ErrInvalidData
	}

	err := pm.SetNewestMasterLogID(oplog.ID)
	if err != nil {
		return err
	}

	err = pm.SetMasterLog0Hash(oplog.Hash)
	if err != nil {
		return err
	}

	pm.RegisterMaster(master, false, false)

	return nil
}
