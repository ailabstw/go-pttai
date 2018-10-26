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
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *BaseProtocolManager) GetOpKeyInfoFromHash(hash *common.Address, isLocked bool) (*KeyInfo, error) {
	if !isLocked {
		pm.lockOpKeyInfo.RLock()
		defer pm.lockOpKeyInfo.RUnlock()
	}

	keyInfo, ok := pm.opKeyInfos[*hash]
	if !ok {
		return nil, ErrInvalidKeyInfo
	}

	return keyInfo, nil
}

func (pm *BaseProtocolManager) GetNewestOpKey(isLocked bool) (*KeyInfo, error) {
	if !isLocked {
		pm.lockOpKeyInfo.RLock()
		defer pm.lockOpKeyInfo.RUnlock()
	}

	if pm.newestOpKeyInfo == nil {
		return nil, ErrInvalidKey
	}

	expireRenewTS, err := pm.getExpireRenewOpKeyTS()
	if err != nil {
		return nil, err
	}

	if pm.newestOpKeyInfo.UpdateTS.IsLess(expireRenewTS) {
		return nil, ErrInvalidKey
	}

	return pm.newestOpKeyInfo, nil
}

func (pm *BaseProtocolManager) GetOldestOpKey(isLocked bool) (*KeyInfo, error) {
	if !isLocked {
		pm.lockOpKeyInfo.RLock()
		defer pm.lockOpKeyInfo.RUnlock()
	}

	if len(pm.opKeyInfos) == 0 {
		return nil, ErrInvalidKey
	}

	expireRenewTS, err := pm.getExpireRenewOpKeyTS()
	if err != nil {
		return nil, err
	}

	if pm.oldestOpKeyInfo != nil && !pm.oldestOpKeyInfo.UpdateTS.IsLess(expireRenewTS) {
		return pm.oldestOpKeyInfo, nil
	}

	return pm.getOldestOpKeyFullScan(true)
}

func (pm *BaseProtocolManager) getNewestOpKeyFullScan(isLocked bool) (*KeyInfo, error) {
	if !isLocked {
		pm.lockOpKeyInfo.RLock()
		defer pm.lockOpKeyInfo.RUnlock()
	}

	expireRenewTS, err := pm.getExpireRenewOpKeyTS()
	if err != nil {
		return nil, err
	}

	for _, keyInfo := range pm.opKeyInfos {
		pm.checkNewestOpKeyInfo(keyInfo, expireRenewTS)
	}

	if pm.newestOpKeyInfo == nil {
		return nil, ErrInvalidKey
	}

	return pm.newestOpKeyInfo, nil
}

func (pm *BaseProtocolManager) getOldestOpKeyFullScan(isLocked bool) (*KeyInfo, error) {
	if !isLocked {
		pm.lockOpKeyInfo.RLock()
		defer pm.lockOpKeyInfo.RUnlock()
	}

	expireRenewTS, err := pm.getExpireRenewOpKeyTS()
	if err != nil {
		return nil, err
	}

	for _, keyInfo := range pm.opKeyInfos {
		pm.checkOldestOpKeyInfo(keyInfo, expireRenewTS)
	}

	if pm.oldestOpKeyInfo == nil {
		return nil, ErrInvalidKey
	}

	return pm.oldestOpKeyInfo, nil
}

func (pm *BaseProtocolManager) RegisterOpKeyInfo(keyInfo *KeyInfo, isLocked bool, isPttLocked bool) error {
	if !isLocked {
		pm.lockOpKeyInfo.Lock()
		defer pm.lockOpKeyInfo.Unlock()
	}

	pm.opKeyInfos[*keyInfo.Hash] = keyInfo

	expireRenewTS, err := pm.getExpireRenewOpKeyTS()
	if err != nil {
		return err
	}

	pm.checkNewestOpKeyInfo(keyInfo, expireRenewTS)
	if pm.oldestOpKeyInfo == nil {
		pm.checkOldestOpKeyInfo(keyInfo, expireRenewTS)
	}

	ptt := pm.Ptt()
	entityID := pm.Entity().GetID()
	ptt.AddOpKey(keyInfo.Hash, entityID, isPttLocked)

	return nil
}

func (pm *BaseProtocolManager) getExpireOpKeyTS() (types.Timestamp, error) {
	now, err := types.GetTimestamp()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	now.Ts -= pm.expireOpKeySeconds

	return now, nil
}

func (pm *BaseProtocolManager) getExpireRenewOpKeyTS() (types.Timestamp, error) {
	now, err := types.GetTimestamp()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	now.Ts -= pm.expireOpKeySeconds
	now.Ts += pm.renewOpKeySeconds

	return now, nil
}

