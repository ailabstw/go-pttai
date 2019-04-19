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

package webrtc

import (
	"reflect"
	"sync"
	"testing"

	"github.com/ailabstw/go-pttai/p2p/discover"
	signalserver "github.com/ailabstw/pttai-signal-server"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/pion/webrtc"
)

func TestWebrtc_CreateOffer(t *testing.T) {
	type fields struct {
		isClosed            int32
		client              *signalserver.Client
		writeChan           chan *writeSignal
		quitChan            chan struct{}
		config              webrtc.Configuration
		api                 *webrtc.API
		offerConnMapLock    sync.RWMutex
		offerConnMap        map[discv5.NodeID]*offerConnInfo
		handleAnswerChannel func(conn *WebrtcConn)
	}
	type args struct {
		nodeID discover.NodeID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *WebrtcConn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Webrtc{
				isClosed:            tt.fields.isClosed,
				client:              tt.fields.client,
				writeChan:           tt.fields.writeChan,
				quitChan:            tt.fields.quitChan,
				config:              tt.fields.config,
				api:                 tt.fields.api,
				offerConnMapLock:    tt.fields.offerConnMapLock,
				offerConnMap:        tt.fields.offerConnMap,
				handleAnswerChannel: tt.fields.handleAnswerChannel,
			}
			got, err := w.CreateOffer(tt.args.nodeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Webrtc.CreateOffer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Webrtc.CreateOffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebrtc_createOffer(t *testing.T) {
	type fields struct {
		isClosed            int32
		client              *signalserver.Client
		writeChan           chan *writeSignal
		quitChan            chan struct{}
		config              webrtc.Configuration
		api                 *webrtc.API
		offerConnMapLock    sync.RWMutex
		offerConnMap        map[discv5.NodeID]*offerConnInfo
		handleAnswerChannel func(conn *WebrtcConn)
	}
	type args struct {
		nodeID    discv5.NodeID
		offerChan chan *WebrtcConn
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		wantPeerConnection *webrtc.PeerConnection
		wantErr            bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Webrtc{
				isClosed:            tt.fields.isClosed,
				client:              tt.fields.client,
				writeChan:           tt.fields.writeChan,
				quitChan:            tt.fields.quitChan,
				config:              tt.fields.config,
				api:                 tt.fields.api,
				offerConnMapLock:    tt.fields.offerConnMapLock,
				offerConnMap:        tt.fields.offerConnMap,
				handleAnswerChannel: tt.fields.handleAnswerChannel,
			}
			gotPeerConnection, err := w.createOffer(tt.args.nodeID, tt.args.offerChan)
			if (err != nil) != tt.wantErr {
				t.Errorf("Webrtc.createOffer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPeerConnection, tt.wantPeerConnection) {
				t.Errorf("Webrtc.createOffer() = %v, want %v", gotPeerConnection, tt.wantPeerConnection)
			}
		})
	}
}
