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

package me

import (
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) GetPeerType(peer *pkgservice.PttPeer) pkgservice.PeerType {
	switch {
	case pm.IsMyDevice(peer):
		return pkgservice.PeerTypeMe
	case pm.IsImportantPeer(peer):
		return pkgservice.PeerTypeImportant
	case pm.IsMemberPeer(peer):
		return pkgservice.PeerTypeMember
	case pm.IsPendingPeer(peer):
		return pkgservice.PeerTypePending
	}
	return pkgservice.PeerTypeRandom
}

func (pm *ProtocolManager) IsMyDevice(peer *pkgservice.PttPeer) bool {
	log.Debug("IsMyDevice: start")
	//pm.LockMyNodes.RLock()
	//defer pm.LockMyNodes.RUnlock()

	peerID := peer.GetID()
	raftID, err := peerID.ToRaftID()
	if err != nil {
		return false
	}

	myNode, ok := pm.MyNodes[raftID]
	if !ok {
		log.Debug("IsMyDevice: not my device")

		return false
	}

	myID := pm.Entity().GetID()

	if !reflect.DeepEqual(peer.UserID, myID) {
		return false
	}

	log.Debug("IsMyDevice: done", "myNode", myNode.Status)

	if myNode.Status != types.StatusAlive {
		return false
	}

	return true
}

func (pm *ProtocolManager) IsImportantPeer(peer *pkgservice.PttPeer) bool {
	return false
}

func (pm *ProtocolManager) IsMemberPeer(peer *pkgservice.PttPeer) bool {
	return false
}

func (pm *ProtocolManager) IsPendingPeer(peer *pkgservice.PttPeer) bool {
	log.Debug("IsPendingPeer: start")
	peerID := peer.GetID()
	raftID, err := peerID.ToRaftID()
	if err != nil {
		return false
	}

	myNode, ok := pm.MyNodes[raftID]
	if !ok {
		return false
	}

	myID := pm.Entity().GetID()

	if !reflect.DeepEqual(peer.UserID, myID) {
		return false
	}

	log.Debug("IsPendingPeer: done", "peer", peer, "myNode", myNode.Status)

	return myNode.Status < types.StatusAlive
}

func (pm *ProtocolManager) IsFitPeer(peer *pkgservice.PttPeer) pkgservice.PeerType {
	if pm.IsMyDevice(peer) {
		return pkgservice.PeerTypeMe
	}
	if pm.IsPendingPeer(peer) {
		return pkgservice.PeerTypePending
	}

	return pkgservice.PeerTypeRandom
}

func (pm *ProtocolManager) RegisterPeer(peer *pkgservice.PttPeer, peerType pkgservice.PeerType) error {
	pm.BaseProtocolManager.RegisterPeer(peer, peerType)

	pm.postRegisterPeer(peer)

	log.Debug("RegisterPeer: done", "peer", peer)

	return nil
}

func (pm *ProtocolManager) RegisterPendingPeer(peer *pkgservice.PttPeer) error {
	pm.BaseProtocolManager.RegisterPendingPeer(peer)

	log.Debug("RegisterPendingPeer: to postRegisterPeer", "peer", peer)

	pm.postRegisterPeer(peer)

	log.Debug("RegisterPendingPeer: done", "peer", peer)

	return nil
}

func (pm *ProtocolManager) postRegisterPeer(peer *pkgservice.PttPeer) error {
	peerID := peer.GetID()

	raftID, err := peerID.ToRaftID()
	if err != nil {
		return err
	}

	myNode := pm.MyNodes[raftID]
	myNode.Peer = peer
	if myNode.Status == types.StatusInternalPending {
		pm.InitMeInfo(peer)
	}

	return nil
}
