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

package p2p

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	signalserver "github.com/ailabstw/pttai-signal-server"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/pion/datachannel"
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
	OfferChan chan *WebrtcInfo
}

type WebrtcInfo struct {
	NodeID discover.NodeID

	isClosed int32

	PeerConn *webrtc.PeerConnection
	DataConn datachannel.ReadWriteCloser
}

func (info *WebrtcInfo) Close() {
	isSwapped := atomic.CompareAndSwapInt32(&info.isClosed, 0, 1)
	if !isSwapped {
		return
	}

	info.DataConn.Close()
	info.PeerConn.Close()
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

	handleAnswerChannel func(info *WebrtcInfo)
}

func NewWebrtc(
	nodeID discover.NodeID,
	privKey *ecdsa.PrivateKey,
	url url.URL,
	h func(info *WebrtcInfo),
) (*Webrtc, error) {

	// XXX we may need the unified nodeID type.
	var tmpNodeID discv5.NodeID
	copy(tmpNodeID[:], nodeID[:])

	client, err := signalserver.NewClient(tmpNodeID, privKey, url)
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

		handleAnswerChannel: h,
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

	close(w.quitChan)

	w.client.Close()

	w.offerConnMapLock.Lock()
	defer w.offerConnMapLock.Unlock()

	for _, info := range w.offerConnMap {
		info.PeerConn.Close()
	}

	w.offerConnMap = make(map[discv5.NodeID]*offerConnInfo)

	return nil
}

/*
CreateOffer actively create a new offer for nodeID.
1. create peerConn
2. create data-channel.
3. provide offer
*/

func (w *Webrtc) CreateOffer(nodeID discover.NodeID) (*WebrtcInfo, error) {
	if w.isClosed != 0 {
		return nil, ErrInvalidWebrtc
	}

	// XXX we may need the unified nodeID type.
	var tmpNodeID discv5.NodeID
	copy(tmpNodeID[:], nodeID[:])

	offerChan := make(chan *WebrtcInfo)

	peerConn, err := w.createOffer(tmpNodeID, offerChan)
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(context.Background(), TimeoutSecondConnectWebrtc*time.Second)
	defer cancel()

	var info *WebrtcInfo
	select {
	case tmp, ok := <-offerChan:
		if ok {
			info = tmp
		}
	case <-w.quitChan:
	case <-tctx.Done():
	}

	if info == nil {
		w.removeFromOfferConnMap(tmpNodeID, peerConn)
		peerConn.Close()
		return nil, ErrInvalidWebrtcOffer
	}

	return info, nil
}

func (w *Webrtc) createOffer(nodeID discv5.NodeID, offerChan chan *WebrtcInfo) (peerConnection *webrtc.PeerConnection, err error) {

	peerConnection, err = w.api.NewPeerConnection(w.config)
	if err != nil {
		return nil, err
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Info("CreateConn: ICE Connection State has changed", "nodeID", nodeID, "state", connectionState)
	})

	var offer webrtc.SessionDescription
	offer, err = peerConnection.CreateOffer(nil)
	if err != nil {
		peerConnection.Close()
		return nil, err
	}

	var offerID string
	offerID, err = w.parseOfferID(offer)
	if err != nil {
		peerConnection.Close()
		return nil, err
	}
	defer func() {
		if err != nil {
			peerConnection.Close()
		}
	}()

	info := &offerConnInfo{
		PeerConn:  peerConnection,
		offerID:   offerID,
		OfferChan: offerChan,
	}

	err = w.addToOfferConnMap(nodeID, info)
	if err != nil {
		peerConnection.Close()
		return nil, err
	}
	defer func() {
		if err != nil {
			w.removeFromOfferConnMap(nodeID, peerConnection)
		}
	}()

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return nil, err
	}

	//
	marshaled, err := json.Marshal(offer)
	if err != nil {
		return nil, err
	}

	sig := &writeSignal{
		ToID: nodeID,
		Msg:  marshaled,
	}

	err = w.tryPassWriteChan(sig)
	if err != nil {
		return nil, err
	}

	return peerConnection, nil
}

/*
receiveOffer receives offer from fromID and we need to create the corresponding answer.
*/
func (w *Webrtc) receiveOffer(fromID discv5.NodeID, offer webrtc.SessionDescription) error {

	offerID, err := w.parseOfferID(offer)
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
			info, err := dataChannelToWebrtcInfo(d, fromID, peerConn)
			if err != nil {
				return

			}

			w.handleAnswerChannel(info)
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
		info, err := dataChannelToWebrtcInfo(dataChannel, fromID, peerConn)
		if err != nil {
			return
		}

		select {
		case offerInfo.OfferChan <- info:
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

func dataChannelToWebrtcInfo(
	dataChannel *webrtc.DataChannel,
	fromID discv5.NodeID,
	peerConn *webrtc.PeerConnection,
) (*WebrtcInfo, error) {

	dataConn, err := dataChannel.Detach()
	if err != nil {
		return nil, err
	}

	// XXX we may need the unified nodeID type.
	var nodeID discover.NodeID
	copy(nodeID[:], fromID[:])

	info := &WebrtcInfo{
		NodeID: nodeID,

		PeerConn: peerConn,
		DataConn: dataConn,
	}

	return info, nil
}

func (w *Webrtc) parseOfferID(offer webrtc.SessionDescription) (string, error) {
	return "", types.ErrNotImplemented
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
