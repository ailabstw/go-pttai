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
	"crypto/ecdsa"
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/crypto/bip32"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

// KeyInfo
type KeyInfo struct {
	*BaseObject `json:"b"`

	Hash *common.Address `json:"H"`

	Key         *ecdsa.PrivateKey `json:"-"`
	KeyBytes    []byte            `json:"K"`
	PubKeyBytes []byte            `json:"-"`

	UpdateTS types.Timestamp `json:"UT"`

	Extra *KeyExtraInfo `json:"e,omitempty"`

	SyncInfo *BaseSyncInfo `json:"s,omitempty"`

	CreateLogID *types.PttID `json:"c,omitempty"`

	Count int `json:"-"`
}

func NewJoinKeyInfo(entityID *types.PttID) (*KeyInfo, error) {
	key, err := deriveJoinKey()
	if err != nil {
		return nil, err
	}
	extendedKey, err := bip32.PrivKeyToExtKey(key, nil)
	if err != nil {
		return nil, err
	}

	return newKeyInfo(extendedKey, nil, entityID, nil, nil, nil, nil, nil)
}

func NewOpKeyInfo(entityID *types.PttID, doerID *types.PttID, masterKey *ecdsa.PrivateKey, db *pttdb.LDBBatch, dbLock *types.LockMap, fullDBPrefix []byte, fullDBIdxPrefix []byte) (*KeyInfo, error) {
	key, extra, err := deriveOpKey(masterKey)
	if err != nil {
		return nil, err
	}

	return newKeyInfo(key, extra, entityID, doerID, db, dbLock, fullDBPrefix, fullDBIdxPrefix)
}

func NewSignKeyInfo(doerID *types.PttID, masterKey *ecdsa.PrivateKey) (*KeyInfo, error) {
	key, extra, err := deriveSignKey(masterKey)
	if err != nil {
		return nil, err
	}
	return newKeyInfo(key, extra, nil, doerID, nil, nil, nil, nil)
}

func newKeyInfo(extendedKey *bip32.ExtendedKey, extra *KeyExtraInfo, entityID *types.PttID, doerID *types.PttID, db *pttdb.LDBBatch, dbLock *types.LockMap, fullDBPrefix []byte, fullDBIdxPrefix []byte) (*KeyInfo, error) {

	key, err := extendedKey.ToPrivkey()
	if err != nil {
		return nil, err
	}

	privBytes, err := extendedKey.PrivkeyBytes()
	if err != nil {
		return nil, err
	}
	pubBytes := extendedKey.PubkeyBytes()
	hash := crypto.PubkeyBytesToAddress(pubBytes)

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	id := keyInfoHashToID(&hash)

	return &KeyInfo{
		BaseObject: NewObject(id, ts, doerID, entityID, nil, types.StatusInvalid, db, dbLock, fullDBPrefix, fullDBIdxPrefix),

		Hash:        &hash,
		Key:         key,
		KeyBytes:    privBytes,
		PubKeyBytes: pubBytes,
		UpdateTS:    ts,
		Extra:       extra,
	}, nil
}

func keyInfoIDToHash(id *types.PttID) *common.Address {
	hash := &common.Address{}
	copy(hash[:], id[:common.AddressLength])

	return hash
}

func keyInfoHashToID(hash *common.Address) *types.PttID {
	id := &types.PttID{}
	copy(id[:common.AddressLength], hash[:])

	return id
}

func deriveJoinKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

func deriveOpKey(masterKey *ecdsa.PrivateKey) (*bip32.ExtendedKey, *KeyExtraInfo, error) {
	return deriveKeyBIP32(masterKey)
}

func deriveSignKey(masterKey *ecdsa.PrivateKey) (*bip32.ExtendedKey, *KeyExtraInfo, error) {
	return deriveKeyBIP32(masterKey)
}

func deriveKeyBIP32(masterKey *ecdsa.PrivateKey) (*bip32.ExtendedKey, *KeyExtraInfo, error) {
	var err error
	var extendedKey *bip32.ExtendedKey
	var salt *types.Salt

	var idx uint32
	for idx = 0; idx < MaxIterDeriveKeyBIP32; idx++ {
		extendedKey, salt, err = bip32.Key(masterKey, idx)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, nil, err
	}

	masterPubBytes := crypto.FromECDSAPub(&masterKey.PublicKey)

	extraBIP32 := &KeyBIP32{
		Parent: masterPubBytes,
		Salt:   salt,
		Child:  idx,
	}

	extra := &KeyExtraInfo{
		KeyType: KeyTypeBIP32,
		Data:    extraBIP32,
	}

	return extendedKey, extra, nil
}

func (k *KeyInfo) Init(db *pttdb.LDBBatch, dbLock *types.LockMap, entityID *types.PttID, fullDBPrefix []byte, fullDBIdxPrefix []byte) error {
	k.SetDB(db, dbLock, entityID, fullDBPrefix, fullDBIdxPrefix)
	key, err := crypto.ToECDSA(k.KeyBytes)
	if err != nil {
		return err
	}

	pubKeyBytes := crypto.FromECDSAPub(&key.PublicKey)

	k.Key = key
	k.PubKeyBytes = pubKeyBytes

	return nil
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

	_, err = k.db.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeyInfo) NewEmptyObj() Object {
	return &KeyInfo{BaseObject: &BaseObject{EntityID: k.EntityID, db: k.db, dbLock: k.dbLock, fullDBPrefix: k.fullDBPrefix}}

}

func (k *KeyInfo) GetNewObjByID(id *types.PttID, isLocked bool) (Object, error) {
	newK := k.NewEmptyObj()
	err := newK.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newK, nil
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
 * Block Info
 **********/

/*
GetBlockInfo implements Object method
*/
func (k *KeyInfo) GetBlockInfo() BlockInfo {
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
	syncInfo, ok := theSyncInfo.(*BaseSyncInfo)
	if !ok {
		return ErrInvalidSyncInfo
	}

	k.SyncInfo = syncInfo

	return nil
}
