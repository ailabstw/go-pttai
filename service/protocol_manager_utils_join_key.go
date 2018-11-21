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
	"reflect"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) GetJoinKeyFromHash(hash *common.Address) (*KeyInfo, error) {
	pm.lockJoinKeyInfo.RLock()
	defer pm.lockJoinKeyInfo.RUnlock()

	log.Debug("GetJoinKeyFromHash: to for-loop", "e", pm.Entity().GetID(), "hash", hash, "joinKeyInfos", pm.joinKeyInfos)

	var keyInfo *KeyInfo = nil
	for _, eachKeyInfo := range pm.joinKeyInfos {
		log.Debug("GetJoinKeyFromHash (in for-loop)", "eachHash", eachKeyInfo.Hash)
		if reflect.DeepEqual(hash, eachKeyInfo.Hash) {
			keyInfo = eachKeyInfo
			break
		}
	}

	if keyInfo == nil {
		return nil, ErrInvalidKeyInfo
	}

	return keyInfo, nil
}

func (pm *BaseProtocolManager) GetJoinKey() (*KeyInfo, error) {
	pm.lockJoinKeyInfo.RLock()
	defer pm.lockJoinKeyInfo.RUnlock()

	lenKeyInfo := len(pm.joinKeyInfos)

	if lenKeyInfo == 0 {
		return nil, ErrInvalidKeyInfo
	}

	return pm.joinKeyInfos[lenKeyInfo-1], nil
}

func (pm *BaseProtocolManager) CreateJoinKeyLoop() error {
	ticker := time.NewTicker(RenewJoinKeySeconds)
	defer ticker.Stop()

	pm.createJoinKey()

loop:
	for {
		select {
		case <-ticker.C:
			pm.createJoinKey()
		case <-pm.QuitSync():
			log.Debug("CreateJoinKeyLoop: QuitSync", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
			break loop
		}
	}

	return nil
}

func (pm *BaseProtocolManager) createJoinKey() error {
	status := pm.Entity().GetStatus()
	statusClass := types.StatusToStatusClass(status)
	if statusClass >= types.StatusClassDeleted {
		return nil
	}

	myEntity := pm.Ptt().GetMyEntity()
	status = myEntity.GetStatus()
	statusClass = types.StatusToStatusClass(status)
	if statusClass >= types.StatusClassDeleted {
		return nil
	}

	if !pm.IsMaster(myEntity.GetID(), false) {
		return nil
	}

	pm.lockJoinKeyInfo.Lock()
	defer pm.lockJoinKeyInfo.Unlock()

	entityID := pm.Entity().GetID()
	newKeyInfo, err := NewJoinKeyInfo(entityID)
	if err != nil {
		return err
	}

	if len(pm.joinKeyInfos) > 2 {
		origKeyInfo := pm.joinKeyInfos[0]
		pm.ptt.RemoveJoinKey(origKeyInfo.Hash, entityID, false)
		pm.joinKeyInfos = pm.joinKeyInfos[1:]
	}

	pm.joinKeyInfos = append(pm.joinKeyInfos, newKeyInfo)
	log.Debug("createJoinKeyInfo: to AddJoinKey", "e", pm.Entity().GetID(), "joinKeyInfos", pm.joinKeyInfos)
	pm.ptt.AddJoinKey(newKeyInfo.Hash, entityID, false)

	return nil
}

func (pm *BaseProtocolManager) JoinKeyList() []*KeyInfo {
	return pm.joinKeyInfos
}

func (pm *BaseProtocolManager) IsJoinKeyHash(hash *common.Address) bool {
	pm.lockJoinKeyInfo.RLock()
	defer pm.lockJoinKeyInfo.RUnlock()

	for _, eachKeyInfo := range pm.joinKeyInfos {
		if reflect.DeepEqual(eachKeyInfo.Hash, hash) {
			return true
		}
	}

	return false
}

func (pm *BaseProtocolManager) GetJoinType(hash *common.Address) (JoinType, error) {
	return JoinTypeInvalid, types.ErrNotImplemented
}

func (pm *BaseProtocolManager) CleanJoinKey() {
	pm.lockJoinKeyInfo.Lock()
	defer pm.lockJoinKeyInfo.Unlock()

	entityID := pm.Entity().GetID()

	for _, keyInfo := range pm.joinKeyInfos {
		pm.ptt.RemoveJoinKey(keyInfo.Hash, entityID, false)
	}
}
