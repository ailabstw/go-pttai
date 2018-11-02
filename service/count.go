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
	"hash"
	"math"
	"sync"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/spaolacci/murmur3"
)

type Count struct {
	Bits types.BitVector `json:"B"` // bitvector to hold the occupied buckets

	// IsFull types.Bool `json:"i"`

	lock sync.RWMutex
	hash hash.Hash64

	p uint
	m uint64

	db         *pttdb.LDBBatch
	dbPrefixID *types.PttID
	dbID       *types.PttID
	dbPrefix   []byte
}

func NewCount(db *pttdb.LDBBatch, dbPrefixID *types.PttID, dbID *types.PttID, dbPrefix []byte, p uint, isNewBits bool) (*Count, error) {
	c := &Count{}
	m := uint64(1 << p)
	c.m = m
	c.p = p

	c.SetHash()
	c.SetDB(db, dbPrefixID, dbID, dbPrefix)
	if isNewBits {
		c.NewBits()
	}

	return c, nil
}

func (c *Count) NewBits() {
	bits := types.NewBitVector(c.m)
	c.Bits = bits
}

func (c *Count) SetBits(bits types.BitVector) {
	c.Bits = bits
}

func (c *Count) SetHash() {
	c.hash = murmur3.New64()
}

func (c *Count) SetDB(db *pttdb.LDBBatch, dbPrefixID *types.PttID, dbID *types.PttID, dbPrefix []byte) {
	c.db = db
	c.dbPrefixID = dbPrefixID
	c.dbID = dbID
	c.dbPrefix = dbPrefix
}

func (c *Count) Save() error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	key, err := c.MarshalKey()
	if err != nil {
		return err
	}

	marshaled, err := c.Bits.Marshal()
	if err != nil {
		return err
	}

	err = c.db.DB().Put(key, marshaled)
	if err != nil {
		return err
	}

	return nil
}

func (c *Count) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{c.dbPrefix, c.dbPrefixID[:], c.dbID[:]})
}

func (c *Count) Load() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	key, err := c.MarshalKey()
	if err != nil {
		return err
	}

	val, err := c.db.DBGet(key)
	if err != nil {
		return err
	}
	bits, err := types.UnmarshalBitVector(val)
	if err != nil {
		return err
	}

	c.SetBits(bits)

	return nil
}

/*

Return: isNew
*/
func (c *Count) AddWithIsNew(item []byte) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.hash.Reset()
	c.hash.Write(item)
	hash := c.hash.Sum64()
	bucket := hash >> (64 - c.p) // top p bits are the bucket
	return c.Bits.SetWithIsNew(bucket)
}

/*

Return: isNew
*/
func (c *Count) Add(item []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.hash.Reset()
	c.hash.Write(item)
	hash := c.hash.Sum64()
	bucket := hash >> (64 - c.p) // top p bits are the bucket
	log.Debug("Add: to Set", "hash", hash, "bucket", bucket)
	c.Bits.Set(bucket)
}

// Distinct returns the estimate of the number of distinct elements seen
// if the backing BitVector is full it returns m, the size of the BitVector
func (c *Count) Count() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	zeroCount := c.m - c.Bits.PopCount()
	if zeroCount > 0 {
		return uint64(float64(c.m) * math.Log(float64(c.m)/float64(zeroCount)))
	}
	return (1 << c.p)
}

// Union the estimate of two LinearCounting reducing the precision to the minimum of the two sets
// the function will return nil and an error if the hash functions mismatch
func (c *Count) Union(c2 *Count) (*Count, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c2.lock.RLock()
	defer c2.lock.RUnlock()

	// for each bucket take the OR of the two LinearCounting
	combinedC, err := NewCount(c.db, c.dbPrefixID, c.dbID, c.dbPrefix, c.p, true)
	if err != nil {
		return nil, err
	}

	for i := range combinedC.Bits {
		combinedC.Bits[i] = c.Bits[i] | c2.Bits[i]
	}
	return combinedC, nil
}

// Intersect the estimate of two LinearCounting reducing the precision to the minimum of the two sets
// the function will return nil and an error if the hash functions mismatch
func (c *Count) Intersect(c2 *Count) (*Count, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c2.lock.RLock()
	defer c2.lock.RUnlock()

	// for each bucket take the AND of the two LinearCounting
	combinedC, err := NewCount(c.db, c.dbPrefixID, c.dbID, c.dbPrefix, c.p, true)
	if err != nil {
		return nil, err
	}
	for i := range combinedC.Bits {
		combinedC.Bits[i] = c.Bits[i] & c2.Bits[i]
	}
	return combinedC, nil
}
