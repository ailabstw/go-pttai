// Copyright 2019 The go-pttai Authors
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
	"encoding/binary"
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type ContentBlock struct {
	BlockID uint32   `json:"ID"`
	Buf     [][]byte `json:"B,omitempty"`
}

func NewContentBlock(blockID uint32, buf [][]byte) *ContentBlock {
	return &ContentBlock{
		BlockID: blockID,
		Buf:     buf,
	}
}

type Block struct {
	V          types.Version
	ID         *types.PttID
	BlockID    uint32 `json:"bID"`
	SubBlockID uint8  `json:"sID"`

	Buf []byte `json:"B,omitempty"`

	ObjID *types.PttID `json:"o,omitempty"`

	Hash     []byte        `json:"H,omitempty"`
	Salt     types.Salt    `json:"s,omitempty"`
	Sig      []byte        `json:"S,omitempty"`
	Pub      []byte        `json:"K,omitempty"`
	KeyExtra *KeyExtraInfo `json:"k,omitempty"`

	db           *pttdb.LDBBatch
	fullDBPrefix []byte
}

func NewEmptyBlock() *Block {
	return &Block{}
}

func NewBlock(
	blockID uint32,
	subBlockID uint8,
	buf []byte,

) (*Block, error) {
	return &Block{
		V: types.CurrentVersion,

		BlockID:    blockID,
		SubBlockID: subBlockID,

		Buf: buf,
	}, nil
}

func (b *Block) SetDB(
	db *pttdb.LDBBatch,
	fullDBPrefix []byte,

	objID *types.PttID,
	id *types.PttID,

) {
	b.V = types.CurrentVersion
	b.db = db
	b.fullDBPrefix = fullDBPrefix
	b.ObjID = objID
	b.ID = id
}

func (b *Block) Save() error {
	if b.db == nil {
		return ErrInvalidBlock
	}

	key, err := b.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := b.Marshal()
	if err != nil {
		return err
	}

	log.Debug("Block.Save: to put", "key", key)
	err = b.db.DB().Put(key, marshaled)
	if err != nil {
		return err
	}

	return nil
}

func (b *Block) RemoveAll() error {
	iter, err := b.GetIter(pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		b.db.DBDelete(key)
	}

	return nil
}

func (b *Block) GetIter(listOrder pttdb.ListOrder, isLocked bool) (iterator.Iterator, error) {
	prefix, err := b.Prefix()
	if err != nil {
		return nil, err
	}

	log.Debug("Block.GetIter: to new-iterator", "prefix", prefix)
	return b.db.DB().NewIteratorWithPrefix(nil, prefix, listOrder)
}

func (b *Block) Prefix() ([]byte, error) {
	if b.ID == nil {
		return common.Concat([][]byte{b.fullDBPrefix, b.ObjID[:]})
	}
	return common.Concat([][]byte{b.fullDBPrefix, b.ObjID[:], b.ID[:]})
}

func (b *Block) MarshalKey() ([]byte, error) {
	marshaledBlockID := make([]byte, 4) // uint32
	binary.BigEndian.PutUint32(marshaledBlockID, b.BlockID)

	marshaledSubBlockID := []byte{b.SubBlockID} // uint8

	return common.Concat([][]byte{b.fullDBPrefix, b.ObjID[:], b.ID[:], marshaledBlockID, marshaledSubBlockID})
}

func (b *Block) Marshal() ([]byte, error) {
	return json.Marshal(b)
}

func (b *Block) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, b)
}

func (b *Block) Sign(key *KeyInfo) error {
	b.Hash = nil
	b.Salt = types.Salt{}
	b.Sig = nil
	b.Pub = nil
	b.KeyExtra = nil

	marshaled, err := b.Marshal()
	if err != nil {
		return err
	}

	bytesWithSalts, hash, sig, pubBytes, err := SignData(marshaled, key)
	if err != nil {
		return err
	}

	b.Hash = hash
	copy(b.Salt[:], bytesWithSalts[len(marshaled):])
	b.Sig = sig
	b.Pub = pubBytes
	b.KeyExtra = key.Extra

	log.Debug("Block.Sign: to return", "blockID", b.BlockID, "subBlockID", b.SubBlockID, "hash", hash, "sig", sig, "salt", b.Salt)

	return nil
}

func (b *Block) Verify(expectedHash []byte, creatorID *types.PttID) error {
	origHash, origSalt, origSig, origPub, origKeyExtra := b.Hash, b.Salt, b.Sig, b.Pub, b.KeyExtra
	defer func() {
		b.Hash, b.Salt, b.Sig, b.Pub, b.KeyExtra = origHash, origSalt, origSig, origPub, origKeyExtra
	}()

	b.Hash = nil
	b.Salt = types.Salt{}
	b.Sig = nil
	b.Pub = nil
	b.KeyExtra = nil

	marshaled, err := b.Marshal()
	if err != nil {
		return err
	}

	bytesWithSalt := append(marshaled, origSalt[:]...)

	log.Debug("Block.Verify: to verify", "blockID", b.BlockID, "subBlockID", b.SubBlockID, "hash", origHash, "expectedHash", expectedHash, "sig", origSig, "salt", origSalt)

	return VerifyData(bytesWithSalt, expectedHash, origSig, origPub, creatorID, origKeyExtra)
}
