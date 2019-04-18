package signalserver

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/gorilla/websocket"
)

type challenge struct {
	Challenge []byte `json:"C"`
}

type challengeResponse struct {
	NodeID discv5.NodeID

	Signature []byte
	Hash      [32]byte
}

type challengeAck struct {
	NodeID discv5.NodeID
}

type Server struct {
	nodeChannels sync.Map

	nodeChannelsWriteLock sync.Mutex

	upgrader websocket.Upgrader
}

func (s *Server) writeLoop(nc *NodeConn) error {
looping:
	for {
		select {
		case signal, ok := <-nc.writeChan:
			if !ok {
				break looping
			}

			err := nc.Conn.WsConn.WriteJSON(signal)
			if err != nil {
				return err
			}
		case <-nc.quitChan:
			break looping
		}
	}

	return nil
}

func (s *Server) readLoop(nc *NodeConn) error {
	for {
		signal := &Signal{}
		err := nc.Conn.WsConn.ReadJSON(signal)
		if err != nil {
			return err
		}

		err = s.dispatch(signal)
		if err != nil {
			return err
		}
	}
}

func (s *Server) dispatch(signal *Signal) error {
	if nc, ok := s.nodeChannels.Load(signal.ToID); ok {
		(nc.(*NodeConn)).writeChan <- signal
	}
	return nil
}

func NewServer() *Server {
	return &Server{
		nodeChannels: sync.Map{},
		upgrader:     websocket.Upgrader{},
	}
}

func (s *Server) generateChallenge() []byte {
	challenge := make([]byte, 256)
	io.ReadFull(rand.Reader, challenge)

	return challenge
}

func (s *Server) verifyNode(challenge []byte, resp *challengeResponse) error {
	if resp.Hash != crypto.Keccak256Hash(challenge) {
		return fmt.Errorf("hash incorrect from node %s", resp.NodeID)
	}

	publicKey, err := resp.NodeID.Pubkey()
	if err != nil {
		return err
	}

	// check signature match nodeID(public key)
	verified := crypto.VerifySignature(crypto.FromECDSAPub(publicKey), resp.Hash[:], resp.Signature[:64])
	if !verified {
		return fmt.Errorf("unable to verify signature from node %s", resp.NodeID)
	}

	return nil
}

func (s *Server) identifyNodeID(conn *Conn) (discv5.NodeID, error) {
	c := s.generateChallenge()

	tmpC := &challenge{Challenge: c}

	// send challenge to conn
	err := conn.WsConn.WriteJSON(tmpC)
	// log.Printf("server.identifyNodeID: after WriteJSON signal: %v, e: %v", signal, err)
	if err != nil {
		return discv5.NodeID{}, err
	}

	resp := &challengeResponse{}
	err = conn.WsConn.ReadJSON(resp)
	if err != nil {
		return discv5.NodeID{}, err
	}

	err = s.verifyNode(c, resp)
	if err != nil {
		return discv5.NodeID{}, err
	}

	return resp.NodeID, nil
}

func (s *Server) newNodeConn(nodeID discv5.NodeID, wsConn *Conn) (*NodeConn, error) {
	// check already exists
	s.nodeChannelsWriteLock.Lock()
	defer s.nodeChannelsWriteLock.Unlock()

	if origConn, exists := s.nodeChannels.Load(nodeID); exists {
		(origConn.(*NodeConn)).Conn.Close()
	}

	nc := NewNodeConn(nodeID, wsConn)
	s.nodeChannels.Store(nodeID, nc)

	return nc, nil
}

// SignalHandler will
func (s *Server) SignalHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: http error handler
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn := Conn{0, wsConn}
	defer func() {
		conn.Close()
	}()

	// 1. authendication
	nodeID, err := s.identifyNodeID(&conn)
	if err != nil {
		return
	}

	// create a NodeConn, which will create a read loop goroutine for the websocket connection
	nodeConn, err := s.newNodeConn(nodeID, &conn)
	if err != nil {
		return
	}
	defer func() {
		s.removeFromNodeChannels(nodeConn)
	}()

	cack := &challengeAck{NodeID: nodeID}
	nodeConn.Conn.WsConn.WriteJSON(cack)

	// write loop
	go s.writeLoop(nodeConn)

	// websocket read loop
	s.readLoop(nodeConn)
}

func (s *Server) removeFromNodeChannels(nodeConn *NodeConn) {

	close(nodeConn.quitChan)

	s.nodeChannelsWriteLock.Lock()
	defer s.nodeChannelsWriteLock.Unlock()

	if origConn, exists := s.nodeChannels.Load(nodeConn.NodeID); exists && origConn.(*NodeConn) == nodeConn {
		s.nodeChannels.Delete(nodeConn.NodeID)
	}
}
