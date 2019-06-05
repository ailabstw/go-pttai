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

	"github.com/ailabstw/go-pttai/log"
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
	datachannel.ReadWriteCloser

	NodeID discover.NodeID

	isClosed int32

	PeerConn *webrtc.PeerConnection
}

func (info *webrtcInfo) Close() {
	isSwapped := atomic.CompareAndSwapInt32(&info.isClosed, 0, 1)
	if !isSwapped {
		return
	}

	info.ReadWriteCloser.Close()
	info.PeerConn.Close()
}

type WebrtcConn struct {
	info       *webrtcInfo
	localAddr  *webrtcAddr
	remoteAddr *webrtcAddr
}

func NewWebrtcConn(nodeID discv5.NodeID, fromID discv5.NodeID, info *webrtcInfo) (*WebrtcConn, error) {

	localAddr := parseWebrtcAddr(nodeID, info)

	remoteAddr := parseWebrtcAddr(fromID, info)

	conn := &WebrtcConn{
		info:       info,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	return conn, nil
}

func (w *WebrtcConn) Read(b []byte) (int, error) {
	readBytes := 0
	var err error

	buf := make([]byte, PACKET_SIZE+1)
	var eachN = 0
	var n = 0
	var firstByte = uint8(0)

looping:
	for pb := b; len(pb) > 0; pb = pb[readBytes:] {
		// 1. read to buf
		eachN, err = w.info.ReadWriteCloser.Read(buf)
		firstByte = uint8(255)
		if err == nil && eachN > 0 {
			firstByte = buf[0]
		}
		log.Debug("Read: (in-for-loop): after Read", "eachN", eachN, "firstByte", firstByte, "e", err)
		if err != nil {
			return 0, err
		}

		// 2. copy to pb
		readBytes = eachN - 1
		if readBytes > len(pb) {
			log.Error("Read: PacketTooLarge", "readBytes", readBytes, "pb", len(pb))
			return 0, ErrPacketTooLarge
		}

		// 3. copy
		copy(pb, buf[1:eachN])
		n += readBytes

		// 4. check PACKET_END
		if buf[0] == PACKET_END {
			break looping
		}
	}

	if buf[0] != PACKET_END {
		log.Error("Read: PacketTooLarge", "buf[0]", buf[0])
		return 0, ErrPacketTooLarge
	}

	log.Debug("Read: done read", "n", n, "e", err)

	return n, err
}

func (w *WebrtcConn) Write(b []byte) (int, error) {
	buf := make([]byte, PACKET_SIZE+1)
	lenPB := 0
	var pbuf []byte
	n := 0
	var err error

	eachN := 0

	log.Debug("Write: to write", "b", len(b))

	isEnd := false
looping:
	for pb := b; len(pb) > 0; pb = pb[lenPB:] {
		// 1. set lenPB
		isEnd = len(pb) <= PACKET_SIZE
		if isEnd {
			lenPB = len(pb)
		} else {
			lenPB = PACKET_SIZE
		}

		// 2. set buf[0]
		if isEnd {
			buf[0] = PACKET_END
		} else {
			buf[0] = PACKET_NOT_END
		}

		// 3. copy and set pbuf
		copy(buf[1:], pb[:lenPB])
		pbuf = buf[:lenPB+1]

		eachN, err = w.info.ReadWriteCloser.Write(pbuf)
		log.Debug("Write: (in-for-loop): after Write", "eachN", eachN, "e", err)
		if err != nil {
			break looping
		}
		n += eachN - 1

		// 4. break if no need to loop
		if len(pb) <= PACKET_SIZE {
			break looping
		}
	}

	log.Debug("Write: done write", "b", len(b), "n", n, "e", err)

	return n, err
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
