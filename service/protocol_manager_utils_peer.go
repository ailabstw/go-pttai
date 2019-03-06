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
	"github.com/ailabstw/go-pttai/log"
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

func (pm *BaseProtocolManager) RegisterPeer(peer *PttPeer, peerType PeerType, isLocked bool) (err error) {
	log.Debug("RegisterPeer: start", "peer", peer, "peerType", peerType)

	if !isLocked {
		pm.Peers().Lock()
		defer pm.Peers().Unlock()
	}

	if peerType == PeerTypeRandom {
		return nil
	}

	if peerType == PeerTypePending {
		return pm.RegisterPendingPeer(peer, true)
	}

	// We just primitively check the existence without lock
	// to avoid the deadlock in chan.
	// The consequence of entering race-condition is just doing sync multiple-times.
	origPeer := pm.Peers().Peer(peer.GetID(), true)
	if origPeer != nil {
		return pm.Peers().Register(peer, peerType, true)
	}
	if !pm.isStart {
		return nil
	}

	log.Debug("RegisterPeer: to NewPeerCh", "peer", peer, "peerType", peerType, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name(), "status", pm.Entity().GetStatus())

	select {
	case pm.NewPeerCh() <- peer:
		err = pm.Peers().Register(peer, peerType, true)
	case <-pm.NoMorePeers():
		err = p2p.DiscQuitting
	}

	log.Debug("RegisterPeer: after NewPeerCh", "e", err, "peer", peer, "peerType", peerType, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	return err
}

func (pm *BaseProtocolManager) RegisterPendingPeer(peer *PttPeer, isLocked bool) error {
	return pm.peers.Register(peer, PeerTypePending, isLocked)
}

func (pm *BaseProtocolManager) UnregisterPeer(peer *PttPeer, isForceReset bool, isForceNotReset bool, isPttLocked bool) error {

	peerType := pm.GetPeerType(peer)

	log.Debug("UnregisterPeer: to peers.Unregister", "peer", peer, "peerType", peerType)

	err := pm.peers.Unregister(peer, false)
	if err != nil {
		return err
	}

	if isForceNotReset {
		return nil
	}

	if !isForceReset && peerType < peer.PeerType {
		return nil
	}

	pm.Ptt().ResetPeerType(peer, isPttLocked, isForceReset)

	return nil
}

func (pm *BaseProtocolManager) UnregisterPeerByOtherUserID(id *types.PttID, isResetPeerType bool, isPttLocked bool) error {

	peer, peerType, err := pm.peers.UnregisterPeerByOtherUserID(id, false)
	if err != nil {
		return err
	}

	if peer == nil {
		return nil
	}

	if !isResetPeerType && peerType < peer.PeerType {
		return nil
	}

	pm.Ptt().ResetPeerType(peer, isPttLocked, isResetPeerType)

	return nil
}

func (pm *BaseProtocolManager) GetPeerType(peer *PttPeer) PeerType {
	return pm.getPeerType(peer)
}

func (pm *BaseProtocolManager) defaultGetPeerType(peer *PttPeer) PeerType {
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

func (pm *BaseProtocolManager) IsMyDevice(peer *PttPeer) bool {
	return pm.isMyDevice(peer)
}
func (pm *BaseProtocolManager) defaultIsMyDevice(peer *PttPeer) bool {
	return peer.PeerType == PeerTypeMe
}

func (pm *BaseProtocolManager) IsImportantPeer(peer *PttPeer) bool {
	return pm.isImportantPeer(peer)
}

func (pm *BaseProtocolManager) defaultIsImportantPeer(peer *PttPeer) bool {
	if peer.UserID == nil {
		return false
	}

	return pm.isMaster(peer.UserID, false)
}

func (pm *BaseProtocolManager) IsMemberPeer(peer *PttPeer) bool {
	return pm.isMemberPeer(peer)
}

func (pm *BaseProtocolManager) defaultIsMemberPeer(peer *PttPeer) bool {
	if peer.UserID == nil {
		return false
	}

	return pm.IsMember(peer.UserID, false)
}

func (pm *BaseProtocolManager) IsPendingPeer(peer *PttPeer) bool {
	return pm.isPendingPeer(peer)
}

func (pm *BaseProtocolManager) GetPendingPeerByUserID(id *types.PttID, isLocked bool) (*PttPeer, error) {
	return pm.Peers().GetPendingPeerByUserID(id, isLocked)
}

func (pm *BaseProtocolManager) defaultIsPendingPeer(peer *PttPeer) bool {
	return pm.Peers().IsPendingPeer(peer, false)
}

func (pm *BaseProtocolManager) IsSuspiciousID(id *types.PttID, nodeID *discover.NodeID) bool {
	return false
}

func (pm *BaseProtocolManager) IsGoodID(id *types.PttID, nodeID *discover.NodeID) bool {
	return true
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

func (pm *BaseProtocolManager) CleanPeers() {
	peerList := pm.Peers().PeerList(false)
	for _, peer := range peerList {
		pm.UnregisterPeer(peer, false, false, false)
	}

	pendingPeerList := pm.Peers().PendingPeerList(false)
	for _, peer := range pendingPeerList {
		pm.UnregisterPeer(peer, false, false, false)
	}

	peers, _ := pm.GetPeers()

	log.Debug("CleanPeers: done", "entity", pm.Entity().GetID(), "peers", peers)
}
