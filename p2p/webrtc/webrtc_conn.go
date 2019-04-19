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
	"time"

	"github.com/ailabstw/go-pttai/common/types"
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

type WebrtcConn struct {
	info       *webrtcInfo
	localAddr  *webrtcAddr
	remoteAddr *webrtcAddr
}

func NewWebrtcConn(info *webrtcInfo) (*WebrtcConn, error) {

	localAddr, err := parseWebrtcAddr(info.PeerConn.CurrentLocalDescription())
	if err != nil {
		return nil, err
	}

	remoteAddr, err := parseWebrtcAddr(info.PeerConn.CurrentRemoteDescription())
	if err != nil {
		return nil, err
	}

	conn := &WebrtcConn{
		info:       info,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	return conn, nil
}

func (w *WebrtcConn) Read(b []byte) (int, error) {
	return 0, types.ErrNotImplemented
}

func (w *WebrtcConn) Write(b []byte) (int, error) {
	return 0, types.ErrNotImplemented
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
