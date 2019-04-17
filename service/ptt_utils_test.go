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
	"bytes"
	"crypto/aes"
	"crypto/ecdsa"
	"reflect"
	"sync"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/common"
)

func TestPtt_EncryptData(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		config         *Config
		myID           *types.PttID
		myIDHash       uint64
		myKey          *ecdsa.PrivateKey
		myNodeID       *discover.NodeID
		eventMux       *event.TypeMux
		peerLock       sync.RWMutex
		myPeers        map[discover.NodeID]*PttPeer
		importantPeers map[discover.NodeID]*PttPeer
		memberPeers    map[discover.NodeID]*PttPeer
		randomPeers    map[discover.NodeID]*PttPeer
		newPeerCh      chan *PttPeer
		noMorePeers    chan struct{}
		peerHash       map[uint64]bool
		peerWG         sync.WaitGroup
		services       map[string]Service
		entityLock     sync.RWMutex
		hashLock       sync.RWMutex
		entities       map[types.PttID]Entity
		joins          map[common.Address]*types.PttID
		ops            map[common.Address]*types.PttID
		myEntity       Entity
		quitSync       chan struct{}
		syncWG         sync.WaitGroup
		server         *p2p.Server
		protocols      []p2p.Protocol
		apis           []rpc.API
		networkID      uint32
	}
	type args struct {
		op   OpType
		data []byte
		key  *KeyInfo
	}

	// prepare test-cases
	tests := []struct {
		name    string
		p       *BasePtt
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{op: tDefaultOp, data: tDefaultDataBytes, key: tDefaultKeyInfo},
			want: tDefaultEncData,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			got, err := p.EncryptData(tt.args.op, tt.args.data, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ptt.EncryptData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ptt.EncryptData() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPtt_DecryptData(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		config         *Config
		myID           *types.PttID
		myIDHash       uint64
		myKey          *ecdsa.PrivateKey
		myNodeID       *discover.NodeID
		eventMux       *event.TypeMux
		peerLock       sync.RWMutex
		myPeers        map[discover.NodeID]*PttPeer
		importantPeers map[discover.NodeID]*PttPeer
		memberPeers    map[discover.NodeID]*PttPeer
		randomPeers    map[discover.NodeID]*PttPeer
		newPeerCh      chan *PttPeer
		noMorePeers    chan struct{}
		peerHash       map[uint64]bool
		peerWG         sync.WaitGroup
		services       map[string]Service
		entityLock     sync.RWMutex
		hashLock       sync.RWMutex
		entities       map[types.PttID]Entity
		joins          map[common.Address]*types.PttID
		ops            map[common.Address]*types.PttID
		myEntity       Entity
		quitSync       chan struct{}
		syncWG         sync.WaitGroup
		server         *p2p.Server
		protocols      []p2p.Protocol
		apis           []rpc.API
		networkID      uint32
	}
	type args struct {
		encMsg []byte
		key    *KeyInfo
	}

	// prepare test-cases
	tests := []struct {
		name    string
		p       *BasePtt
		args    args
		want    OpType
		want1   []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				encMsg: tDefaultEncData,
				key:    tDefaultKeyInfo,
			},
			want:  tDefaultOp,
			want1: tDefaultDataBytes,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			got, got1, err := p.DecryptData(tt.args.encMsg, tt.args.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("Ptt.DecryptData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ptt.DecryptData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Ptt.DecryptData() got1 = %v, want1 %v", got1, tt.want1)
			}
		})
	}

	// teardown test
}

