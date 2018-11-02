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
)

func TestSignData(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		bytes   []byte
		keyInfo *KeyInfo
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    []byte
		want1   []byte
		want2   []byte
		want3   []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{bytes: tDefaultBytes2, keyInfo: tDefaultSignKeyInfo2},
			want:    tDefaultBytesWithSalt2,
			want1:   tDefaultHash2,
			want2:   tDefaultSig2,
			want3:   tDefaultPubBytes2,
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := SignData(tt.args.bytes, tt.args.keyInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SignData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SignData() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("SignData() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("SignData() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}

	// teardown test
}

func TestVerifyData(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		bytesWithSalt []byte
		sig           []byte
		keyBytes      []byte
		doerID        *types.PttID
		extra         *KeyExtraInfo
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{bytesWithSalt: tDefaultBytesWithSalt2, sig: tDefaultSig2, keyBytes: tDefaultPubBytes2, doerID: tDefaultSignKeyInfo2.CreatorID},
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyData(tt.args.bytesWithSalt, tt.args.sig, tt.args.keyBytes, tt.args.doerID, tt.args.extra); (err != nil) != tt.wantErr {
				t.Errorf("VerifyData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}
