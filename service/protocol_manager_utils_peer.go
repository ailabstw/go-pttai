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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

func (pm *BaseProtocolManager) Peers() *PttPeerSet {
	return pm.peers
}

func (pm *BaseProtocolManager) NewPeerCh() chan *PttPeer {
	return pm.newPeerCh
}

func (pm *BaseProtocolManager) SetNoMorePeers(noMorePeers chan struct{}) {
	pm.noMorePeers = noMorePeers
}

func (pm *BaseProtocolManager) NoMorePeers() chan struct{} {
	return pm.noMorePeers
}

func (pm *BaseProtocolManager) RegisterPeer(peer *PttPeer) error {
	peerType := pm.GetPeerType(peer)
	if peerType == PeerTypeRandom {
		return nil
	}

	select {
	case pm.NewPeerCh() <- peer:
		return pm.peers.Register(peer, peerType, false)
	case <-pm.NoMorePeers():
		return p2p.DiscQuitting
	}
}

func (pm *BaseProtocolManager) UnregisterPeer(peer *PttPeer) error {
	return pm.peers.Unregister(peer, false)
}

func (pm *BaseProtocolManager) GetPeerType(peer *PttPeer) PeerType {
	switch {
	case peer.PeerType == PeerTypeMe:
		return PeerTypeMe
	case pm.IsImportantPeer(peer):
		return PeerTypeImportant
	case pm.IsMemberPeer(peer):
		return PeerTypeMember
	}
	return PeerTypeRandom
}

func (pm *BaseProtocolManager) CountPeers() (int, error) {
	peerList := pm.peers.PeerList(false)
	return len(peerList), nil
}

func (pm *BaseProtocolManager) GetPeers() ([]*PttPeer, error) {
	peerList := pm.peers.PeerList(false)
	return peerList, nil

}

func (b *BaseProtocolManager) IsMyDevice(peer *PttPeer) bool {
	return false
}
func (b *BaseProtocolManager) IsImportantPeer(peer *PttPeer) bool {
	return false
}
func (b *BaseProtocolManager) IsMemberPeer(peer *PttPeer) bool {
	return false
}

func (pm *BaseProtocolManager) IsSuspiciousID(id *types.PttID, nodeID *discover.NodeID) bool {
	return false
}

func (pm *BaseProtocolManager) IsGoodID(id *types.PttID, nodeID *discover.NodeID) bool {
	return true
}
