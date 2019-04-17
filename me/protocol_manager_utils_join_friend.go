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

package me

import (
	"reflect"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
)

func (pm *ProtocolManager) GetJoinFriendKeyFromHash(hash *common.Address) (*pkgservice.KeyInfo, error) {
	pm.lockJoinFriendKeyInfo.RLock()
	defer pm.lockJoinFriendKeyInfo.RUnlock()

	var keyInfo *pkgservice.KeyInfo = nil
	for _, eachKeyInfo := range pm.joinFriendKeyInfos {
		if reflect.DeepEqual(hash, eachKeyInfo.Hash) {
			keyInfo = eachKeyInfo
			break
		}
	}

	if keyInfo == nil {
		return nil, pkgservice.ErrInvalidKeyInfo
	}

	return keyInfo, nil
}

func (pm *ProtocolManager) GetJoinFriendKey() (*pkgservice.KeyInfo, error) {
	pm.lockJoinFriendKeyInfo.RLock()
	defer pm.lockJoinFriendKeyInfo.RUnlock()

	lenKeyInfo := len(pm.joinFriendKeyInfos)

	if lenKeyInfo == 0 {
		return nil, pkgservice.ErrInvalidKeyInfo
	}

	return pm.joinFriendKeyInfos[lenKeyInfo-1], nil
}

func (pm *ProtocolManager) CreateJoinFriendKeyLoop() error {
	ticker := time.NewTicker(RenewJoinFriendKeySeconds)
	defer ticker.Stop()

	pm.createJoinFriendKey()

loop:
	for {
		select {
		case <-ticker.C:
			pm.createJoinFriendKey()
		case <-pm.QuitSync():
			log.Debug("CreateJoinFriendKeyLoop: QuitSync", "entity", pm.Entity().GetID())
			break loop
		}
	}

	return nil
}

func (pm *ProtocolManager) createJoinFriendKey() error {
	status := pm.Entity().GetStatus()
	statusClass := types.StatusToStatusClass(status)
	if statusClass >= types.StatusClassDeleted {
		return nil
	}

	ptt := pm.Ptt()
	myEntity := ptt.GetMyEntity()
	status = myEntity.GetStatus()
	statusClass = types.StatusToStatusClass(status)
	if statusClass >= types.StatusClassDeleted {
		return nil
	}

	if !pm.IsMaster(myEntity.GetID(), false) {
		return nil
	}

	pm.lockJoinFriendKeyInfo.Lock()
	defer pm.lockJoinFriendKeyInfo.Unlock()

	entityID := pm.Entity().GetID()
	newKeyInfo, err := pkgservice.NewJoinKeyInfo(entityID)
	if err != nil {
		return err
	}

	if len(pm.joinFriendKeyInfos) > 2 {
		origKeyInfo := pm.joinFriendKeyInfos[0]
		ptt.RemoveJoinKey(origKeyInfo.Hash, entityID, false)
		pm.joinFriendKeyInfos = pm.joinFriendKeyInfos[1:]
	}

	pm.joinFriendKeyInfos = append(pm.joinFriendKeyInfos, newKeyInfo)
	ptt.AddJoinKey(newKeyInfo.Hash, entityID, false)

	return nil
}

func (pm *ProtocolManager) JoinFriendKeyList() []*pkgservice.KeyInfo {
	return pm.joinFriendKeyInfos
}

func (pm *ProtocolManager) IsJoinFriendKeyHash(hash *common.Address) bool {
	pm.lockJoinFriendKeyInfo.RLock()
	defer pm.lockJoinFriendKeyInfo.RUnlock()

	for _, eachKeyInfo := range pm.joinFriendKeyInfos {
		if reflect.DeepEqual(eachKeyInfo.Hash, hash) {
			return true
		}
	}

	return false
}

func (pm *ProtocolManager) CleanJoinFriendKey() {
	pm.lockJoinFriendKeyInfo.Lock()
	defer pm.lockJoinFriendKeyInfo.Unlock()

	entityID := pm.Entity().GetID()

	ptt := pm.Ptt()
	for _, keyInfo := range pm.joinFriendKeyInfos {
		ptt.RemoveJoinKey(keyInfo.Hash, entityID, false)
	}
}

func (pm *ProtocolManager) IsJoinFriendRequests(hash *common.Address) bool {
	pm.lockJoinFriendRequest.RLock()
	defer pm.lockJoinFriendRequest.RUnlock()

	_, ok := pm.joinFriendRequests[*hash]

	return ok
}

func (pm *ProtocolManager) GetFriendRequests() ([]*pkgservice.JoinRequest, error) {
	pm.lockJoinFriendRequest.RLock()
	defer pm.lockJoinFriendRequest.RUnlock()

	theList := make([]*pkgservice.JoinRequest, len(pm.joinFriendRequests))
	i := 0
	for _, request := range pm.joinFriendRequests {
		theList[i] = request
		i++
	}
	return theList, nil
}

func (pm *ProtocolManager) RemoveFriendRequests(hash []byte) (bool, error) {
	pm.lockJoinFriendRequest.Lock()
	defer pm.lockJoinFriendRequest.Unlock()

	addr := &common.Address{}
	copy(addr[:], hash)
	_, ok := pm.joinFriendRequests[*addr]
	if !ok {
		return false, types.ErrAlreadyDeleted
	}

	delete(pm.joinFriendRequests, *addr)

	return true, nil
}
