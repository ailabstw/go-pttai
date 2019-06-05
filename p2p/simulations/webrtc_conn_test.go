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

package simulations

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
	signalserver "github.com/ailabstw/pttai-signal-server"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type status struct {
	Version   uint32
	NetworkID uint32
}

type Msg struct {
	Bytes []byte
}

func readStatus(peer *p2p.Peer, networkID uint32, rw p2p.MsgReadWriter) error {
	msg, err := rw.ReadMsg()
	if err != nil {
		return err
	}

	if msg.Code != uint64(CodeTypeStatus) {
		return ErrInvalidMsgCode
	}

	if msg.Size > ProtocolMaxMsgSize {
		return ErrMsgTooLarge
	}

	status := &status{}
	err = msg.Decode(&status)
	if err != nil {
		return err
	}

	if status.NetworkID != networkID {
		return ErrInvalidData
	}

	if uint(status.Version) != Version {
		return ErrInvalidData
	}

	return nil
}

func handshake(peer *p2p.Peer, networkID uint32, rw p2p.MsgReadWriter) error {
	errc := make(chan error, 2)

	go func() {
		errc <- p2p.Send(rw, uint64(CodeTypeStatus), &status{
			Version:   uint32(Version),
			NetworkID: networkID,
		})
	}()

	go func() {
		errc <- readStatus(peer, networkID, rw)
	}()

	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}

	return nil
}

func handlePeer(peer *p2p.Peer, rw p2p.MsgReadWriter, theBytes []byte, msgChan chan p2p.Msg) error {
	// 1. basic handshake
	err := handshake(peer, NetworkID, rw)
	if err != nil {
		return err
	}

	key2ID := discover.PubkeyID(&tKey2.PublicKey)

	peerID := peer.ID()

	isKey1 := false

	log.Debug("handlePeer: after handshake", "key2ID", key2ID, "peerID", peerID, "isKey1", isKey1)

	if peerID == key2ID {
		// I'm with key1
		isKey1 = true

		sendMsg := &Msg{
			Bytes: theBytes,
		}

		_ = p2p.Send(rw, uint64(CodeTypeTest), sendMsg)
	} else {
		// I'm with key2
		msg, _ := rw.ReadMsg()
		msgChan <- msg
	}

looping:
	for {
		_, err = rw.ReadMsg()
		if err != nil {
			log.Error("uable to ReadMsg (for-loop)", "e", err, "peerID", peerID, "isKey1", isKey1)
			break looping
		}
	}

	return err
}

func testProtoRun(peer *p2p.Peer, rw p2p.MsgReadWriter, theBytes []byte, msgChan chan p2p.Msg) error {

	log.Debug("testProtoRun: get new peer", "peer", peer)

	err := handlePeer(peer, rw, theBytes, msgChan)
	log.Debug("testProtoRun: after HandlePeer", "peer", peer, "e", err)

	return nil
}

func TestWebrtcConn_Write(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	addr := "127.0.0.1:9489"
	go func() {
		server := signalserver.NewServer()

		srv := &http.Server{Addr: addr}
		r := mux.NewRouter()
		r.HandleFunc("/signal", server.SignalHandler)
		srv.Handler = r

		srv.ListenAndServe()
	}()

	theBytes, _ := ioutil.ReadFile(LargeFileFilename)
	msgChan := make(chan p2p.Msg)

	testProto := p2p.Protocol{
		Name:    "testProto",
		Version: 1,
		Length:  5,
		Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
			return testProtoRun(peer, rw, theBytes, msgChan)
		},
		NodeInfo: func() interface{} { return nil },
		PeerInfo: func(id discover.NodeID) interface{} { return nil },
	}

	var srv1 *p2p.Server
	var srv2 *p2p.Server

	// srv1
	go func() {
		config1 := p2p.Config{
			PrivateKey:      tKey1,
			MaxPeers:        10,
			SignalServerURL: url.URL{Scheme: "ws", Host: "127.0.0.1:9489", Path: "/signal"},
		}

		srv1 = &p2p.Server{Config: config1}
		srv1.Protocols = []p2p.Protocol{testProto}

		_ = srv1.Start()

		nodeID := discover.PubkeyID(&tKey2.PublicKey)
		node := discover.NewWebrtcNode(nodeID)
		srv1.AddPeer(node)
	}()

	go func() {
		config2 := p2p.Config{
			PrivateKey:      tKey2,
			MaxPeers:        10,
			SignalServerURL: url.URL{Scheme: "ws", Host: "127.0.0.1:9489", Path: "/signal"},
		}

		srv2 = &p2p.Server{Config: config2}
		srv2.Protocols = []p2p.Protocol{testProto}
		_ = srv2.Start()
	}()

	msg, ok := <-msgChan

	srv1.Stop()
	srv2.Stop()

	assert.Equal(t, true, ok)

	assert.Equal(t, uint64(CodeTypeTest), msg.Code)

	data := &Msg{}

	err := msg.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, len(theBytes), len(data.Bytes))
}
