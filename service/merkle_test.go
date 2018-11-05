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
	"reflect"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

func TestMerkle_SaveMerkleTree(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tDefaultOplog.Save(true)
	tDefaultOplog2.Save(true)

	// define test-structure
	type fields struct {
		DBPrefix       []byte
		DBMerklePrefix []byte
		PrefixID       *types.PttID
		db             *pttdb.LDBBatch
	}
	type args struct {
		ts types.Timestamp
	}

	// prepare test-cases
	tests := []struct {
		name    string
		m       *Merkle
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			m:    tDefaultMerkle,
			args: args{types.Timestamp{Ts: 1234567890, NanoTs: 0}},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			if err := m.SaveMerkleTree(tt.args.ts); (err != nil) != tt.wantErr {
				t.Errorf("Merkle.SaveMerkleTree() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestMerkle_GetMerkleTreeList(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tDefaultOplog.Save(true)
	tDefaultOplog2.Save(true)
	tDefaultMerkle.SaveMerkleTree(types.Timestamp{Ts: 1234567890, NanoTs: 0})

	// define test-structure
	type fields struct {
		DBPrefix       []byte
		DBMerklePrefix []byte
		PrefixID       *types.PttID
		db             *pttdb.LDBBatch
	}
	type args struct {
		ts types.Timestamp
	}

	// prepare test-cases
	tests := []struct {
		name    string
		m       *Merkle
		args    args
		want    []*MerkleNode
		want1   []*MerkleNode
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			m:     tDefaultMerkle,
			args:  args{types.Timestamp{Ts: 1234567892, NanoTs: 0}},
			want:  []*MerkleNode{},
			want1: []*MerkleNode{tDefaultMerkleNode1Now, tDefaultMerkleNode2Now},
		},
		{
			m:     tDefaultMerkle,
			args:  args{types.Timestamp{Ts: 1234571490, NanoTs: 0}}, // +3600
			want:  []*MerkleNode{tDefaultMerkleNodeDay},
			want1: []*MerkleNode{},
		},
		{
			m:     tDefaultMerkle,
			args:  args{types.Timestamp{Ts: 1234654290, NanoTs: 0}}, // +86400
			want:  []*MerkleNode{tDefaultMerkleNodeDay},
			want1: []*MerkleNode{},
		},
		{
			m:     tDefaultMerkle,
			args:  args{types.Timestamp{Ts: 1237332690, NanoTs: 0}}, // +86400 * 32
			want:  []*MerkleNode{tDefaultMerkleNodeMonth},
			want1: []*MerkleNode{},
		},
		{
			m:     tDefaultMerkle,
			args:  args{types.Timestamp{Ts: 1266535890, NanoTs: 0}}, // +86400 * 370
			want:  []*MerkleNode{tDefaultMerkleNodeYear},
			want1: []*MerkleNode{},
		},
	}

	iter, _ := tDBOplog.DB().NewIteratorWithPrefix(tDBOplogMerklePrefix, tDBOplogMerklePrefix, pttdb.ListOrderNext)
	defer iter.Release()

	i := 0
	for iter.Next() {
		t.Logf("TestMerkle_GetMerkleTreeList: (in-loop): i: %d key: %v", i, iter.Key())
		i++
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			got, got1, err := m.GetMerkleTreeList(tt.args.ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merkle.GetMerkleTreeList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merkle.GetMerkleTreeList() = %v, want %v", got[0], tt.want[0])
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Merkle.GetMerkleTreeList() = %v, want %v", got1[1], tt.want1[1])
				//t.Errorf("Merkle.GetMerkleTreeList() = %v, want1 %v", got1[0], tt.want1[0])
			}
		})
	}

	// teardown test
}
