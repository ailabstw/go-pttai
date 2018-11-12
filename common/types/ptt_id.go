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

package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"reflect"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/shengdoushi/base58"
)

var RandRead = func(b []byte) (int, error) {
	return rand.Read(b)
}

/*
PttID is the id-representation of all applications.

PttID is constructed from PKI-framework, with priv-key and pub-key.
PttID is the address-representation of pub-key.
*/
type PttID [SizePttID]byte

/*
NewPttID generates new PttID.
*/
func NewPttID() (*PttID, error) {
	p := &PttID{}

	nRead, err := RandRead(p[:])
	if err != nil {
		return nil, err
	}
	if nRead != SizePttID {
		return nil, err
	}

	return p, nil
}

/*
NewPttID generates new PttID.
*/
func NewPttIDWithPostifx(postfix []byte) (*PttID, error) {
	p := &PttID{}

	lenPostfix := len(postfix)
	lenRand := SizePttID - lenPostfix
	nRead, err := RandRead(p[:lenRand])
	if err != nil {
		return nil, err
	}
	if nRead != lenRand {
		return nil, err
	}
	copy(p[lenRand:], postfix)

	return p, nil
}

/*
NewPttIDFromKey generates new PttID with the 1st part as the key-address and the 2nd part as random.
*/
func NewPttIDFromKey(key *ecdsa.PrivateKey) (*PttID, error) {
	addr := crypto.PubkeyToAddress(key.PublicKey)
	p := &PttID{}
	nCopy := copy(p[:], addr[:])
	if nCopy != common.AddressLength {
		return nil, ErrInvalidID
	}

	nRead, err := RandRead(p[common.AddressLength:])
	if err != nil {
		return nil, ErrInvalidID
	}
	if nRead != common.AddressLength {
		return nil, ErrInvalidID
	}

	return p, nil
}

/*
NewPttIDFromKey generates new PttID with the 1st part as the key-address and the 2nd part as postfix.
*/
func NewPttIDFromKeyPostfix(key *ecdsa.PrivateKey, postfix []byte) (*PttID, error) {
	addr := crypto.PubkeyToAddress(key.PublicKey)

	return NewPttIDFromAddrPostfix(&addr, postfix)
}

/*
NewPttIDFromPubkeyPostfix generates new PttID with the 1st part as the key-address and the 2nd part as postfix.
*/
func NewPttIDFromPubkeyPostfix(key *ecdsa.PublicKey, postfix []byte) (*PttID, error) {
	addr := crypto.PubkeyToAddress(*key)

	return NewPttIDFromAddrPostfix(&addr, postfix)
}

/*
NewPttIDFromAddrPostfix generates new PttID with the 1st part as the key-address and the 2nd part as postfix.
*/
func NewPttIDFromAddrPostfix(addr *common.Address, postfix []byte) (*PttID, error) {
	p := &PttID{}
	nCopy := copy(p[:], addr[:])
	if nCopy != common.AddressLength {
		return nil, ErrInvalidID
	}

	nCopy = copy(p[common.AddressLength:], postfix)
	if nCopy != common.AddressLength {
		return nil, ErrInvalidID
	}

	return p, nil
}

/*
NewPttIDWithRefID generates new PttID with the 1st part as the key-address and the 2nd part as the addr-part of refID.
*/
func NewPttIDWithRefID(key *ecdsa.PrivateKey, refID *PttID) (*PttID, error) {

	return NewPttIDFromKeyPostfix(key, refID[:common.AddressLength])
}

/*
NewPttIDWithRefID generates new PttID with the 1st part as the key-address and the 2nd part as the addr-part of refID.
*/
func NewPttIDWithPubkeyAndRefID(key *ecdsa.PublicKey, refID *PttID) (*PttID, error) {
	return NewPttIDFromPubkeyPostfix(key, refID[:common.AddressLength])
}

func UnmarshalTextPttID(b []byte) (*PttID, error) {
	if len(b) == 0 {
		return nil, nil
	}

	p := &PttID{}

	err := p.UnmarshalText(b)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PttID) MarshalText() ([]byte, error) {
	pBytes, err := p.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return pBytes[1:(len(pBytes) - 1)], nil
}

func (p *PttID) UnmarshalText(b []byte) error {
	return p.UnmarshalJSON(b)
}

func (p *PttID) MarshalJSON() ([]byte, error) {
	marshaled := []byte(base58.Encode(p[:], myAlphabet))
	return common.Concat([][]byte{quoteBytes, marshaled, quoteBytes})
}

func (p *PttID) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return ErrInvalidID
	}

	if b[0] == '"' { // hack for possibly json-stringify strings
		b = b[1:(len(b) - 1)]
	}

	decodedBytes, err := base58.Decode(string(b), myAlphabet)
	if err != nil {
		return err
	}

	if len(decodedBytes) != SizePttID {
		return ErrInvalidID
	}

	copy(p[:], decodedBytes)

	return nil
}

// Serialize is used for direct request/response
func (p *PttID) Marshal() ([]byte, error) {
	return p[:], nil
}

// Deserialize is used for direct request/response
func (p *PttID) Unmarshal(theIDStr []byte) error {
	if len(theIDStr) != SizePttID {
		return ErrInvalidID
	}
	copy(p[:], theIDStr)
	return nil
}

func (p *PttID) IsSameKey(key *ecdsa.PrivateKey) bool {
	return p.IsSamePubKey(&key.PublicKey)
}

func (p *PttID) IsSamePubKey(key *ecdsa.PublicKey) bool {
	addr := crypto.PubkeyToAddress(*key)
	return reflect.DeepEqual(p[:common.AddressLength], addr[:])
}

func (p *PttID) IsSameKeyWithPttID(p2 *PttID) bool {
	if p2 == nil {
		return false
	}

	return reflect.DeepEqual(p[:common.AddressLength], p2[:common.AddressLength])
}

func (p *PttID) IsSameKeyWithNodeID(n *discover.NodeID) bool {
	if n == nil {
		return false
	}

	return reflect.DeepEqual(p[:common.AddressLength], n[:common.AddressLength])
}

func (p *PttID) Clone() *PttID {
	p2 := &PttID{}
	copy(p2[:], p[:])
	return p2
}
