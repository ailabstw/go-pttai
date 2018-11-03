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
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) CreateOpKeyLoop() error {
	log.Debug("CreateOpKeyLoop: start")
	err := pm.TryCreateOpKeyInfo()
	log.Debug("CreateOpKeyLoop: after 1st TryCreateOpKeyInfo", "e", err)
	if err != nil {
		return nil
	}

	toRenewSeconds := pm.GetToRenewOpKeySeconds()
	log.Debug("CreateOpKeyLoop: after getToRenewSeconds", "toRenewSeconds", toRenewSeconds)
	ticker := time.NewTimer(time.Duration(toRenewSeconds) * time.Second)

loop:
	for {
		select {
		case <-ticker.C:
			ticker.Stop()

			pm.TryCreateOpKeyInfo()

			toRenewSeconds = pm.GetToRenewOpKeySeconds()
			log.Debug("CreateOpKeyLoop: after getToRenewSeconds", "toRenewSeconds", toRenewSeconds)
			ticker = time.NewTimer(time.Duration(toRenewSeconds) * time.Second)
		case <-pm.QuitSync():
			break loop
		}
	}

	ticker.Stop()

	return nil
}

func (pm *BaseProtocolManager) GetToRenewOpKeySeconds() int {
	entity := pm.Entity()

	if entity.GetStatus() < types.StatusAlive {
		return 5
	}

	renewSeconds := pm.RenewOpKeySeconds()
	minRenewSeconds := renewSeconds / 2

	// XXX int issue
	return randNum(int(minRenewSeconds), int(renewSeconds))
}

func (pm *BaseProtocolManager) TryCreateOpKeyInfo() error {
	toRenewTS, err := pm.ToRenewOpKeyTS()
	if err != nil {
		log.Warn("TryCreateOpKeyInfo: unable to get toRenewTS", "e", err)
		return err
	}

	keyInfo, err := pm.GetNewestOpKey(false)
	if err == nil && toRenewTS.IsLess(keyInfo.UpdateTS) {
		return nil
	}

	return pm.CreateOpKeyInfo()
}

func (pm *BaseProtocolManager) ToRenewOpKeyTS() (types.Timestamp, error) {
	toRenewTS, err := types.GetTimestamp()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	toRenewTS.Ts -= pm.RenewOpKeySeconds()

	return toRenewTS, nil
}

func (pm *BaseProtocolManager) CreateOpKeyInfo() error {
	ptt := pm.Ptt()
	entity := pm.Entity()

	entityID := pm.Entity().GetID()
	myEntity := ptt.GetMyEntity()
	myID := myEntity.GetID()

	// 1. validate
	if entity.GetStatus() != types.StatusAlive {
		log.Warn("CreateOpKeyInfo: status not alive", "status", entity.GetStatus())
		return nil
	}

	if !pm.IsMaster(myID) {
		log.Warn("CreateOpKeyInfo: not master")
		return nil
	}

	// 2. new-key (new item)
	keyInfo, err := myEntity.NewOpKeyInfo(entityID, pm.DBOpKeyInfo(), pm.DBObjLock())
	if err != nil {
		log.Warn("CreateOpKeyInfo: unable to NewOpKeyInfo")
		return err
	}

	// 3. new oplog
	opData := &OpKeyOpCreateOpKey{}
	oplog, err := NewOpKeyOplog(keyInfo.ID, keyInfo.UpdateTS, myID, OpKeyOpTypeCreateOpKey, opData, pm.DBOpKeyInfo(), entityID, pm.DBOpKeyLock())
	if err != nil {
		return err
	}

	err = pm.SignOplog(oplog.BaseOplog)
	if err != nil {
		return err
	}

	// 4. key-info (item set oplog and save)
	keyInfo.LogID = oplog.ID
	keyInfo.UpdateTS = oplog.UpdateTS
	keyInfo.Status = oplog.ToStatus()

	err = keyInfo.Save(false)
	if err != nil {
		log.Warn("CreateOpKeyInfo: unable to save")
		return err
	}

	// 5. sign oplog and save

	err = oplog.Save(false)
	if err != nil {
		return err
	}

	// 6. broadcast oplog
	pm.BroadcastOpKeyOplog(oplog)

	// 7. postprocess
	if oplog.MasterLogID == nil {
		return nil
	}

	err = pm.RegisterOpKeyInfo(keyInfo, false)
	if err != nil {
		log.Warn("CreateOpKeyInfo: unable to register")
		return err
	}

	return nil
}
