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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

func NewEmptyOpKey() *KeyInfo {
	return &KeyInfo{BaseObject: &BaseObject{}}
}

func OpKeysToObjs(typedObjs []*KeyInfo) []Object {
	objs := make([]Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToOpKeys(objs []Object) []*KeyInfo {
	typedObjs := make([]*KeyInfo, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*KeyInfo)
	}
	return typedObjs
}

func (pm *BaseProtocolManager) SetOpKeyObjDB(opKey *KeyInfo) {
	opKey.SetDB(pm.DBOpKey(), pm.DBObjLock(), pm.Entity().GetID(), pm.dbOpKeyPrefix, pm.dbOpKeyIdxPrefix, nil, nil)
}

func (k *KeyInfo) Init(
	pm ProtocolManager,
) error {

	pm.SetOpKeyObjDB(k)

	key, err := crypto.ToECDSA(k.KeyBytes)
	if err != nil {
		return err
	}

	pubKeyBytes := crypto.FromECDSAPub(&key.PublicKey)

	k.Key = key
	k.PubKeyBytes = pubKeyBytes

	return nil
}

func (k *KeyInfo) NewEmptyObj() Object {
	newObj := NewEmptyOpKey()
	newObj.CloneDB(k.BaseObject)
	return newObj
}

func (k *KeyInfo) GetNewObjByID(id *types.PttID, isLocked bool) (Object, error) {
	newObj := k.NewEmptyObj()
	newObj.SetID(id)
	err := newObj.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newObj, nil
}

func (k *KeyInfo) SetUpdateTS(ts types.Timestamp) {
	k.UpdateTS = ts
}

func (k *KeyInfo) GetUpdateTS() types.Timestamp {
	return k.UpdateTS
}

func (k *KeyInfo) GetByID(isLocked bool) error {
	var err error

	val, err := k.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return k.Unmarshal(val)
}

func (k *KeyInfo) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = k.Lock()
		if err != nil {
			return err
		}
		defer k.Unlock()
	}

	if k.Key == nil && k.KeyBytes != nil {
		k.Key, err = crypto.ToECDSA(k.KeyBytes)
		if err != nil {
			return err
		}
	}
	key, err := k.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := k.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := k.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: k.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	log.Debug("KeyInfo: to Save", "idxKey", idxKey)

	_, err = k.db.ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}
func (k *KeyInfo) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := k.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{k.fullDBPrefix, marshalTimestamp, k.ID[:]})
}

func (k *KeyInfo) Marshal() ([]byte, error) {
	return json.Marshal(k)
}

func (k *KeyInfo) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, k)
	if err != nil {
		return err
	}

	if k.KeyBytes != nil {
		k.Key, err = crypto.ToECDSA(k.KeyBytes)
		if err != nil {
			return err
		}

		k.PubKeyBytes = crypto.FromECDSAPub(&k.Key.PublicKey)
	}

	return nil
}

/**********
 * Sync Info
 **********/

func (k *KeyInfo) GetSyncInfo() SyncInfo {
	if k.SyncInfo == nil {
		return nil
	}
	return k.SyncInfo
}

func (k *KeyInfo) SetSyncInfo(theSyncInfo SyncInfo) error {
	if theSyncInfo == nil {
		k.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*BaseSyncInfo)
	if !ok {
		return ErrInvalidSyncInfo
	}

	k.SyncInfo = syncInfo

	return nil
}
