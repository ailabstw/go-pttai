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
	"crypto/ecdsa"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"sort"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/pttdb"
)

func SignData(bytes []byte, key *ecdsa.PrivateKey) ([]byte, []byte, []byte, []byte, error) {
	salt, err := types.NewSalt()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bytesWithSalt, err := common.Concat([][]byte{bytes, salt[:]})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	hash := crypto.Keccak256(bytesWithSalt)
	sig, err := crypto.Sign(hash, key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	pubBytes := crypto.FromECDSAPub(&key.PublicKey)

	return bytesWithSalt, hash, sig, pubBytes, nil
}

func VerifyData(bytesWithSalt []byte, sig []byte, keyBytes []byte) error {
	hash := crypto.Keccak256(bytesWithSalt)

	isGood := crypto.VerifySignature(keyBytes, hash, sig[:64])
	if !isGood {
		return ErrInvalidData
	}
	return nil
}

/*
func SelectBestServicePeers(ps *ServicePeerSet, n int) ([]Peer, error) {
	peers := ps.Peers()
	lenPeers := len(peers)

	psList := make([]Peer, lenPeers)
	var i int = 0
	for _, p := range peers {
		psList[i] = p
		i++
	}
	psList, err := ShufflePeers(psList)
	if err != nil {
		return nil, err
	}

	var lenList int
	if lenPeers < n {
		lenList = lenPeers
	} else {
		lenList = n
	}

	return psList[:lenList], nil
}
*/

/*
func ShufflePeers(src []Peer) ([]Peer, error) {
	dest := make([]Peer, len(src))
	perm := mrand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}

	return dest, nil

}
*/

func DBPrefix(dbPrefix []byte, id *types.PttID) ([]byte, error) {
	return common.Concat([][]byte{dbPrefix, id[:]})
}

func GenChallenge() []byte {
	challenge := make([]byte, SizeChallenge)
	io.ReadFull(rand.Reader, challenge)

	return challenge
}

func NewPttIDWithKey(myID *types.PttID, db *pttdb.LDBDatabase) (*types.PttID, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	id, err := types.NewPttIDWithRefID(key, myID)
	if err != nil {
		return nil, err
	}

	keyBytes := crypto.FromECDSA(key)

	err = db.Put(id[:], keyBytes)
	if err != nil {
		return nil, err
	}

	return id, nil
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
