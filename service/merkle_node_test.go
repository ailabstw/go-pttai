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
)

func TestMerkleNode_MarshalJSON(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		Level     MerkleTreeLevel
		Hash      []byte
		UpdateTS  types.Timestamp
		NChildren uint32
	}

	// prepare test-cases
	tests := []struct {
		name    string
		m       *MerkleNode
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			m:    tDefaultMerkleNode,
			want: tDefaultMerkleNodeBytes,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			got, err := m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MerkleNode.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MerkleNode.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestMerkleNode_UnmarshalJSON(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		Level     MerkleTreeLevel
		Hash      []byte
		UpdateTS  types.Timestamp
		NChildren uint32
	}
	type args struct {
		b []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		m       *MerkleNode
		args    args
		want    *MerkleNode
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{b: tDefaultMerkleNodeBytes},
			m:    &MerkleNode{},
			want: tDefaultMerkleNode,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			if err := m.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("MerkleNode.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, tt.want) {
				t.Errorf("MerkleNode.UnmarshalJSON() = %v, want %v", m, tt.want)
			}
		})
	}

	// teardown test
}
