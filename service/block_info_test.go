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
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

func TestBlockInfo_SetIsAllGood(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		V            types.Version
		ID           *types.PttID
		NBlock       int
		Hashs        [][][]byte
		IsGood       types.BoolDoubleArray
		MediaIDs     []*types.PttID
		UpdaterID    *types.PttID
		db           *pttdb.LDBBatch
		dbLock       *types.LockMap
		fullDBPrefix []byte
		objID        *types.PttID
		setMediaDB   func(media *Media)
	}

	// prepare test-cases
	tests := []struct {
		name string
		b    *BlockInfo
	}{
		// TODO: Add test cases.
		{
			b: &BlockInfo{},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.b
			b.SetIsAllGood()

			if !b.GetIsAllGood() {
				t.Logf("TestBlockInfo_SetIsAllGood: not is all good")
			}
		})
	}

	// teardown test
}
