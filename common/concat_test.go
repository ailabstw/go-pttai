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

package common

import (
	"reflect"
	"testing"
)

func TestConcat(t *testing.T) {
	// setup test
	a := []byte("a")
	b := []byte("b")
	c := []byte("c")

	// define test-structure
	type args struct {
		slices [][]byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{[][]byte{a, b, c}},
			want: []byte("abc"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Concat(tt.args.slices)
			if (err != nil) != tt.wantErr {
				t.Errorf("Concat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Concat() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
