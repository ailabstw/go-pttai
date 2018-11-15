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

package pttdb

import (
	"reflect"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
)

type TestBase struct {
	V         types.Version
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`
	EntityID  *types.PttID    `json:"e,omitempty"`

	LogID       *types.PttID `json:"l,omitempty"`
	UpdateLogID *types.PttID `json:"u,omitempty"`

	Status types.Status `json:"S"`
}

type TestTemp struct {
	BaseObj  *TestBase       `json:"b"`
	UpdateTS types.Timestamp `json:"UT"`
}

func TestDBWithStatus_Unmarshal(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	dataBytes := []byte(`{"b":{"CT":{"T":1,"NT":5},"ID":"1234567890234567890","S":7},"UT":{"T":4,"NT":6}}`)

	expectD := &DBWithStatus{BaseObj: &DBStatus{Status: types.StatusAlive}, UpdateTS: types.Timestamp{4, 6}}

	// define test-structure
	type fields struct {
		BaseObj  *DBStatus
		UpdateTS types.Timestamp
	}
	type args struct {
		data []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		d       *DBWithStatus
		args    args
		want    *DBWithStatus
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			d:    &DBWithStatus{},
			args: args{dataBytes},
			want: expectD,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.d
			if err := d.Unmarshal(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("DBWithStatus.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("DBWithStatus: d: %v %v", d.BaseObj, d.UpdateTS)

			if !reflect.DeepEqual(d, tt.want) {
				t.Errorf("DBWithStatus.Unmarshal() d = %v, want = %v", d, tt.want)
			}
		})
	}

	// teardown test
}
