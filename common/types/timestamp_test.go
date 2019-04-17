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

func TestTimestamp_IsEqMilli(t *testing.T) {
	// setup test

	// define test-structure
	type fields struct {
		Ts     int64
		NanoTs uint32
	}
	type args struct {
		t2 Timestamp
	}

	// prepare test-cases
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(1000000),
			},
			args: args{t2: Timestamp{Ts: 1, NanoTs: 1000001}},
			want: true,
		},
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(1999999),
			},
			args: args{t2: Timestamp{Ts: 1, NanoTs: 1000001}},
			want: true,
		},
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(2000000),
			},
			args: args{t2: Timestamp{Ts: 1, NanoTs: 1000001}},
			want: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := &Timestamp{
				Ts:     tt.fields.Ts,
				NanoTs: tt.fields.NanoTs,
			}
			if got := tmp.IsEqMilli(tt.args.t2); got != tt.want {
				t.Errorf("Timestamp.IsEqMilli() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestTimestamp_Marshal(t *testing.T) {
	// setup test

	// define test-structure
	type fields struct {
		Ts     int64
		NanoTs uint32
	}

	// prepare test-cases
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{1234567890, 999999999},
			want:   []byte{0, 0, 0, 0, 73, 150, 2, 210, 59, 154, 201, 255},
		},
		{
			fields: fields{9999999999, 999999999},
			want:   []byte{0, 0, 0, 2, 84, 11, 227, 255, 59, 154, 201, 255},
		},
		{
			fields: fields{9999999999, 1},
			want:   []byte{0, 0, 0, 2, 84, 11, 227, 255, 0, 0, 0, 1},
		},
		{
			fields: fields{1, 2},
			want:   []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2},
		},
		{
			fields: fields{772, 258},
			want:   []byte{0, 0, 0, 0, 0, 0, 3, 4, 0, 0, 1, 2},
		},
		{
			fields: fields{253, 254},
			want:   []byte{0, 0, 0, 0, 0, 0, 0, 253, 0, 0, 0, 254},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := &Timestamp{
				Ts:     tt.fields.Ts,
				NanoTs: tt.fields.NanoTs,
			}
			got, err := tmp.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Timestamp.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Timestamp.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestUnmarshalTimestamp(t *testing.T) {
	// setup test

	// define test-structure
	type args struct {
		theBytes []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    Timestamp
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{[]byte{0, 0, 0, 0, 73, 150, 2, 210, 59, 154, 201, 255}},
			want: Timestamp{1234567890, 999999999},
		},
		{
			args: args{[]byte{0, 0, 0, 2, 84, 11, 227, 255, 59, 154, 201, 255}},
			want: Timestamp{9999999999, 999999999},
		},
		{
			args: args{[]byte{0, 0, 0, 2, 84, 11, 227, 255, 0, 0, 0, 1}},
			want: Timestamp{9999999999, 1},
		},
		{
			args: args{[]byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2}},
			want: Timestamp{1, 2},
		},
		{
			args: args{[]byte{0, 0, 0, 0, 0, 0, 3, 4, 0, 0, 1, 2}},
			want: Timestamp{772, 258},
		},
		{
			args: args{[]byte{0, 0, 0, 0, 0, 0, 0, 253, 0, 0, 0, 254}},
			want: Timestamp{253, 254},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalTimestamp(tt.args.theBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestGetTimestamp(t *testing.T) {
	// setup test

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		want    Timestamp
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetTimestamp()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			/*
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetTimestamp() = %v, want %v", got, tt.want)
				}
			*/
		})
	}

	// teardown test
}

func TestTimestampEq(t *testing.T) {
	// setup test

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		t1      Timestamp
		t2      Timestamp
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			t1:   Timestamp{1, 1},
			t2:   Timestamp{1, 1},
			want: true,
		},
		{
			t1:   Timestamp{1, 1},
			t2:   Timestamp{1, 2},
			want: false,
		},
		{
			t1:   Timestamp{1, 1},
			t2:   Timestamp{2, 1},
			want: false,
		},
		{
			t1:   Timestamp{1, 1},
			t2:   Timestamp{2, 2},
			want: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t1 == tt.t2
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TimestampEq() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestTimestamp_ToMilli(t *testing.T) {
	// setup test

	// define test-structure
	type fields struct {
		Ts     int64
		NanoTs uint32
	}

	// prepare test-cases
	tests := []struct {
		name   string
		fields fields
		want   Timestamp
	}{
		// TODO: Add test cases.
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(1000000),
			},
			want: Timestamp{Ts: 1, NanoTs: 1000000},
		},
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(1000001),
			},
			want: Timestamp{Ts: 1, NanoTs: 1000000},
		},
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(1999999),
			},
			want: Timestamp{Ts: 1, NanoTs: 1000000},
		},
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(2000000),
			},
			want: Timestamp{Ts: 1, NanoTs: 2000000},
		},
		{
			fields: fields{
				Ts:     int64(1),
				NanoTs: uint32(999999999),
			},
			want: Timestamp{Ts: 1, NanoTs: 999000000},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t2 := &Timestamp{
				Ts:     tt.fields.Ts,
				NanoTs: tt.fields.NanoTs,
			}
			if got := t2.ToMilli(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Timestamp.ToMilli() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
