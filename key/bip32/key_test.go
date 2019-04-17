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

package bip32

import (
	"crypto/ecdsa"
	"reflect"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
)

func TestExtendedKey_PubkeyBytes(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		key       []byte
		pubKey    []byte
		chainCode []byte
		childNum  uint32
		isPrivate bool
	}

	// prepare test-cases
	tests := []struct {
		name string
		key  *ExtendedKey
		want []byte
	}{
		// TODO: Add test cases.
		{
			key:  tDefaultExtendedKey,
			want: tDefaultPubKeyBytes,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.key

			if got := k.PubkeyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtendedKey.PubkeyBytes() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

/*
TestExtendedKey_Child tests validity of bip32 kdf.
	1. private-key => child-private-key
	2. pub-key => child-public-key
	3. private-key => pub-key (in globals_test)
	4. child-private-key => child-public-key (in globals_test)
*/
func TestExtendedKey_Child(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		key       []byte
		pubKey    []byte
		chainCode []byte
		childNum  uint32
		isPrivate bool
	}
	type args struct {
		idx uint32
	}

	// prepare test-cases
	tests := []struct {
		name    string
		key     *ExtendedKey
		args    args
		want    *ExtendedKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			key:  tDefaultExtendedKey,
			args: args{1},
			want: tChildExtendedKey,
		},
		{
			key:  tDefaultExtendedPubKey,
			args: args{1},
			want: tChildExtendedPubKey,
		},
		{
			key:  tDefaultExtendedKey2,
			args: args{1},
			want: tChildExtendedKey2,
		},
		{
			key:  tDefaultExtendedPubKey2,
			args: args{1},
			want: tChildExtendedPubKey2,
		},
		{
			key:  tDefaultExtendedKey2,
			args: args{2},
			want: tChildExtendedKey3,
		},
		{
			key:  tDefaultExtendedPubKey2,
			args: args{2},
			want: tChildExtendedPubKey3,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.key

			got, err := k.Child(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtendedKey.Child() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtendedKey.Child() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPrivKeyToExtKey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		key       *ecdsa.PrivateKey
		chainCode []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    *ExtendedKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{key: tDefaultKey, chainCode: tDefaultSalt[:]},
			want:    tDefaultExtendedKey,
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrivKeyToExtKey(tt.args.key, tt.args.chainCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrivKeyToExtKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrivKeyToExtKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPubKeyToExtKey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		pub       *ecdsa.PublicKey
		chainCode []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    *ExtendedKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{pub: tDefaultPubKey, chainCode: tDefaultSalt[:]},
			want:    tDefaultExtendedPubKey,
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PubKeyToExtKey(tt.args.pub, tt.args.chainCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("PubKeyToExtKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PubKeyToExtKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestKey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		masterKey *ecdsa.PrivateKey
		idx       uint32
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    *ExtendedKey
		want1   *types.Salt
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{masterKey: tDefaultKey, idx: 1},
			want:    tChildExtendedKey,
			want1:   tDefaultSalt,
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Key(tt.args.masterKey, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Key() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Key() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Key() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	// teardown test
}

func TestExtendedKey_PrivkeyBytes(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		key       []byte
		pubKey    []byte
		chainCode []byte
		childNum  uint32
		isPrivate bool
	}

	// prepare test-cases
	tests := []struct {
		name    string
		key     *ExtendedKey
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			key:     tDefaultExtendedKey,
			want:    tDefaultKeyBytes,
			wantErr: false,
		},
		{
			key:     tDefaultExtendedPubKey,
			want:    nil,
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.key
			got, err := k.PrivkeyBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtendedKey.PrivkeyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtendedKey.PrivkeyBytes() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestExtendedKey_ToPrivkey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		key       []byte
		pubKey    []byte
		chainCode []byte
		childNum  uint32
		isPrivate bool
	}

	// prepare test-cases
	tests := []struct {
		name    string
		key     *ExtendedKey
		want    *ecdsa.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			key:     tDefaultExtendedKey,
			want:    tDefaultKey,
			wantErr: false,
		},
		{
			key:     tDefaultExtendedPubKey,
			want:    nil,
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.key
			got, err := k.ToPrivkey()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtendedKey.ToPrivkey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtendedKey.ToPrivkey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestExtendedKey_ToPubkey(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		key       []byte
		pubKey    []byte
		chainCode []byte
		childNum  uint32
		isPrivate bool
	}

	// prepare test-cases
	tests := []struct {
		name    string
		key     *ExtendedKey
		want    *ecdsa.PublicKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			key:     tDefaultExtendedPubKey,
			want:    tDefaultPubKey,
			wantErr: false,
		},
		{
			key:     tDefaultExtendedKey,
			want:    tDefaultPubKey,
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.key

			got, err := k.ToPubkey()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtendedKey.ToPubkey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtendedKey.ToPubkey() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
