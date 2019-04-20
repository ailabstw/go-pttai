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
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	signalserver "github.com/ailabstw/pttai-signal-server"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/pion/webrtc"
)

type writeSignal struct {
	ToID  discv5.NodeID
	Msg   []byte
	Extra []byte
}

type offerConnInfo struct {
	PeerConn  *webrtc.PeerConnection
	offerID   string
	OfferChan chan *WebrtcConn
}

type Webrtc struct {
	isClosed int32

	client *signalserver.Client

	writeChan chan *writeSignal
	quitChan  chan struct{}

	config webrtc.Configuration
	api    *webrtc.API

	offerConnMapLock sync.RWMutex
	offerConnMap     map[discv5.NodeID]*offerConnInfo

	handleChannel func(conn *WebrtcConn)

	nodeID discv5.NodeID
}

func NewWebrtc(
	nodeID discover.NodeID,
	privKey *ecdsa.PrivateKey,
	url url.URL,
	h func(conn *WebrtcConn),
) (*Webrtc, error) {

	// XXX we may need the unified nodeID type.
	var tmpNodeID discv5.NodeID
	copy(tmpNodeID[:], nodeID[:])

	log.Debug("NewWebrtc: to NewClient", "nodeID", nodeID, "url", url)
	client, err := signalserver.NewClient(tmpNodeID, privKey, url)
	log.Debug("NewWebrtc: after NewClient", "e", err, "nodeID", nodeID)
	if err != nil {
		return nil, err
	}

	s := webrtc.SettingEngine{}
	s.DetachDataChannels()

	api := webrtc.NewAPI(webrtc.WithSettingEngine(s))

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	w := &Webrtc{
		client: client,

		writeChan: make(chan *writeSignal),
		quitChan:  make(chan struct{}),

		config: config,
		api:    api,

		offerConnMap: make(map[discv5.NodeID]*offerConnInfo),

		handleChannel: h,

		nodeID: tmpNodeID,
	}

	go w.ReadLoop()

	go w.WriteLoop()

	return w, nil
}

func (w *Webrtc) Close() error {
	isSwapped := atomic.CompareAndSwapInt32(&w.isClosed, 0, 1)
	if !isSwapped {
		return nil
	}

	log.Debug("Webrtc.Close: start")

	close(w.quitChan)

	w.client.Close()

	w.offerConnMapLock.Lock()
	defer w.offerConnMapLock.Unlock()

	for _, info := range w.offerConnMap {
		info.PeerConn.Close()
	}

	w.offerConnMap = make(map[discv5.NodeID]*offerConnInfo)

	log.Debug("Webrtc.Close: done")

	return nil
}

/*
CreateOffer actively create a new offer for nodeID.
1. create peerConn
2. create data-channel.
3. provide offer
*/

func (w *Webrtc) CreateOffer(nodeID discover.NodeID) (*WebrtcConn, error) {
	if w.isClosed != 0 {
		return nil, ErrInvalidWebrtc
	}

	log.Debug("CreateOffer: start", "me", w.nodeID, "nodeID", nodeID)
	// XXX we may need the unified nodeID type.
	var tmpNodeID discv5.NodeID
	copy(tmpNodeID[:], nodeID[:])

	offerChan := make(chan *WebrtcConn)

	peerConn, err := w.createOffer(tmpNodeID, offerChan)
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(context.Background(), TimeoutSecondConnectWebrtc*time.Second)
	defer cancel()

	var conn *WebrtcConn
	select {
	case tmp, ok := <-offerChan:
		if ok {
			conn = tmp
		}
	case <-w.quitChan:
	case <-tctx.Done():
	}

	if conn == nil {
		w.removeFromOfferConnMap(tmpNodeID, peerConn)
		peerConn.Close()
		return nil, ErrInvalidWebrtcOffer
	}

	log.Debug("CreateOffer: done", "conn", conn)

	return conn, nil
}

func (w *Webrtc) createOffer(nodeID discv5.NodeID, offerChan chan *WebrtcConn) (peerConn *webrtc.PeerConnection, err error) {

	peerConn, err = w.api.NewPeerConnection(w.config)
	if err != nil {
		return nil, err
	}

	peerConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Info("createOffer: ICE Connection State has changed", "nodeID", nodeID, "state", connectionState)
	})

	var offer webrtc.SessionDescription
	offer, err = peerConn.CreateOffer(nil)
	if err != nil {
		peerConn.Close()
		return nil, err
	}

	var offerID string
	offerID, err = parseOfferID(&offer)
	if err != nil {
		peerConn.Close()
		return nil, err
	}
	defer func() {
		if err != nil {
			peerConn.Close()
		}
	}()

	info := &offerConnInfo{
		PeerConn:  peerConn,
		offerID:   offerID,
		OfferChan: offerChan,
	}

	err = w.addToOfferConnMap(nodeID, info)
	if err != nil {
		peerConn.Close()
		return nil, err
	}
	defer func() {
		if err != nil {
			w.removeFromOfferConnMap(nodeID, peerConn)
		}
	}()

	err = peerConn.SetLocalDescription(offer)
	if err != nil {
		return nil, err
	}

	//
	marshalled, err := json.Marshal(offer)
	if err != nil {
		return nil, err
	}

	sig := &writeSignal{
		ToID: nodeID,
		Msg:  marshalled,
	}

	err = w.tryPassWriteChan(sig)
	if err != nil {
		return nil, err
	}

	return peerConn, nil
}

