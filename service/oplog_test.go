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

func TestOplog_MarshalMerkleKey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		V              types.Version
		ID             *types.PttID
		DoerID         *types.PttID
		UpdateTS       types.Timestamp
		ObjID          *types.PttID
		Op             OpType
		Data           interface{}
		db             *pttdb.LDBBatch
		dbPrefixID     *types.PttID
		dbPrefix       []byte
		dbIdxPrefix    []byte
		dbMerklePrefix []byte
		Hash           []byte
		Salt           types.Salt
		Sig            []byte
		Pubkey         []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		o       *BaseOplog
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			o:    tDefaultOplog,
			want: tDefaultOplogMerkleKey,
		},
		{
			o:    tDefaultOplog2,
			want: tDefaultOplog2MerkleKey,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.o
			got, err := o.MarshalMerkleKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("Oplog.MarshalMerkleKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Oplog.MarshalMerkleKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
