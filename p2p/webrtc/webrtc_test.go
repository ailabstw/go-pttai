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
	"net/http"
	"net/url"
	"testing"

	signalserver "github.com/ailabstw/pttai-signal-server"

	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func handleWebrtcWithTest(t *testing.T) func(conn *WebrtcConn) {
	return func(conn *WebrtcConn) {
		t.Logf("handleWebrtc: start: conn: %v", conn)
	}
}

func TestWebrtc(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	addr := "127.0.0.1:9488"
	go func() {
		server := signalserver.NewServer()

		srv := &http.Server{Addr: addr}
		r := mux.NewRouter()
		r.HandleFunc("/signal", server.SignalHandler)
		srv.Handler = r

		srv.ListenAndServe()
	}()

	url := url.URL{Scheme: "ws", Host: addr, Path: "/signal"}

	handle := handleWebrtcWithTest(t)

	key1, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("failed generate key1: %v", err)
	}
	nodeID1 := discover.PubkeyID(&key1.PublicKey)

	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("failed generate key2: %v", err)
	}
	nodeID2 := discover.PubkeyID(&key2.PublicKey)

	c1, err := NewWebrtc(nodeID1, key1, url, handle)
	t.Logf("TestClientSendReceive: after c1: e: %v", err)
	assert.NoError(t, err)

	_, err = NewWebrtc(nodeID2, key2, url, handle)
	t.Logf("TestClientSendReceive: after c2: e: %v", err)
	assert.NoError(t, err)

	_, err = c1.CreateOffer(nodeID2)
	assert.NoError(t, err)
}
