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

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/crypto/bip32"
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

	return newKeyInfo(extendedKey, nil, entityID, nil)
}

func NewOpKeyInfo(entityID *types.PttID, doerID *types.PttID, masterKey *ecdsa.PrivateKey) (*KeyInfo, error) {
	key, extra, err := deriveOpKey(masterKey)
	if err != nil {
		return nil, err
	}

	return newKeyInfo(key, extra, entityID, doerID)
}

func NewSignKeyInfo(doerID *types.PttID, masterKey *ecdsa.PrivateKey) (*KeyInfo, error) {
	key, extra, err := deriveSignKey(masterKey)
	if err != nil {
		return nil, err
	}
	return newKeyInfo(key, extra, nil, doerID)
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

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	id := keyInfoHashToID(&hash)

	return &KeyInfo{
		BaseObject: NewObject(id, ts, doerID, entityID, nil, types.StatusInvalid),

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
