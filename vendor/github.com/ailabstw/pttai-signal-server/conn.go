package signalserver

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/gorilla/websocket"
)

type Conn struct {
	isClosed int32
	WsConn   *websocket.Conn
}

func (conn *Conn) Close() {
	isSwapped := atomic.CompareAndSwapInt32(&conn.isClosed, 0, 1)
	if !isSwapped {
		return
	}

	conn.WsConn.Close()
}

type NodeConn struct {
	NodeID discv5.NodeID

	Conn *Conn

	writeChan chan *Signal
	quitChan  chan struct{}
}

func NewNodeConn(nodeID discv5.NodeID, conn *Conn) *NodeConn {
	w := make(chan *Signal)
	q := make(chan struct{})

	nc := &NodeConn{
		NodeID: nodeID,

		Conn: conn,

		writeChan: w,
		quitChan:  q,
	}

	return nc
}
