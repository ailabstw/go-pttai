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

func (pm *BaseProtocolManager) RegisterOpKey(keyInfo *KeyInfo, isLocked bool) error {
	log.Debug("RegisterOpKeyInfo: start", "isLocked", isLocked, "key", keyInfo.Hash, "status", keyInfo.Status)
	if !isLocked {
		pm.lockOpKeyInfo.Lock()
		defer pm.lockOpKeyInfo.Unlock()
	}

	if keyInfo.Status != types.StatusAlive {
		return nil
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
	ptt.AddOpKey(keyInfo.Hash, entityID, false)
	log.Debug("RegisterOpKeyInfo: after ptt.AddOpKey", "hash", keyInfo.Hash, "entity", entityID, "service", pm.Entity().Service().Name())
	log.Debug("RegisterOpKeyInfo: done", "newestOpKey", pm.newestOpKeyInfo, "oldestOpKey", pm.oldestOpKeyInfo)

	return nil
}

func (pm *BaseProtocolManager) loadOpKeyInfos() ([]*KeyInfo, error) {
	e := pm.Entity()

	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	iter, err := opKey.BaseObject.GetObjIterWithObj(nil, pttdb.ListOrderNext, false)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	// XXX hack for skip expire-op-key
	/*
		expireTS, err := pm.getExpireOpKeyTS()
		if err != nil {
			return nil, err
		}
	*/

	log.Trace("loadOpKeyInfo: to for-loop")

	opKeyInfos := make([]*KeyInfo, 0)
	toRemoveOpKeys := make([][]byte, 0)
	toExpireOpKeyInfos := make([]*KeyInfo, 0)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		keyInfo := &KeyInfo{}
		err = keyInfo.Unmarshal(val)
		if err != nil {
			log.Warn("loadOpKeyInfo: unable to unmarshal", "key", key, "e", err)
			toRemoveOpKeys = append(toRemoveOpKeys, common.CloneBytes(key))
			continue
		}

		pm.SetOpKeyObjDB(keyInfo)

		// XXX hack for skip expire-op-key
		/*
			if keyInfo.UpdateTS.IsLess(expireTS) {
				log.Warn("loadOpKeyInfo: expire", "key", key, "expireTS", expireTS, "UpdateTS", keyInfo.UpdateTS)
				toExpireOpKeyInfos = append(toExpireOpKeyInfos, keyInfo)
				continue
			}
		*/

		opKeyInfos = append(opKeyInfos, keyInfo)
	}

	// to remove
	keyInfo := NewEmptyOpKey()
	pm.SetOpKeyObjDB(keyInfo)
	for _, key := range toRemoveOpKeys {
		err = keyInfo.DeleteByKey(key, true)
		if err != nil {
			log.Error("loadOpKeyInfos: unable to delete key", "name", e.Name(), "e", err)
		}
	}

	pm.lockOpKeyInfo.Lock()
	defer pm.lockOpKeyInfo.Unlock()
	for _, eachKeyInfo := range toExpireOpKeyInfos {
		pm.ExpireOpKeyInfo(eachKeyInfo, true)
	}

	return opKeyInfos, nil
}

func (pm *BaseProtocolManager) registerOpKeys(opKeys []*KeyInfo, isLocked bool) {
	if !isLocked {
		pm.lockOpKeyInfo.Lock()
		defer pm.lockOpKeyInfo.Unlock()
	}

	for _, keyInfo := range opKeys {
		pm.RegisterOpKey(keyInfo, true)
	}

}

func (pm *BaseProtocolManager) GetOpKeyFromHash(hash *common.Address, isLocked bool) (*KeyInfo, error) {
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
		log.Warn("GetNewestOpKey: newestOpKeyInfo nil to getNewestOpKeyFullScan", "e", pm.Entity().GetID())
		return pm.getNewestOpKeyFullScan(true)
	}

	expireRenewTS, err := pm.getExpireRenewOpKeyTS()
	if err != nil {
		log.Warn("GetNewestOpKey: unable to get expireRenewTS", "e", err)
		return nil, err
	}

	if pm.newestOpKeyInfo.UpdateTS.IsLess(expireRenewTS) {
		log.Warn("GetNewestOpKey: key expired renew ts", "key", pm.newestOpKeyInfo.UpdateTS, "expireRenew:", expireRenewTS)
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

	key, err := pm.getOldestOpKeyFullScan(true)
	log.Debug("GetOldestOpKey: after getOldestOpKeyFullScan", "key", key, "entity", pm.Entity().GetID(), "e", err)
	return key, err
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
	log.Debug("getNewestOpKeyFullScan: after checkNewestOpKeyInfo", "newestOpKey", pm.newestOpKeyInfo)
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
	if pm.expireOpKeySeconds < pm.renewOpKeySeconds {
		return now, nil
	}

	now.Ts += pm.renewOpKeySeconds

	return now, nil
}

func (pm *BaseProtocolManager) ExpireOpKeyInfo(keyInfo *KeyInfo, isLocked bool) error {
	if !isLocked {
		pm.lockOpKeyInfo.Lock()
		defer pm.lockOpKeyInfo.Unlock()
	}

	// XXX hack for skip expire-op-key
	return nil

	// return pm.RemoveOpKeyFromHash(keyInfo.Hash, true, true, true)
}

