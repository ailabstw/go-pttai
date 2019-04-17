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

// MIT License
//
// Copyright (c) 2017 bmkessler
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// https://github.com/bmkessler/streamstats/blob/master/bitvector.go

package types

import (
	"reflect"
	"testing"
)

func TestBitVector_PopCount(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name string
		b    BitVector
		want uint64
	}{
		// TODO: Add test cases.
		{
			b:    BitVector{1},
			want: 1,
		},
		{
			b:    BitVector{2},
			want: 1,
		},
		{
			b:    BitVector{4},
			want: 1,
		},
		{
			b:    BitVector{8},
			want: 1,
		},
		{
			b:    BitVector{16},
			want: 1,
		},
		{
			b:    BitVector{32},
			want: 1,
		},
		{
			b:    BitVector{64},
			want: 1,
		},
		{
			b:    BitVector{128},
			want: 1,
		},
		{
			b:    BitVector{256},
			want: 1,
		},
		{
			b:    BitVector{512},
			want: 1,
		},
		{
			b:    BitVector{1024},
			want: 1,
		},
		{
			b:    BitVector{2048},
			want: 1,
		},
		{
			b:    BitVector{4096},
			want: 1,
		},
		{
			b:    BitVector{8192},
			want: 1,
		},
		{
			b:    BitVector{16384},
			want: 1,
		},
		{
			b:    BitVector{32768},
			want: 1,
		},
		{
			b:    BitVector{65536},
			want: 1,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.PopCount(); got != tt.want {
				t.Errorf("BitVector.PopCount() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestUnmarshalBitVector(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	b := BitVector{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	theBytes, _ := b.Marshal()

	// define test-structure
	type args struct {
		theBytes []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    BitVector
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{theBytes},
			want: b,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalBitVector(tt.args.theBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBitVector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalBitVector() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