/*
receiveOffer receives offer from fromID and we need to create the corresponding answer.
*/
func (w *Webrtc) receiveOffer(fromID discv5.NodeID, offer webrtc.SessionDescription) error {

	offerID, err := parseOfferID(&offer)
	if err != nil {
		return err
	}

	peerConn, err := w.api.NewPeerConnection(w.config)
	if err != nil {
		return err
	}

	peerConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Info("receiveOffer: ICE Conn State changed", "state", connectionState)
	})

	peerConn.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			log.Info("receiveOffer: channel created", "fromID", fromID, "peerConn", peerConn)
			conn, err := dataChannelToWebrtcConn(d, w.nodeID, fromID, peerConn)
			if err != nil {
				return

			}

			w.handleChannel(conn)
		})
	})

	err = peerConn.SetRemoteDescription(offer)
	if err != nil {
		peerConn.Close()
		return err
	}

	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		peerConn.Close()
		return err
	}

	err = peerConn.SetLocalDescription(answer)
	if err != nil {
		peerConn.Close()
		return err
	}

	// try pass write chan
	marshalled, err := json.Marshal(answer)
	if err != nil {
		peerConn.Close()
		return err
	}

	sig := &writeSignal{
		ToID:  fromID,
		Msg:   marshalled,
		Extra: []byte(offerID),
	}

	err = w.tryPassWriteChan(sig)
	if err != nil {
		peerConn.Close()
		return err
	}

	return nil
}

func (w *Webrtc) receiveAnswer(fromID discv5.NodeID, answer webrtc.SessionDescription, offerID string) error {

	offerInfo, err := w.tryGetOfferInfo(fromID, offerID)
	if err != nil {
		return err
	}

	peerConn := offerInfo.PeerConn

	dataChannel, err := peerConn.CreateDataChannel("data", nil)
	if err != nil {
		peerConn.Close()
		return err
	}

	dataChannel.OnOpen(func() {
		log.Info("receiveAnswer: channel created", "fromID", fromID, "peerConn", peerConn)

		conn, err := dataChannelToWebrtcConn(dataChannel, w.nodeID, fromID, peerConn)
		if err != nil {
			return
		}

		select {
		case offerInfo.OfferChan <- conn:
		case <-w.quitChan:
		}
	})

	err = peerConn.SetRemoteDescription(answer)
	if err != nil {
		peerConn.Close()
		return err
	}

	return nil
}

func (w *Webrtc) addToOfferConnMap(nodeID discv5.NodeID, info *offerConnInfo) error {

	w.offerConnMapLock.Lock()
	defer w.offerConnMapLock.Unlock()

	if writeInfo, ok := w.offerConnMap[nodeID]; ok {
		writeInfo.PeerConn.Close()
	}

	w.offerConnMap[nodeID] = info

	return nil
}

func (w *Webrtc) removeFromOfferConnMap(nodeID discv5.NodeID, peerConn *webrtc.PeerConnection) error {
	w.offerConnMapLock.Lock()
	defer w.offerConnMapLock.Unlock()

	if writeInfo, ok := w.offerConnMap[nodeID]; ok && writeInfo.PeerConn == peerConn {
		delete(w.offerConnMap, nodeID)
	}

	return nil
}

func (w *Webrtc) tryGetOfferInfo(fromID discv5.NodeID, offerID string) (*offerConnInfo, error) {
	w.offerConnMapLock.Lock()
	defer w.offerConnMapLock.Unlock()

	if writeInfo, ok := w.offerConnMap[fromID]; ok {
		delete(w.offerConnMap, fromID)

		return writeInfo, nil
	}

	return nil, ErrInvalidWebrtcOffer
}

func (w *Webrtc) tryPassWriteChan(sig *writeSignal) error {
	select {
	case w.writeChan <- sig:
	case <-w.quitChan:
		return ErrInvalidWebrtc
	}

	return nil
}

func (w *Webrtc) ReadLoop() error {
	for {
		sig, err := w.client.Receive()
		if err != nil {
			return err
		}

		err = w.processSignal(sig)
		if err != nil {
			log.Warn("ReadLoop: unable to processSignal", "e", err, "sig", sig)
		}
	}

	w.Close()

	return nil
}

func (w *Webrtc) WriteLoop() error {
looping:
	for {
		select {
		case sig, ok := <-w.writeChan:
			if !ok {
				break looping
			}
			err := w.client.Send(sig.ToID, sig.Msg, sig.Extra)
			if err != nil {
				break looping
			}
		case <-w.quitChan:
			break looping
		}
	}

	w.Close()

	return nil
}

func (w *Webrtc) processSignal(signal *signalserver.Signal) error {
	session := webrtc.SessionDescription{}
	err := json.Unmarshal(signal.Msg, &session)
	if err != nil {
		return err
	}

	switch session.Type {
	case webrtc.SDPTypeOffer:
		err = w.receiveOffer(signal.FromID, session)
	case webrtc.SDPTypeAnswer:
		err = w.receiveAnswer(signal.FromID, session, string(signal.Extra))
	}

	return nil
}
