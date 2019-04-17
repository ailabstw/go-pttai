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
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) CreateOpKeyLoop() error {
	log.Debug("CreateOpKeyLoop: start", "entity", pm.Entity().IDString())
	err := pm.TryCreateOpKeyInfo()
	log.Debug("CreateOpKeyLoop: after 1st TryCreateOpKeyInfo", "e", err, "entity", pm.Entity().IDString())

	toRenewSeconds := pm.GetToRenewOpKeySeconds()
	log.Debug("CreateOpKeyLoop: after getToRenewSeconds", "toRenewSeconds", toRenewSeconds, "entity", pm.Entity().IDString())
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
		case <-pm.ForceOpKey():
			log.Debug("CreateOpKeyLoop: ForceOpKey", "entity", pm.Entity().IDString())
			pm.CreateOpKey()
		case <-pm.QuitSync():
			log.Debug("CreateOpKeyLoop: QuitSync", "entity", pm.Entity().IDString())
			break loop
		}
	}

	ticker.Stop()

	return nil
}

func (pm *BaseProtocolManager) GetToRenewOpKeySeconds() int {
	entity := pm.Entity()

	if entity.GetStatus() < types.StatusToBeSynced {
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

	return pm.CreateOpKey()
}

func (pm *BaseProtocolManager) ToRenewOpKeyTS() (types.Timestamp, error) {
	toRenewTS, err := types.GetTimestamp()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	toRenewTS.Ts -= pm.RenewOpKeySeconds()

	return toRenewTS, nil
}

/***
 * CreateObject
 ***/
func (pm *BaseProtocolManager) CreateOpKey() error {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	// 1. validate
	if !pm.IsMaster(myID, false) {
		log.Warn("CreateOpKeyInfo: not master", "entity", pm.Entity().IDString())
		return nil
	}

	// 2. create object
	_, err := pm.ForceCreateObject(
		nil,
		OpKeyOpTypeCreateOpKey,

		nil,

		pm.NewOpKey,
		pm.NewOpKeyOplogWithTS,
		nil,

		pm.SetOpKeyDB,
		pm.broadcastOpKeyOplogsCore,
		pm.broadcastOpKeyOplogCore,

		pm.postcreateOpKey,
	)
	log.Debug("CreateOpKey: done", "e", err, "entity", pm.Entity().IDString())
	if err != nil {
		log.Warn("CreateOpKeyInfo: unable to CreateObj", "e", err, "entity", pm.Entity().IDString())
		return err
	}
	return nil
}

func (pm *BaseProtocolManager) ForceCreateOpKey() error {
	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	// 1. validate
	if !pm.IsMaster(myID, false) {
		log.Warn("CreateOpKeyInfo: not master")
		return nil
	}

	// 2. create object
	_, err := pm.ForceCreateObject(
		nil,
		OpKeyOpTypeCreateOpKey,

		nil,

		pm.NewOpKey,
		pm.NewOpKeyOplogWithTS,
		nil,

		pm.SetOpKeyDB,
		pm.broadcastOpKeyOplogsCore,
		pm.broadcastOpKeyOplogCore,

		pm.postcreateOpKey,
	)
	if err != nil {
		log.Warn("CreateOpKeyInfo: unable to CreateObj", "e", err)
		return err
	}
	return nil
}

func (pm *BaseProtocolManager) NewOpKey(data CreateData) (Object, OpData, error) {
	entity := pm.Entity()
	myEntity := pm.Ptt().GetMyEntity()

	keyInfo, err := myEntity.NewOpKeyInfo(entity.GetID(), pm.SetOpKeyObjDB)
	if err != nil {
		return nil, nil, err
	}

	return keyInfo, &OpKeyOpCreateOpKey{}, nil
}

func (pm *BaseProtocolManager) postcreateOpKey(theOpKey Object, oplog *BaseOplog) error {
	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	pm.RegisterOpKey(opKey, false)

	return nil
}

func (pm *BaseProtocolManager) ForceOpKey() chan struct{} {
	return pm.forceOpKey
}
