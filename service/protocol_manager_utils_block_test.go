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
	"reflect"
	"testing"
)

func TestScrambleBuf(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		buf [][]byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{buf: tDefaultBuf},
			want: tDefaultScrambleBuf,
		},
		{
			args: args{buf: tDefaultBuf2},
			want: tDefaultScrambleBuf2,
		},
		{
			args: args{buf: tDefaultBuf3},
			want: tDefaultScrambleBuf3,
		},
		{
			args: args{buf: tDefaultBuf4},
			want: tDefaultScrambleBuf4,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScrambleBuf(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScrambleBuf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScrambleBuf() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestUnscrambleBuf(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		buf [][]byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{buf: tDefaultScrambleBuf},
			want: tDefaultBuf,
		},
		{
			args: args{buf: tDefaultScrambleBuf2},
			want: tDefaultBuf2,
		},
		{
			args: args{buf: tDefaultScrambleBuf3},
			want: tDefaultBuf3,
		},
		{
			args: args{buf: tDefaultScrambleBuf4},
			want: tDefaultBuf4,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnscrambleBuf(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnscrambleBuf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnscrambleBuf() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
