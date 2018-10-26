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
	"github.com/syndtr/goleveldb/leveldb"
)

// KeyInfo
type KeyInfo struct {
	V  types.Version
	ID *types.PttID

	Key         *ecdsa.PrivateKey `json:"-"`
	KeyBytes    []byte            `json:"K"`
	PubKeyBytes []byte            `json:"-"`

	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`

	EntityID *types.PttID `json:"EID"`
	DoerID   *types.PttID `json:"DID"`

	LogID *types.PttID  `json:"l,omitempty"`
	Extra *KeyExtraInfo `json:"e,omitempty"`
}

func NewJoinKeyInfo(entityID *types.PttID, doerID *types.PttID) (*KeyInfo, error) {
	key, err := deriveJoinKey()
	if err != nil {
		return nil, err
	}
	extendedKey, err := bip32.PrivKeyToExtKey(key, nil)
	if err != nil {
		return nil, err
	}

	return newKeyInfo(extendedKey, nil, entityID, doerID)
}

func NewOpKeyInfo(entityID *types.PttID, doerID *types.PttID, masterKey *ecdsa.PrivateKey) (*KeyInfo, error) {
	key, extra, err := deriveOpKey(masterKey)
	if err != nil {
		return nil, err
	}

	return newKeyInfo(key, extra, entityID, doerID)
}

func NewSignKeyInfo(entityID *types.PttID, doerID *types.PttID, masterKey *ecdsa.PrivateKey) (*KeyInfo, error) {
	key, extra, err := deriveSignKey(masterKey)
	if err != nil {
		return nil, err
	}
	return newKeyInfo(key, extra, entityID, doerID)
}

func newKeyInfo(extendedKey *bip32.ExtendedKey, extra *KeyExtraInfo, entityID *types.PttID, doerID *types.PttID) (*KeyInfo, error) {

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

	updateTS, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	id := &types.PttID{}
	copy(id[:common.AddressLength], hash[:])

	return &KeyInfo{
		V:           types.CurrentVersion,
		ID:          id,
		Key:         key,
		KeyBytes:    privBytes,
		PubKeyBytes: pubBytes,
		UpdateTS:    updateTS,
		EntityID:    entityID,
		DoerID:      doerID,
		Extra:       extra,
	}, nil
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

func (k *KeyInfo) Save(db *pttdb.LDBBatch) error {
	var err error
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

	_, err = db.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeyInfo) LoadNewest(db *pttdb.LDBBatch) error {
	iter, err := db.DB().NewPrevIteratorWithPrefix(nil, k.DBPrefix())
	if err != nil {
		return err
	}
	defer iter.Release()

	log.Debug("LoadNewest: to loop for", "DBPrefix", k.DBPrefix(), "iter", iter)

	var val []byte = nil
	for iter.Prev() {
		val = iter.Value()
		log.Debug("LoadNewest: in-loop", "val", val)
		break

	}
	if val == nil {
		return leveldb.ErrNotFound
	}

	err = k.Unmarshal(val)
	if err != nil {
		return err
	}

	k.Key, err = crypto.ToECDSA(k.KeyBytes)
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

func (k *KeyInfo) DeleteKey(key []byte, db *pttdb.LDBBatch) error {
	idxKey, err := k.KeyToIdxKey(key)
	if err != nil {
		return err
	}

	log.Debug("DeleteKey", "idxKey", idxKey)

	err = db.DeleteAll(idxKey)

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

	k.Key, err = crypto.ToECDSA(k.KeyBytes)
	if err != nil {
		return err
	}

	return nil
}
