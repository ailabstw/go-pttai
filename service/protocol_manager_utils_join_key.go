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

func (b *BaseProtocolManager) GetJoinKeyFromHash(hash *common.Address) (*KeyInfo, error) {
	b.lockJoinKeyInfo.RLock()
	defer b.lockJoinKeyInfo.RUnlock()

	log.Debug("GetJoinKeyFromHash: to for-loop", "e", b.Entity().GetID(), "hash", hash, "joinKeyInfos", b.joinKeyInfos)

	var keyInfo *KeyInfo = nil
	for _, eachKeyInfo := range b.joinKeyInfos {
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

func (b *BaseProtocolManager) GetJoinKey() (*KeyInfo, error) {
	b.lockJoinKeyInfo.RLock()
	defer b.lockJoinKeyInfo.RUnlock()

	lenKeyInfo := len(b.joinKeyInfos)

	if lenKeyInfo == 0 {
		return nil, ErrInvalidKeyInfo
	}

	return b.joinKeyInfos[lenKeyInfo-1], nil
}

func (b *BaseProtocolManager) CreateJoinKeyLoop() error {
	ticker := time.NewTicker(RenewJoinKeySeconds)
	defer ticker.Stop()

	b.createJoinKey()

loop:
	for {
		select {
		case <-ticker.C:
			b.createJoinKey()
		case <-b.QuitSync():
			break loop
		}
	}

	return nil
}

func (b *BaseProtocolManager) createJoinKey() error {
	status := b.Entity().GetStatus()
	statusClass := types.StatusToStatusClass(status)
	if statusClass >= types.StatusClassDeleted {
		return nil
	}

	myEntity := b.Ptt().GetMyEntity()
	status = myEntity.GetStatus()
	statusClass = types.StatusToStatusClass(status)
	if statusClass >= types.StatusClassDeleted {
		return nil
	}

	if !b.IsMaster(myEntity.GetID(), false) {
		return nil
	}

	b.lockJoinKeyInfo.Lock()
	defer b.lockJoinKeyInfo.Unlock()

	entityID := b.Entity().GetID()
	newKeyInfo, err := NewJoinKeyInfo(entityID)
	if err != nil {
		return err
	}

	if len(b.joinKeyInfos) > 2 {
		origKeyInfo := b.joinKeyInfos[0]
		b.ptt.RemoveJoinKey(origKeyInfo.Hash, entityID, false)
		b.joinKeyInfos = b.joinKeyInfos[1:]
	}

	b.joinKeyInfos = append(b.joinKeyInfos, newKeyInfo)
	log.Debug("createJoinKeyInfo: to AddJoinKey", "e", b.Entity().GetID(), "joinKeyInfos", b.joinKeyInfos)
	b.ptt.AddJoinKey(newKeyInfo.Hash, entityID, false)

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