func (pm *BaseProtocolManager) RemoveOpKeyFromHash(hash *common.Address, isLocked bool, isDeleteOplog bool, isDeleteDB bool) error {

	// XXX hack for skip expire-op-key

	return nil

	/*
		entityID := pm.Entity().GetID()

		log.Debug("RemoveOpKeyFromHash: start")

		if !isLocked {
			pm.lockOpKeyInfo.Lock()
			defer pm.lockOpKeyInfo.Unlock()
		}

		keyInfo, ok := pm.opKeyInfos[*hash]
		if !ok {
			return nil
		}

		// delete pm.opKeyInfos
		delete(pm.opKeyInfos, *hash)

		if isDeleteOplog {
			if keyInfo.LogID != nil {
				pm.removeOpKeyOplog(keyInfo.LogID, false)
			} else {
				log.Error("keyInfo: unable to get", "keyInfo", keyInfo)
			}
			if keyInfo.CreateLogID != nil {
				pm.removeOpKeyOplog(keyInfo.CreateLogID, false)
			}
		}

		// delete db
		if isDeleteDB {
			keyInfo.Delete(false)
		}

		log.Debug("to ptt.RemoveOpKey")

		// ptt
		ptt := pm.Ptt()
		ptt.RemoveOpKey(hash, entityID, false)

		pm.getNewestOpKeyFullScan(true)

		return nil
	*/
}

func (pm *BaseProtocolManager) removeOpKeyOplog(logID *types.PttID, isLocked bool) error {
	return nil

	/*
		if !isLocked {
			err := pm.dbOpKeyLock.Lock(logID)
			if err != nil {
				return err
			}
			defer pm.dbOpKeyLock.Unlock(logID)
		}

		// delete oplog
		oplog := &BaseOplog{ID: logID}
		pm.SetOpKeyDB(oplog)
		oplog.Delete(false)

		return nil
	*/
}

func (pm *BaseProtocolManager) checkOldestOpKeyInfo(keyInfo *KeyInfo, expireRenewTS types.Timestamp) {
	// XXX hack for skipping newest-op-key
	/*
		if keyInfo.UpdateTS.IsLess(expireRenewTS) {
			return
		}
	*/

	if pm.oldestOpKeyInfo == nil || keyInfo.UpdateTS.IsLess(pm.oldestOpKeyInfo.UpdateTS) {
		pm.oldestOpKeyInfo = keyInfo
	}

	return
}

func (pm *BaseProtocolManager) checkNewestOpKeyInfo(keyInfo *KeyInfo, expireRenewTS types.Timestamp) {
	// XXX hack for skipping newest-op-key
	/*
		if keyInfo.UpdateTS.IsLess(expireRenewTS) {
			log.Warn("checkNewestOpKeyInfo: key expired renew ts", "key", keyInfo.UpdateTS, "expireTS", expireRenewTS)
			return
		}
	*/

	if pm.newestOpKeyInfo == nil || pm.newestOpKeyInfo.UpdateTS.IsLess(keyInfo.UpdateTS) {
		pm.newestOpKeyInfo = keyInfo
		log.Debug("checkNewestOpKeyInfo: after set newestOpKeyInfo", "keyInfo", keyInfo.UpdateTS, "id", pm.Entity().GetID())
	}

	return
}

func (pm *BaseProtocolManager) OpKeys() map[common.Address]*KeyInfo {
	return pm.opKeyInfos
}

func (pm *BaseProtocolManager) OpKeyList() []*KeyInfo {
	pm.lockOpKeyInfo.RLock()
	defer pm.lockOpKeyInfo.RUnlock()

	lenOpKeyInfos := len(pm.opKeyInfos)
	opKeyInfoList := make([]*KeyInfo, lenOpKeyInfos)
	i := 0
	for _, keyInfo := range pm.opKeyInfos {
		opKeyInfoList[i] = keyInfo
		i++
	}

	return opKeyInfoList
}

func (pm *BaseProtocolManager) RenewOpKeySeconds() int64 {
	return pm.renewOpKeySeconds
}

func (pm *BaseProtocolManager) ExpireOpKeySeconds() int64 {
	return pm.expireOpKeySeconds
}

func (pm *BaseProtocolManager) DBOpKey() *pttdb.LDBBatch {
	return pm.db
}

func (pm *BaseProtocolManager) DBOpKeyLock() *types.LockMap {
	return pm.dbOpKeyLock
}

func (pm *BaseProtocolManager) CleanOpKey() {
	pm.lockOpKeyInfo.Lock()
	defer pm.lockOpKeyInfo.Unlock()

	for _, keyInfo := range pm.opKeyInfos {
		pm.RemoveOpKeyFromHash(keyInfo.Hash, true, true, true)
	}
}

func (pm *BaseProtocolManager) DBOpKeyPrefix() []byte {
	return pm.dbOpKeyPrefix
}

func (pm *BaseProtocolManager) DBOpKeyIdxPrefix() []byte {
	return pm.dbOpKeyIdxPrefix
}
