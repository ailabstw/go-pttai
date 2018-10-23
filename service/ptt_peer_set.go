// Copyright 2018 The go-pttai Authors
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

package service

import (
	"sync"

	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

type PttPeerSet struct {
	peers  map[discover.NodeID]*PttPeer
	lock   sync.RWMutex
	closed bool
}

func NewPttPeerSet() (*PttPeerSet, error) {
	return &PttPeerSet{
		peers: make(map[discover.NodeID]*PttPeer),
	}, nil
}

func (ps *PttPeerSet) Peers() map[discover.NodeID]*PttPeer {
	return ps.peers
}

func (ps *PttPeerSet) Lock() *sync.RWMutex {
	return &ps.lock
}

func (ps *PttPeerSet) IsClosed() bool {
	return ps.closed
}

func (ps *PttPeerSet) Register(peer *PttPeer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return ErrClosed
	}

	id := peer.ID()
	if origPeer, ok := ps.peers[id]; ok {
		if origPeer == peer {
			return nil
		} else {
			return ErrAlreadyRegistered
		}
	}
	ps.peers[id] = peer

	return nil
}

func (ps *PttPeerSet) Unregister(peer *PttPeer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	id := peer.ID()
	_, ok := ps.peers[id]
	if !ok {
		return ErrNotRegistered
	}
	delete(ps.peers, id)

	return nil
}

func (ps *PttPeerSet) UnregisterPeers() error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for k := range ps.peers {
		delete(ps.peers, k)
	}

	return nil
}

func (ps *PttPeerSet) PeerList() []*PttPeer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	lenPeers := len(ps.peers)

	peerList := make([]*PttPeer, lenPeers)
	i := 0
	for _, peer := range ps.peers {
		peerList[i] = peer
		i++
	}

	return peerList
}

// Peer retrieves the registered peer with the given id.
func (ps *PttPeerSet) Peer(id *discover.NodeID) *PttPeer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[*id]
}

// Len returns if the current number of peers in the set.
func (ps *PttPeerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// Close disconnects all peers.
// No new peers can be registered after Close has returned.
func (ps *PttPeerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.Disconnect(p2p.DiscQuitting)
	}
	ps.closed = true
}