func (pm *BaseProtocolManager) RemoveOpKeyInfo(keyInfo *KeyInfo, isLocked bool) error {
	if !isLocked {
		pm.lockOpKeyInfo.Lock()
		defer pm.lockOpKeyInfo.Unlock()
	}
	return pm.RemoveOpKeyInfoFromHash(keyInfo.Hash, true)
}

func (pm *BaseProtocolManager) RemoveOpKeyInfoFromHash(hash *common.Address, isLocked bool) error {
	entityID := pm.Entity().GetID()

	if !isLocked {
		pm.lockOpKeyInfo.Lock()
		defer pm.lockOpKeyInfo.Unlock()
	}

	removeOpKeyInfo, ok := pm.opKeyInfos[*hash]
	if !ok {
		return nil
	}

	// delete pm.opKeyInfos
	delete(pm.opKeyInfos, *hash)

	// delete db
	removeOpKeyInfo.Delete(pm.DBOpKeyInfo(), false)

	pm.getNewestOpKeyFullScan(true)

	// ptt
	ptt := pm.Ptt()
	ptt.RemoveOpKey(hash, entityID, false)

	return nil
}

func (pm *BaseProtocolManager) loadOpKeyInfos() ([]*KeyInfo, error) {
	e := pm.Entity()
	entityID := e.GetID()

	dbPrefix, err := DBPrefix(DBOpKeyPrefix, entityID)
	if err != nil {
		return nil, err
	}

	iter, err := pm.db.DB().NewIteratorWithPrefix(dbPrefix, dbPrefix, pttdb.ListOrderNext)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	expireTS, err := pm.getExpireOpKeyTS()
	if err != nil {
		return nil, err
	}

	opKeyInfos := make([]*KeyInfo, 0)
	toRemoveOpKeys := make([][]byte, 0)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		keyInfo := &KeyInfo{}
		err = keyInfo.Unmarshal(val)
		if err != nil {
			log.Warn("loadOpKeyInfo: unable to unmarshal", "key", key)
			toRemoveOpKeys = append(toRemoveOpKeys, common.CloneBytes(key))
			continue
		}

		if keyInfo.UpdateTS.IsLess(expireTS) {
			log.Warn("loadOpKeyInfo: expire", "key", key, "expireTS", expireTS, "UpdateTS", keyInfo.UpdateTS)
			toRemoveOpKeys = append(toRemoveOpKeys, common.CloneBytes(key))
			continue
		}

		err = keyInfo.Init(pm.dbOpKeyLock)
		if err != nil {
			log.Warn("loadOpKeyInfo: unable to init", "key", key)
			toRemoveOpKeys = append(toRemoveOpKeys, common.CloneBytes(key))
			continue
		}

		opKeyInfos = append(opKeyInfos, keyInfo)
	}

	log.Debug("loadOpKeyInfo: after loop", "toRemoveOpKeys", toRemoveOpKeys)

	for _, key := range toRemoveOpKeys {
		keyInfo := &KeyInfo{}
		err := keyInfo.DeleteKey(key, pm.db)
		if err != nil {
			log.Error("loadOpKeyInfos: unable to delete key", "name", e.Name(), "dbPrefix", dbPrefix, "e", err)
		}
	}

	return opKeyInfos, nil
}

func (pm *BaseProtocolManager) checkOldestOpKeyInfo(keyInfo *KeyInfo, expireRenewTS types.Timestamp) {
	if keyInfo.UpdateTS.IsLess(expireRenewTS) {
		return
	}

	if pm.oldestOpKeyInfo == nil || keyInfo.UpdateTS.IsLess(pm.oldestOpKeyInfo.UpdateTS) {
		pm.oldestOpKeyInfo = keyInfo
	}

	return
}

func (pm *BaseProtocolManager) checkNewestOpKeyInfo(keyInfo *KeyInfo, expireRenewTS types.Timestamp) {
	if keyInfo.UpdateTS.IsLess(expireRenewTS) {
		return
	}

	if pm.newestOpKeyInfo == nil || pm.newestOpKeyInfo.UpdateTS.IsLess(keyInfo.UpdateTS) {
		pm.newestOpKeyInfo = keyInfo
	}

	return
}

func (pm *BaseProtocolManager) RenewOpKeySeconds() uint64 {
	return pm.renewOpKeySeconds
}

func (pm *BaseProtocolManager) ExpireOpKeySeconds() uint64 {
	return pm.expireOpKeySeconds
}

func (pm *BaseProtocolManager) DBOpKeyInfo() *pttdb.LDBBatch {
	return pm.db
}

func (pm *BaseProtocolManager) DBOpKeyLock() *types.LockMap {
	return pm.dbOpKeyLock
}
