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

func TestBaseEntity_AddOwnerID(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tID := &types.PttID{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	}

	tID2 := &types.PttID{
		1, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	}

	tBaseEntity := &BaseEntity{OwnerIDs: []*types.PttID{tID2}}

	tBaseEntity2 := &BaseEntity{}

	tBaseEntity3 := &BaseEntity{OwnerIDs: []*types.PttID{tID}}

	// define test-structure
	type fields struct {
		V           types.Version
		ID          *types.PttID
		CreateTS    types.Timestamp
		CreatorID   *types.PttID
		UpdaterID   *types.PttID
		LogID       *types.PttID
		UpdateLogID *types.PttID
		Status      types.Status
		OwnerIDs    []*types.PttID
		EntityType  EntityType
		pm          ProtocolManager
		name        string
		ptt         Ptt
		service     Service
		db          *pttdb.LDBBatch
		dbLock      *types.LockMap
		SyncInfo    SyncInfo
	}
	type args struct {
		id *types.PttID
	}

	// prepare test-cases
	tests := []struct {
		name string
		e    *BaseEntity
		args args
		want []*types.PttID
	}{
		// TODO: Add test cases.
		{
			e:    tBaseEntity,
			args: args{id: tID},
			want: []*types.PttID{tID, tID2},
		},
		{
			e:    tBaseEntity2,
			args: args{id: tID},
			want: []*types.PttID{tID},
		},
		{
			e:    tBaseEntity3,
			args: args{id: tID2},
			want: []*types.PttID{tID, tID2},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.e
			e.AddOwnerID(tt.args.id)

			got := e.OwnerIDs
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseEntity.AddOwnerID() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
