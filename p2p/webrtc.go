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
	"crypto/ecdsa"
	"encoding/json"
	"net/url"
	"sync"
	"sync/atomic"

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

type writeInfo struct {
	PeerConn *webrtc.PeerConnection
	offerID  string
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
	client *signalserver.Client

	writeChan chan *writeSignal
	quitChan  chan struct{}

	config webrtc.Configuration
	api    *webrtc.API

	writeMapLock sync.RWMutex
	writeMap     map[discv5.NodeID]*writeInfo

	handleChannel func(info *WebrtcInfo)
}

func NewWebrtc(nodeID discover.NodeID, privKey *ecdsa.PrivateKey, url url.URL) (*Webrtc, error) {

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
		client:   client,
		config:   config,
		api:      api,
		quitChan: make(chan struct{}),
	}

	go w.ReadLoop()

	go w.WriteLoop()

	return w, nil
}

func (w *Webrtc) SetChannelHandler(h func(info *WebrtcInfo)) {
	w.handleChannel = h
}

func (w *Webrtc) Close() error {
	close(w.quitChan)
	w.client.Close()

	return nil
}

/*
CreateOffer actively create a new offer for nodeID.
1. create peerConn
2. create data-channel.
3. provide offer
*/
func (w *Webrtc) CreateOffer(nodeID discover.NodeID) error {

	// XXX we may need the unified nodeID type.
	var tmpNodeID discv5.NodeID
	copy(tmpNodeID[:], nodeID[:])

	peerConnection, err := w.api.NewPeerConnection(w.config)
	if err != nil {
		return err
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Info("CreateConn: ICE Connection State has changed", "nodeID", nodeID, "state", connectionState)
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		peerConnection.Close()
		return err
	}

	offerID, err := w.parseOfferID(offer)
	if err != nil {
		peerConnection.Close()
		return err
	}

	err = w.addWriteMap(tmpNodeID, peerConnection, offerID)
	if err != nil {
		peerConnection.Close()
		return err
	}

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		w.removeFromWriteMap(tmpNodeID, peerConnection, offerID)
		peerConnection.Close()
		return err
	}

	//
	marshaled, err := json.Marshal(offer)
	if err != nil {
		w.removeFromWriteMap(tmpNodeID, peerConnection, offerID)
		peerConnection.Close()
		return err
	}

	sig := &writeSignal{
		ToID: tmpNodeID,
		Msg:  marshaled,
	}

	err = w.tryPassWriteChan(sig)
	if err != nil {
		w.removeFromWriteMap(tmpNodeID, peerConnection, offerID)
		peerConnection.Close()
		return err
	}

	return nil
}

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
			w.onConnect(d, fromID, peerConn)
		})
	})

	err = peerConn.SetRemoteDescription(offer)
	if err != nil {
		return err
	}

	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// try pass write chan
	marshalled, err := json.Marshal(answer)
	if err != nil {
		return err
	}

	sig := &writeSignal{
		ToID:  fromID,
		Msg:   marshalled,
		Extra: []byte(offerID),
	}

	err = w.tryPassWriteChan(sig)
	if err != nil {
		return err
	}

	return nil
}

func (w *Webrtc) receiveAnswer(fromID discv5.NodeID, answer webrtc.SessionDescription, offerID string) error {

	writeInfo, err := w.tryGetWriteInfo(fromID, offerID)
	if err != nil {
		return err
	}

	peerConn := writeInfo.PeerConn

	dataChannel, err := peerConn.CreateDataChannel("data", nil)
	if err != nil {
		peerConn.Close()
		return err
	}

	dataChannel.OnOpen(func() {
		w.onConnect(dataChannel, fromID, peerConn)
	})

	err = peerConn.SetRemoteDescription(answer)
	if err != nil {
		peerConn.Close()
		return err
	}

	return nil
}

func (w *Webrtc) onConnect(dataChannel *webrtc.DataChannel, fromID discv5.NodeID, peerConn *webrtc.PeerConnection) {

	dataConn, err := dataChannel.Detach()
	if err != nil {
		return
	}

	// XXX we may need the unified nodeID type.
	var nodeID discover.NodeID
	copy(nodeID[:], fromID[:])

	info := &WebrtcInfo{
		NodeID: nodeID,

		PeerConn: peerConn,
		DataConn: dataConn,
	}

	w.handleChannel(info)
}

func (w *Webrtc) parseOfferID(offer webrtc.SessionDescription) (string, error) {
	return "", types.ErrNotImplemented
}

func (w *Webrtc) addWriteMap(nodeID discv5.NodeID, peerConn *webrtc.PeerConnection, offerID string) error {
	return types.ErrNotImplemented
}

func (w *Webrtc) removeFromWriteMap(nodeID discv5.NodeID, peerConn *webrtc.PeerConnection, offerID string) error {
	return types.ErrNotImplemented
}

func (w *Webrtc) tryGetWriteInfo(fromID discv5.NodeID, offerID string) (*writeInfo, error) {
	return nil, types.ErrNotImplemented
}

func (w *Webrtc) tryPassWriteChan(sig *writeSignal) error {
	w.writeChan <- sig

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
			return err
		}
	}
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
