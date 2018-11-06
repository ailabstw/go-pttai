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
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) CreateOpKeyLoop() error {
	log.Debug("CreateOpKeyLoop: start")
	err := pm.TryCreateOpKeyInfo()
	log.Debug("CreateOpKeyLoop: after 1st TryCreateOpKeyInfo", "e", err)

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
	myID := ptt.GetMyEntity().GetID()

	log.Debug("CreateOpKeyInfo: start")

	// 1. validate
	if !pm.IsMaster(myID) {
		log.Warn("CreateOpKeyInfo: not master")
		return nil
	}

	_, err := pm.CreateObject(nil, OpKeyOpTypeCreateOpKey, pm.NewOpKey, pm.NewOpKeyOplogWithTS, nil, pm.broadcastOpKeyOplogCore, pm.postcreateOpKey)
	if err != nil {
		log.Warn("CreateOpKeyInfo: unable to CreateObj", "e", err)
		return err
	}

	log.Debug("CreateOpKeyInfo: done")

	return nil
}

func (pm *BaseProtocolManager) NewOpKey(data interface{}) (Object, OpData, error) {
	entity := pm.Entity()
	myEntity := pm.Ptt().GetMyEntity()

	key_info, err := myEntity.NewOpKeyInfo(entity.GetID(), pm.DBOpKeyInfo(), pm.DBObjLock())
	if err != nil {
		return nil, nil, err
	}

	return key_info, &OpKeyOpCreateOpKey{}, nil
}

func (pm *BaseProtocolManager) postcreateOpKey(theOpKey Object, oplog *BaseOplog) error {
	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	pm.RegisterOpKeyInfo(opKey, false)

	return nil
}

func (pm *BaseProtocolManager) postfailedCreateOpKey(theOpKey Object, oplog *BaseOplog) error {
	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	return pm.RemoveOpKeyInfoFromHash(opKey.Hash, false, true, true)
}

func (k *KeyInfo) UpdateCreateInfo(oplog *BaseOplog, theOpData OpData, theInfo ProcessInfo) error {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return ErrInvalidData
	}

	info.CreateOpKeyInfo[*oplog.ObjID] = oplog

	return nil
}

func (pm *BaseProtocolManager) CreateOpKeyExistsInInfo(oplog *BaseOplog, theInfo ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return false, ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.DeleteOpKeyInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (k *KeyInfo) UpdateCreateObject(theObj Object) error {
	obj, ok := theObj.(*KeyInfo)
	if !ok {
		return ErrInvalidObject
	}

	key, err := crypto.ToECDSA(obj.KeyBytes)
	if err != nil {
		return err
	}

	origDB, origDBLock := k.db, k.dbLock
	k.BaseObject = obj.BaseObject
	k.db = origDB
	k.dbLock = origDBLock

	k.Hash = obj.Hash
	k.Key = key
	k.KeyBytes = obj.KeyBytes
	k.PubKeyBytes = crypto.FromECDSAPub(&key.PublicKey)
	k.Extra = obj.Extra

	return nil
}

func (k *KeyInfo) NewObjWithOplog(oplog *BaseOplog, theOpData OpData) error {
	NewObjectWithOplog(k, oplog)
	return nil
}
