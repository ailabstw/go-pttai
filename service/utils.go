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
	"bytes"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"sort"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
)

func SignData(bytes []byte, keyInfo *KeyInfo) ([]byte, []byte, []byte, []byte, error) {
	salt, err := types.NewSalt()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bytesWithSalt, err := common.Concat([][]byte{bytes, salt[:]})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	hash := crypto.Keccak256(bytesWithSalt)
	sig, err := crypto.Sign(hash, keyInfo.Key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	keyInfo.Count++

	return bytesWithSalt, hash, sig, keyInfo.PubKeyBytes, nil
}

func VerifyData(bytesWithSalt []byte, sig []byte, pubKeyBytes []byte, doerID *types.PttID, extra *KeyExtraInfo) error {

	isValidKey := verifyDataCheckKey(pubKeyBytes, doerID, extra)
	if !isValidKey {
		return ErrInvalidKey
	}

	hash := crypto.Keccak256(bytesWithSalt)

	isGood := crypto.VerifySignature(pubKeyBytes, hash, sig[:64])
	if !isGood {
		return ErrInvalidData
	}
	return nil
}

func verifyDataCheckKey(pubKeyBytes []byte, doerID *types.PttID, extra *KeyExtraInfo) bool {
	if extra == nil {
		pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
		if err != nil {
			return false
		}
		return doerID.IsSamePubKey(pubKey)
	}

	return extra.IsValid(pubKeyBytes, doerID)
}

func DBPrefix(dbPrefix []byte, id *types.PttID) ([]byte, error) {
	return common.Concat([][]byte{dbPrefix, id[:]})
}

func GenChallenge() []byte {
	challenge := make([]byte, SizeChallenge)
	io.ReadFull(rand.Reader, challenge)

	return challenge
}

func NewPttIDWithMyID(myID *types.PttID) (*types.PttID, error) {
	return types.NewPttIDWithPostifx(myID[:common.AddressLength])
}

func NewPttIDWithMixedIDs(pttIDs []*types.PttID) (*types.PttID, error) {
	if len(pttIDs) == 0 {
		return nil, nil
	}

	bs := make([][]byte, len(pttIDs))

	for i, pttID := range pttIDs {
		bs[i] = pttID[:]
	}

	sort.SliceStable(bs, func(i int, j int) bool {
		return bytes.Compare(bs[i], bs[j]) < 0
	})

	hash := types.Hash(bs...)

	id := &types.PttID{}
	copy(id[:], hash)
	copy(id[common.AddressLength:], bs[0])

	return id, nil
}

func RandomPeer(peerList []*PttPeer) *PttPeer {
	lenPeerList := len(peerList)
	if lenPeerList == 0 {
		return nil
	}

	randNum := mrand.Intn(lenPeerList)
	return peerList[randNum]
}

func randNum(minNum int, maxNum int) int {
	return mrand.Intn(maxNum-minNum) + minNum
}
