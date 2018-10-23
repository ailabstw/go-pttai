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

package types

import (
	"crypto/ecdsa"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ailabstw/go-pttai/p2p/discover"
)

func TestNewPttID(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		want    *PttID
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPttID()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPttID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	// teardown test
}

func TestPttID_MarshalText(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		p       *PttID
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			p:    &tDefaultPttID,
			want: []byte("3uTox6ig7oRwvFGCBiq9eTm9PyXxj572mYXAfpsfjeeL1n7nZzV9Wr"),
		},
		{
			p:    &EmptyID,
			want: []byte("1111111111111111111111111111111111111111"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("PttID.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PttID.MarshalText() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_UnmarshalText(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		b []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		p       *PttID
		args    args
		want    *PttID
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			p:    &PttID{},
			args: args{b: []byte("3uTox6ig7oRwvFGCBiq9eTm9PyXxj572mYXAfpsfjeeL1n7nZzV9Wr")},
			want: &tDefaultPttID,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.UnmarshalText(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("PttID.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.p, tt.want) {
				t.Errorf("PttID.Unmarshal() p = %v, want %v", tt.p, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_MarshalJSON(t *testing.T) {
	// setup test

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		p       *PttID
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			p:    &tDefaultPttID,
			want: []byte("\"3uTox6ig7oRwvFGCBiq9eTm9PyXxj572mYXAfpsfjeeL1n7nZzV9Wr\""),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("PttID.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PttID.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_UnmarshalJSON(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		b []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		p       *PttID
		args    args
		want    *PttID
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			p:    &PttID{},
			args: args{b: []byte("\"3uTox6ig7oRwvFGCBiq9eTm9PyXxj572mYXAfpsfjeeL1n7nZzV9Wr\"")},
			want: &tDefaultPttID,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PttID{}
			err := json.Unmarshal(tt.args.b, p)
			if (err != nil) != tt.wantErr {
				t.Errorf("PttID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(p, tt.want) {
				t.Errorf("PttID.UnmarshalJSON() p = %v, want %v", p, tt.want)
			}
		})
	}

	// teardown test
}

func TestNewPttIDFromKey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		key *ecdsa.PrivateKey
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    *PttID
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{key: tDefaultKey2},
			want: &tDefaultPttID2,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPttIDFromKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPttIDFromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPttIDFromKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_IsSameKey(t *testing.T) {
	// setup test

	// define test-structure
	type args struct {
		key *ecdsa.PrivateKey
	}

	// prepare test-cases
	tests := []struct {
		name string
		p    *PttID
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			p:    &tDefaultPttID2,
			args: args{key: tDefaultKey2},
			want: true,
		},
		{
			p:    &tDefaultPttID,
			args: args{key: tDefaultKey2},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsSameKey(tt.args.key); got != tt.want {
				t.Errorf("PttID.IsSameKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_IsSamePubKey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		key *ecdsa.PublicKey
	}

	// prepare test-cases
	tests := []struct {
		name string
		p    *PttID
		args args
		want bool
	}{
		// TODO: Add test cases.
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsSamePubKey(tt.args.key); got != tt.want {
				t.Errorf("PttID.IsSamePubKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_IsSameKeyWithPttID(t *testing.T) {
	// setup test

	// define test-structure
	type args struct {
		p2 *PttID
	}

	// prepare test-cases
	tests := []struct {
		name string
		p    *PttID
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			p:    &tDefaultPttID2,
			args: args{p2: &tDefaultPttID3},
			want: true,
		},
		{
			p:    &tDefaultPttID2,
			args: args{p2: &tDefaultPttID},
			want: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsSameKeyWithPttID(tt.args.p2); got != tt.want {
				t.Errorf("PttID.IsSameKeyWithPttID() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPttID_IsSameKeyWithNodeID(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		n *discover.NodeID
	}

	// prepare test-cases
	tests := []struct {
		name string
		p    *PttID
		args args
		want bool
	}{
		// TODO: Add test cases.
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsSameKeyWithNodeID(tt.args.n); got != tt.want {
				t.Errorf("PttID.IsSameKeyWithNodeID() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
