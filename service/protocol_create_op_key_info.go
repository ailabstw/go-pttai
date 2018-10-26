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
)

func (pm *BaseProtocolManager) TryCreateOpKeyInfo() error {
	toRenewTS, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	toRenewTS.Ts -= pm.RenewOpKeySeconds()

	keyInfo, err := pm.GetNewestOpKey(false)
	if err == nil && toRenewTS.IsLess(keyInfo.UpdateTS) {
		return nil
	}

	return pm.CreateOpKeyInfo()
}

func (pm *BaseProtocolManager) CreateOpKeyInfo() error {
	ptt := pm.Ptt()
	entity := pm.Entity()

	entityID := pm.Entity().GetID()
	myEntity := ptt.MyEntity()
	myID := myEntity.GetID()

	// 1. validate
	if entity.GetStatus() != types.StatusAlive {
		return nil
	}

	if !pm.IsMaster(myID) {
		return nil
	}

	masterKey := myEntity.MasterKey()

	// 2. new-key (new item)
	keyInfo, err := NewOpKeyInfo(entityID, myID, masterKey)
	if err != nil {
		return err
	}

	// 3. new oplog
	opData := &OpKeyOpAddKey{}
	log, err := NewOpKeyOplog(entityID, keyInfo.UpdateTS, myID, OpKeyOpTypeAddKey, opData, pm.db, entityID, pm.dbOpKeyLock)
	if err != nil {
		return err
	}

	// 4. key-info (item set oplog and save)
	keyInfo.LogID = log.ID

	err = keyInfo.Save(pm.db, false)
	if err != nil {
		return err
	}

	// 5. sign oplog and save
	err = pm.SignOplog(log.Oplog)
	if err != nil {
		return err
	}

	err = log.Save(false)
	if err != nil {
		return err
	}

	// 6. broadcast oplog
	pm.BroadcastOpKeyOplog(log)

	// 7. postprocess
	err = pm.RegisterOpKeyInfo(keyInfo, false, false)
	if err != nil {
		return err
	}

	return nil
}
