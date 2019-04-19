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
	"net"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discv5"

	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/pion/datachannel"
	"github.com/pion/webrtc"
)

type webrtcAddr struct {
	addr string
}

func (addr *webrtcAddr) Network() string {
	return "webrtc"
}

func (addr *webrtcAddr) String() string {
	return addr.addr
}

type webrtcInfo struct {
	NodeID discover.NodeID

	isClosed int32

	PeerConn *webrtc.PeerConnection
	DataConn datachannel.ReadWriteCloser
}

func (info *webrtcInfo) Close() {
	isSwapped := atomic.CompareAndSwapInt32(&info.isClosed, 0, 1)
	if !isSwapped {
		return
	}

	info.DataConn.Close()
	info.PeerConn.Close()
}

type WebrtcConn struct {
	info       *webrtcInfo
	localAddr  *webrtcAddr
	remoteAddr *webrtcAddr
}

func NewWebrtcConn(nodeID discv5.NodeID, fromID discv5.NodeID, info *webrtcInfo) (*WebrtcConn, error) {

	localAddr := parseWebrtcAddr(nodeID)

	remoteAddr := parseWebrtcAddr(fromID)

	conn := &WebrtcConn{
		info:       info,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	return conn, nil
}

func (w *WebrtcConn) Read(b []byte) (int, error) {
	return w.info.DataConn.Read(b)
}

func (w *WebrtcConn) Write(b []byte) (int, error) {
	return w.info.DataConn.Write(b)
}

func (w *WebrtcConn) Close() error {
	w.info.Close()
	return nil
}

func (w *WebrtcConn) LocalAddr() net.Addr {
	return w.localAddr
}

func (w *WebrtcConn) RemoteAddr() net.Addr {
	return w.remoteAddr
}

/*
SetDeadline: skip implementation
*/
func (w *WebrtcConn) SetDeadline(t time.Time) error {
	return nil
}

/*
SetReadDeadline: skip implementation
*/
func (w *WebrtcConn) SetReadDeadline(t time.Time) error {
	return nil
}

/*
SetWriteDeadline: skip implementation
*/
func (w *WebrtcConn) SetWriteDeadline(t time.Time) error {
	return nil
}
