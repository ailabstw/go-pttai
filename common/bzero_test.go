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

func TestBzero(t *testing.T) {
	// setup test

	// define test-structure
	type args struct {
		buf []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  []byte
	}{
		// TODO: Add test cases.
		{
			args:   args{buf: []byte{1, 2, 3, 4}},
			expect: []byte{0, 0, 0, 0},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Bzero(tt.args.buf); (err != nil) != tt.wantErr {
				t.Errorf("Bzero() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.buf, tt.expect) {
				t.Errorf("Bzero() args.buf: %v, expect: %v", tt.args.buf, tt.expect)
			}
		})
	}

	// teardown test
}
