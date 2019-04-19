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

type WebrtcAddr struct {
	net  string
	addr string
}

func (addr *WebrtcAddr) Network() string {
	return addr.net
}

func (addr *WebrtcAddr) String() string {
	return addr.addr
}

type WebrtcConn struct {
	info *WebrtcInfo
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
	return &WebrtcAddr{}
}

func (w *WebrtcConn) RemoteAddr() net.Addr {
	return &WebrtcAddr{}
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
