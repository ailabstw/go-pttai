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

package types

import (
	"reflect"
	"testing"
)

func TestBoolDoubleArray_MarshalJSON(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure

	bdary := make([][]Bool, 3)
	for i := 0; i < 3; i++ {
		bdary[i] = make([]Bool, 2)
		bdary[i][0] = true
		bdary[i][1] = false
	}

	// prepare test-cases
	tests := []struct {
		name    string
		bdary   BoolDoubleArray
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			bdary: BoolDoubleArray(bdary),
			want:  []byte{50, 49, 48, 49, 48, 49, 48},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bdary.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("BoolDoubleArray.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BoolDoubleArray.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestBoolDoubleArray_UnmarshalJSON(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	bdary := make([][]Bool, 3)
	for i := 0; i < 3; i++ {
		bdary[i] = make([]Bool, 5)
		bdary[i][0] = true
		bdary[i][1] = false
		bdary[i][2] = true
		bdary[i][3] = true
		bdary[i][4] = false
	}

	expected := BoolDoubleArray(bdary)

	// define test-structure
	type args struct {
		theBytes []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		bdary   BoolDoubleArray
		args    args
		want    BoolDoubleArray
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			bdary: BoolDoubleArray{},
			args: args{theBytes: []byte{53,
				49, 48, 49, 49, 48,
				49, 48, 49, 49, 48,
				49, 48, 49, 49, 48,
			}},
			want: expected,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bdary.UnmarshalJSON(tt.args.theBytes); (err != nil) != tt.wantErr {
				t.Errorf("BoolDoubleArray.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.bdary, tt.want) {
				t.Errorf("BoolDoubleArray.UnmarshalJSON() bdary = %v, want %v", tt.bdary, tt.want)

			}
		})
	}

	// teardown test
}

func TestBoolDoubleArray_Access(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	bdary := make([][]Bool, 3)
	for i := 0; i < 3; i++ {
		bdary[i] = make([]Bool, 5)
		bdary[i][0] = true
		bdary[i][1] = false
		bdary[i][2] = true
		bdary[i][3] = true
		bdary[i][4] = false
	}

	expected := BoolDoubleArray(bdary)

	// define test-structure
	type args struct {
		i int
		j int
	}

	// prepare test-cases
	tests := []struct {
		name    string
		bdary   BoolDoubleArray
		args    args
		want    Bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			bdary: expected,
			args:  args{i: 2, j: 2},
			want:  Bool(true),
		},
		{
			bdary: expected,
			args:  args{i: 2, j: 1},
			want:  Bool(false),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.bdary[tt.args.i][tt.args.j]

			if val != tt.want {
				t.Errorf("BoolDoubleArray.UnmarshalJSON() bdary = %v, want %v", tt.bdary, tt.want)

			}
		})
	}

	// teardown test
}