func TestPtt_MarshalData(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		config         *Config
		myID           *types.PttID
		myIDHash       uint64
		myKey          *ecdsa.PrivateKey
		myNodeID       *discover.NodeID
		eventMux       *event.TypeMux
		peerLock       sync.RWMutex
		myPeers        map[discover.NodeID]*PttPeer
		importantPeers map[discover.NodeID]*PttPeer
		memberPeers    map[discover.NodeID]*PttPeer
		randomPeers    map[discover.NodeID]*PttPeer
		newPeerCh      chan *PttPeer
		noMorePeers    chan struct{}
		peerHash       map[uint64]bool
		peerWG         sync.WaitGroup
		services       map[string]Service
		entityLock     sync.RWMutex
		hashLock       sync.RWMutex
		entities       map[types.PttID]Entity
		joins          map[common.Address]*types.PttID
		ops            map[common.Address]*types.PttID
		myEntity       Entity
		quitSync       chan struct{}
		syncWG         sync.WaitGroup
		server         *p2p.Server
		protocols      []p2p.Protocol
		apis           []rpc.API
		networkID      uint32
	}
	type args struct {
		code    CodeType
		hash    *common.Address
		encData []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		p       *BasePtt
		args    args
		want    *PttData
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			p:    tDefaultPtt,
			args: args{code: CodeTypeOp, hash: &tDefaultHash, encData: tDefaultEncData},
			want: tDefaultPttData,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			got, err := p.MarshalData(tt.args.code, tt.args.hash, tt.args.encData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ptt.MarshalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ptt.MarshalData() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestPtt_UnmarshalData(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type fields struct {
		config         *Config
		myID           *types.PttID
		myIDHash       uint64
		myKey          *ecdsa.PrivateKey
		myNodeID       *discover.NodeID
		eventMux       *event.TypeMux
		peerLock       sync.RWMutex
		myPeers        map[discover.NodeID]*PttPeer
		importantPeers map[discover.NodeID]*PttPeer
		memberPeers    map[discover.NodeID]*PttPeer
		randomPeers    map[discover.NodeID]*PttPeer
		newPeerCh      chan *PttPeer
		noMorePeers    chan struct{}
		peerHash       map[uint64]bool
		peerWG         sync.WaitGroup
		services       map[string]Service
		entityLock     sync.RWMutex
		hashLock       sync.RWMutex
		entities       map[types.PttID]Entity
		joins          map[common.Address]*types.PttID
		ops            map[common.Address]*types.PttID
		myEntity       Entity
		quitSync       chan struct{}
		syncWG         sync.WaitGroup
		server         *p2p.Server
		protocols      []p2p.Protocol
		apis           []rpc.API
		networkID      uint32
	}
	type args struct {
		pttData *PttData
	}

	// prepare test-cases
	tests := []struct {
		name    string
		p       *BasePtt
		args    args
		want    CodeType
		want1   *common.Address
		want2   []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			p:     tDefaultPtt,
			args:  args{pttData: tDefaultPttData},
			want:  CodeTypeOp,
			want1: &tDefaultHash,
			want2: tDefaultEncData,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			got, got1, got2, err := p.UnmarshalData(tt.args.pttData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ptt.UnmarshalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ptt.UnmarshalData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Ptt.UnmarshalData() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("Ptt.UnmarshalData() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}

	// teardown test
}

func Test_addAndRemoveBase64Padding(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test empty string",
			args: args{
				value: "",
			},
			want: "",
		},
		{
			name: "test string length is 3",
			args: args{
				value: "123",
			},
			want: "123=",
		},
		{
			name: "test string length is 4",
			args: args{
				value: "1234",
			},
			want: "1234",
		},
		{
			name: "test string length is 5",
			args: args{
				value: "12345",
			},
			want: "12345===",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addBase64Padding(tt.args.value); got != tt.want {
				t.Errorf("addBase64Padding() = %v, want %v", got, tt.want)
			}
			if ori := removeBase64Padding(tt.want); ori != tt.args.value {
				t.Errorf("removeBase64Padding() = %v, want %v", ori, tt.args.value)
			}
		})
	}
}

func Test_aesPadUnPad(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test empty bytes",
			args: args{
				src: []byte(""),
			},
			want:    aes.BlockSize,
			wantErr: false,
		},
		{
			name: "test bytes length is 3",
			args: args{
				src: []byte("123"),
			},
			want:    aes.BlockSize,
			wantErr: false,
		},
		{
			name: "test bytes length is 22",
			args: args{
				src: []byte("1234567890123456789012"),
			},
			want:    aes.BlockSize,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := aesPad(tt.args.src)
			if len(got)%tt.want != 0 {
				t.Errorf("aesPad() length = %v, want %v", len(got), tt.want)
			}
			ori, err := aesUnpad(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ptt.UnmarshalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(ori, tt.args.src) {
				t.Errorf("aesUnpad() = %v, want %v", ori, tt.args.src)
			}
		})
	}
}
