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
	"time"

	signalserver "github.com/ailabstw/pttai-signal-server"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func handleWebrtcWithTest(t *testing.T) func(conn *WebrtcConn) {
	return func(conn *WebrtcConn) {
		t.Logf("handleWebrtcWithTest: start: conn: %v", conn)
		b := make([]byte, 10)
		n, err := conn.Read(b)
		t.Logf("handleWebrtcWithTest: after Read: n: %v e: %v b: %v", n, err, b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, []byte("test"), b[:n])

		n, err = conn.Write([]byte("test2"))
		t.Logf("handleWebrtcWithTest: after Write: n: %v e: %v", n, err)
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
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

	w1, err := NewWebrtc(nodeID1, key1, url, handle)
	t.Logf("TestClientSendReceive: after c1: e: %v", err)
	assert.NoError(t, err)

	go func() {
		_, err = NewWebrtc(nodeID2, key2, url, handle)
		t.Logf("TestClientSendReceive: after c2: e: %v", err)
		assert.NoError(t, err)

		select {}
	}()

	conn1, err := w1.CreateOffer(nodeID2)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	n, err := conn1.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	b := make([]byte, 10)
	n, err = conn1.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("test2"), b[:n])

	time.Sleep(1 * time.Second)

	log.Debug("after Sleep")
}
