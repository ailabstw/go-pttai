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

func (pm *BaseProtocolManager) RegisterPeer(peer *PttPeer, peerType PeerType) error {
	if peerType == PeerTypeRandom {
		return nil
	}

	if peerType == PeerTypePending {
		return pm.RegisterPendingPeer(peer)
	}

	select {
	case pm.NewPeerCh() <- peer:
		return pm.Peers().Register(peer, peerType, false)
	case <-pm.NoMorePeers():
		return p2p.DiscQuitting
	}
}

func (pm *BaseProtocolManager) RegisterPendingPeer(peer *PttPeer) error {
	return pm.peers.Register(peer, PeerTypePending, false)
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
	case pm.IsPendingPeer(peer):
		return PeerTypePending
	}
	return PeerTypeRandom
}

func (pm *BaseProtocolManager) CountPeers() (int, error) {
	pm.peers.RLock()
	defer pm.peers.RUnlock()

	peerList := pm.peers.PeerList(true)
	pendingPeerList := pm.peers.PendingPeerList(true)
	return len(peerList) + len(pendingPeerList), nil
}

func (pm *BaseProtocolManager) GetPeers() ([]*PttPeer, error) {
	pm.peers.RLock()
	defer pm.peers.RUnlock()

	peerList := pm.peers.PeerList(true)
	pendingPeerList := pm.peers.PendingPeerList(true)
	return append(peerList, pendingPeerList...), nil
}

func (pm *BaseProtocolManager) IsMyDevice(peer *PttPeer) bool {
	return pm.Entity().PM().IsMyDevice(peer)
}
func (pm *BaseProtocolManager) IsImportantPeer(peer *PttPeer) bool {
	return pm.Entity().PM().IsImportantPeer(peer)
}
func (pm *BaseProtocolManager) IsMemberPeer(peer *PttPeer) bool {
	return pm.Entity().PM().IsMemberPeer(peer)
}
func (pm *BaseProtocolManager) IsPendingPeer(peer *PttPeer) bool {
	return pm.Entity().PM().IsPendingPeer(peer)
}

func (pm *BaseProtocolManager) IsSuspiciousID(id *types.PttID, nodeID *discover.NodeID) bool {
	return false
}

func (pm *BaseProtocolManager) IsGoodID(id *types.PttID, nodeID *discover.NodeID) bool {
	return true
}
