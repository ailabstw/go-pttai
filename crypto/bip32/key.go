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

// ISC License
//
// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2016-2017 The Lightning Network Developers
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// This file is based on the following implementation:
// https://github.com/btcsuite/btcutil/blob/master/hdkeychain/extendedkey.go

package bip32

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"math/big"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
)

type ExtendedKey struct {
	key       []byte // This will be the pubkey for extended pub keys
	pubKey    []byte // This will only be set for extended priv keys
	chainCode []byte
	childNum  uint32
	isPrivate bool
}

func (k *ExtendedKey) Child(idx uint32) (*ExtendedKey, error) {
	pubBytes := k.PubkeyBytes()
	keyLen := 33
	data := make([]byte, keyLen+4)
	copy(data, pubBytes[:keyLen])
	binary.BigEndian.PutUint32(data[keyLen:], idx)

	hmac512 := hmac.New(sha512.New, k.chainCode)
	hmac512.Write(data)
	ilr := hmac512.Sum(nil)

	il := ilr[:len(ilr)/2]
	childChainCode := ilr[len(ilr)/2:]

	ilNum := new(big.Int).SetBytes(il)

	N := crypto.S256().Params().N

	if ilNum.Cmp(N) >= 0 || ilNum.Sign() == 0 {
		return nil, ErrInvalidChild
	}

	var isPrivate bool
	var childKey []byte

	if k.isPrivate {
		keyNum := new(big.Int).SetBytes(k.key)
		ilNum.Add(ilNum, keyNum)
		ilNum.Mod(ilNum, N)
		childKey = ilNum.Bytes()

		if len(childKey)*8 != crypto.BitSize { // XXX We need valid length
			return nil, ErrInvalidChild
		}

		isPrivate = true
	} else {
		ilx, ily := crypto.S256().ScalarBaseMult(il)
		if ilx.Sign() == 0 || ily.Sign() == 0 {
			return nil, ErrInvalidChild
		}

		pubKey, err := crypto.UnmarshalPubkey(k.key)
		if err != nil {
			return nil, err
		}

		childX, childY := crypto.S256().Add(ilx, ily, pubKey.X, pubKey.Y)
		pk := &ecdsa.PublicKey{Curve: crypto.S256(), X: childX, Y: childY}
		childKey = crypto.FromECDSAPub(pk)
	}

	return &ExtendedKey{key: childKey, chainCode: childChainCode, childNum: idx, isPrivate: isPrivate}, nil
}

func (k *ExtendedKey) ToPrivkey() (*ecdsa.PrivateKey, error) {
	if !k.isPrivate {
		return nil, ErrInvalidKey
	}

	return crypto.ToECDSA(k.key)
}

func (k *ExtendedKey) ToPubkey() (*ecdsa.PublicKey, error) {
	pubBytes := k.PubkeyBytes()

	return crypto.UnmarshalPubkey(pubBytes)
}

func (k *ExtendedKey) PrivkeyBytes() ([]byte, error) {
	if !k.isPrivate {
		return nil, ErrInvalidKey
	}

	return k.key, nil
}

func (k *ExtendedKey) PubkeyBytes() []byte {
	if !k.isPrivate {
		return k.key
	}

	if len(k.pubKey) != 0 {
		return k.pubKey
	}

	key, _ := crypto.ToECDSA(k.key) // extended-key must be valid key

	pubKey := crypto.FromECDSAPub(&key.PublicKey)
	k.pubKey = pubKey

	return pubKey
}

func Key(masterKey *ecdsa.PrivateKey, idx uint32) (*ExtendedKey, *types.Salt, error) {
	salt, err := types.NewSalt()
	if err != nil {
		return nil, nil, err
	}

	extendedMasterKey, err := PrivKeyToExtKey(masterKey, salt[:])
	if err != nil {
		return nil, nil, err
	}

	extendedDerivedKey, err := extendedMasterKey.Child(idx)
	if err != nil {
		return nil, nil, err
	}

	return extendedDerivedKey, salt, nil
}

func PrivKeyToExtKey(key *ecdsa.PrivateKey, chainCode []byte) (*ExtendedKey, error) {
	keyBytes := crypto.FromECDSA(key)
	return &ExtendedKey{
		key:       keyBytes,
		chainCode: chainCode,
		isPrivate: true,
	}, nil
}

func PubKeyToExtKey(pub *ecdsa.PublicKey, chainCode []byte) (*ExtendedKey, error) {
	keyBytes := crypto.FromECDSAPub(pub)
	return &ExtendedKey{
		key:       keyBytes,
		chainCode: chainCode,
		isPrivate: false,
	}, nil
}
