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
	"sync"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

func TestCount_Add(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	c1, _ := NewCount(tDBOplog, &types.PttID{}, &types.PttID{}, tDBCountPrefix, tPCount, true)

	// define test-structure
	type fields struct {
		Bits       types.BitVector
		lock       sync.RWMutex
		hash       hash.Hash64
		db         *pttdb.LDBDatabase
		dbPrefixID *types.PttID
		dbPrefix   []byte
	}
	type args struct {
		item []byte
	}

	// prepare test-cases
	tests := []struct {
		name string
		c    *Count
		args args
	}{
		// TODO: Add test cases.
		{
			c:    c1,
			args: args{item: []byte("test")},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.c
			c.Add(tt.args.item)
		})
	}

	// teardown test
}

func TestCount_Count(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	c1, _ := NewCount(tDBOplog, &types.PttID{}, &types.PttID{}, tDBCountPrefix, tPCount, true)

	c1.Add([]byte("test"))

	c2, _ := NewCount(tDBOplog, &types.PttID{}, &types.PttID{}, tDBCountPrefix, tPCount, true)
	c2.Add([]byte("test"))
	c2.Add([]byte("test"))
	c2.Add([]byte("test"))

	c3, _ := NewCount(tDBOplog, &types.PttID{}, &types.PttID{}, tDBCountPrefix, tPCount, true)
	c3.Add([]byte("test"))
	c3.Add([]byte("test1"))
	c3.Add([]byte("test2"))

	c10, _ := NewCount(tDBOplog, &types.PttID{}, &types.PttID{}, tDBCountPrefix, tPCount, true)
	c10.Add([]byte("test1"))
	c10.Add([]byte("test2"))
	c10.Add([]byte("test3"))
	c10.Add([]byte("test4"))
	c10.Add([]byte("test5"))
	c10.Add([]byte("test6"))
	c10.Add([]byte("test7"))
	c10.Add([]byte("test8"))
	c10.Add([]byte("test19"))
	c10.Add([]byte("test91"))

	// define test-structure
	type fields struct {
		Bits       types.BitVector
		lock       sync.RWMutex
		hash       hash.Hash64
		db         *pttdb.LDBDatabase
		dbPrefixID *types.PttID
		dbPrefix   []byte
	}

	// prepare test-cases
	tests := []struct {
		name string
		c    *Count
		want uint64
	}{
		// TODO: Add test cases.
		{
			c:    c1,
			want: 1,
		},
		{
			c:    c2,
			want: 1,
		},
		{
			c:    c3,
			want: 3,
		},
		{
			c:    c10,
			want: 10,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.c
			if got := c.Count(); got != tt.want {
				t.Errorf("Count.Count() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
