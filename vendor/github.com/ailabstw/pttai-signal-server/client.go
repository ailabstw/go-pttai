package signalserver

import (
	"crypto/ecdsa"
	"net/url"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/gorilla/websocket"
)

type Client struct {
	nodeID discv5.NodeID
	Conn   *Conn
}

/*
Send sends the normal-msg to toID.
*/
func (c *Client) Send(toID discv5.NodeID, msg []byte, extra []byte) error {
	signal := &Signal{FromID: c.nodeID, ToID: toID, Msg: msg, Extra: extra}

	err := c.Conn.WsConn.WriteJSON(signal)
	if err != nil {
		return err
	}

	return nil
}

/*
Receive receives the normal-msg from signal-server.
*/
func (c *Client) Receive() (*Signal, error) {
	signal := &Signal{}
	err := c.Conn.WsConn.ReadJSON(signal)
	if err != nil {
		return nil, err
	}

	if c.nodeID != signal.ToID {
		return nil, ErrInvalidID
	}

	return signal, nil
}

/*
NewClient init a new client and pass the challenge from the signal-server.
*/
func NewClient(nodeID discv5.NodeID, privKey *ecdsa.PrivateKey, url url.URL) (*Client, error) {
	wsConn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Error("NewClient: unable to dial", "e", err, "url", url)
		return nil, err
	}

	c := &challenge{}
	err = wsConn.ReadJSON(c)
	if err != nil {
		log.Error("NewCleint: unable to ReadJSON", "e", err)
		return nil, err
	}

	resp, err := respondChallenge(nodeID, privKey, c)
	if err != nil {
		log.Error("NewClient: unable to respond Challenge", "e", err)
		return nil, err
	}

	err = wsConn.WriteJSON(resp)
	if err != nil {
		log.Error("NewClient: unable to WriteJSON", "e", err)
		return nil, err
	}

	cack := &challengeAck{}
	err = wsConn.ReadJSON(cack)
	if err != nil {
		log.Error("NewClient: unable to ReadJSON from ack", "e", err)
		return nil, err
	}

	if cack.NodeID != nodeID {
		log.Error("NewClient: invalid id", "cack", cack.NodeID, "nodeID", nodeID)
		return nil, ErrInvalidID
	}

	conn := &Conn{isClosed: 0, WsConn: wsConn}

	return &Client{nodeID: nodeID, Conn: conn}, nil
}

func respondChallenge(nodeID discv5.NodeID, privKey *ecdsa.PrivateKey, c *challenge) (*challengeResponse, error) {

	hash := crypto.Keccak256Hash(c.Challenge)

	sig, err := crypto.Sign(hash[:], privKey)
	if err != nil {
		return nil, err
	}

	challengeResponse := &challengeResponse{NodeID: nodeID, Signature: sig, Hash: hash}

	return challengeResponse, nil
}

func (c *Client) Close() {
	c.Conn.Close()
}
