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

type SyncKeyInfo struct {
	*BaseSyncInfo `json:"b"`
}

func EmptySyncKeyInfo() *SyncKeyInfo {
	return &SyncKeyInfo{BaseSyncInfo: &BaseSyncInfo{}}
}

// KeyInfo
type KeyInfo struct {
	*BaseObject `json:"b"`

	Hash *common.Address `json:"H"`

	Key         *ecdsa.PrivateKey `json:"-"`
	KeyBytes    []byte            `json:"K"`
	PubKeyBytes []byte            `json:"-"`

	UpdateTS types.Timestamp `json:"UT"`

	Extra *KeyExtraInfo `json:"e,omitempty"`

	SyncInfo *SyncKeyInfo `json:"s,omitempty"`

	CreateLogID *types.PttID `json:"cl,omitempty"`

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

	return newKeyInfo(extendedKey, nil, entityID, nil, nil, nil)
}

func NewOpKeyInfo(entityID *types.PttID, doerID *types.PttID, masterKey *ecdsa.PrivateKey, db *pttdb.LDBBatch, dbLock *types.LockMap) (*KeyInfo, error) {
	key, extra, err := deriveOpKey(masterKey)
	if err != nil {
		return nil, err
	}

	return newKeyInfo(key, extra, entityID, doerID, db, dbLock)
}

func NewSignKeyInfo(doerID *types.PttID, masterKey *ecdsa.PrivateKey) (*KeyInfo, error) {
	key, extra, err := deriveSignKey(masterKey)
	if err != nil {
		return nil, err
	}
	return newKeyInfo(key, extra, nil, doerID, nil, nil)
}

func NewEmptyKeyInfo() *KeyInfo {
	return &KeyInfo{BaseObject: &BaseObject{}}
}

func newKeyInfo(extendedKey *bip32.ExtendedKey, extra *KeyExtraInfo, entityID *types.PttID, doerID *types.PttID, db *pttdb.LDBBatch, dbLock *types.LockMap) (*KeyInfo, error) {

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
		BaseObject: NewObject(id, ts, doerID, doerID, entityID, nil, types.StatusInvalid, db, dbLock),

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

func (k *KeyInfo) Init(db *pttdb.LDBBatch, dbLock *types.LockMap) error {
	k.SetDB(db, dbLock, DBOpKeyPrefix)
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

	log.Debug("Save: to PutAll", "idxKey", idxKey, "idx", idx, "key", key, "marshaled", marshaled)

	_, err = k.db.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeyInfo) GetByID(isLocked bool) error {
	var err error

	if !isLocked {
		err = k.RLock()
		if err != nil {
			return err
		}
		defer k.RUnlock()
	}

	idxKey, err := k.IdxKey()
	if err != nil {
		return err
	}

	val, err := k.db.GetByIdxKey(idxKey, 0)
	if err != nil {
		return err
	}

	return k.Unmarshal(val)
}

func (k *KeyInfo) GetNewObjByID(id *types.PttID, isLocked bool) (Object, error) {
	newK := &KeyInfo{BaseObject: &BaseObject{ID: id, EntityID: k.EntityID, db: k.db, dbLock: k.dbLock}}

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

func (k *KeyInfo) Delete(isLocked bool) error {
	var err error

	log.Debug("Delete: start")

	if !isLocked {
		err = k.Lock()
		if err != nil {
			return err
		}
		defer k.Unlock()
	}

	idxKey, err := k.IdxKey()
	if err != nil {
		return err
	}

	err = k.db.DeleteAll(idxKey)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeyInfo) IdxPrefix() []byte {
	return append(DBOpKeyIdxPrefix, k.EntityID[:]...)
}

func (k *KeyInfo) IdxKey() ([]byte, error) {
	return common.Concat([][]byte{DBOpKeyIdxPrefix, k.EntityID[:], k.ID[:]})
}

func (k *KeyInfo) DBPrefix() []byte {
	return append(DBOpKeyPrefix, k.EntityID[:]...)
}

func (k *KeyInfo) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := k.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{DBOpKeyPrefix, k.EntityID[:], marshalTimestamp, k.ID[:]})
}

func (k *KeyInfo) KeyToIdxKey(key []byte) ([]byte, error) {

	lenKey := len(key)
	if lenKey != pttdb.SizeDBKeyPrefix+types.SizePttID+types.SizeTimestamp+types.SizePttID {
		return nil, ErrInvalidKey
	}

	idxKey := make([]byte, pttdb.SizeDBKeyPrefix+types.SizePttID+types.SizePttID)

	// prefix
	idxOffset := 0
	nextIdxOffset := pttdb.SizeDBKeyPrefix
	copy(idxKey[:nextIdxOffset], DBOpKeyIdxPrefix)

	// entity-id
	idxOffset = nextIdxOffset
	nextIdxOffset += types.SizePttID

	keyOffset := pttdb.SizeDBKeyPrefix
	nextKeyOffset := keyOffset + types.SizePttID
	copy(idxKey[idxOffset:nextIdxOffset], key[keyOffset:nextKeyOffset])

	// id
	idxOffset = nextIdxOffset
	nextIdxOffset += types.SizePttID

	keyOffset = lenKey - types.SizePttID
	nextKeyOffset = lenKey
	copy(idxKey[idxOffset:nextIdxOffset], key[keyOffset:nextKeyOffset])

	return idxKey, nil
}

func (k *KeyInfo) DeleteKey(key []byte) error {
	idxKey, err := k.KeyToIdxKey(key)
	if err != nil {
		return err
	}

	err = k.db.DeleteAll(idxKey)

	if err != nil {
		return err
	}

	return nil
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
