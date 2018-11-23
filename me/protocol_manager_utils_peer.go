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
	case pm.IsPendingPeer(peer):
		return pkgservice.PeerTypePending
	}
	return pkgservice.PeerTypeRandom
}

func (pm *ProtocolManager) IsMyDevice(peer *pkgservice.PttPeer) bool {
	//pm.LockMyNodes.RLock()
	//defer pm.LockMyNodes.RUnlock()

	peerID := peer.GetID()
	raftID, err := peerID.ToRaftID()
	if err != nil {
		return false
	}

	myNode, ok := pm.MyNodes[raftID]
	if !ok {
		return false
	}

	log.Debug("IsMyDevice: to check status", "myNode.Status", myNode.Status)

	if myNode.Status < types.StatusSync {
		return false
	}

	myIDs := pm.Entity().GetOwnerIDs()

	log.Debug("IsMyDevice: to check ownerIDs", "myIDs", myIDs)

	for _, myID := range myIDs {

		log.Debug("IsMyDevice: to check ownerID", "myID", myID, "userID", peer.UserID)

		if reflect.DeepEqual(peer.UserID, myID) {
			return true
		}
	}

	return false
}

func (pm *ProtocolManager) IsImportantPeer(peer *pkgservice.PttPeer) bool {
	return false
}

func (pm *ProtocolManager) IsMemberPeer(peer *pkgservice.PttPeer) bool {
	return false
}

func (pm *ProtocolManager) IsPendingPeer(peer *pkgservice.PttPeer) bool {
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

	return myNode.Status < types.StatusSync
}

func (pm *ProtocolManager) RegisterPeer(peer *pkgservice.PttPeer, peerType pkgservice.PeerType) error {

	log.Debug("RegisterPeer: start", "peer", peer, "userID", peer.UserID, "peerType", peerType)
	pm.BaseProtocolManager.RegisterPeer(peer, peerType)

	pm.postRegisterPeer(peer)

	return nil
}

func (pm *ProtocolManager) RegisterPendingPeer(peer *pkgservice.PttPeer) error {
	pm.BaseProtocolManager.RegisterPendingPeer(peer)

	pm.postRegisterPeer(peer)

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

func (pm *ProtocolManager) LoadPeers() error {
	ptt := pm.myPtt
	opKey, err := pm.GetOldestOpKey(false)
	if err != nil {
		return err
	}

	myNodeID := ptt.MyNodeID()
	for _, myNode := range pm.MyNodes {
		if reflect.DeepEqual(myNode.NodeID, myNodeID) {
			continue
		}
		log.Debug("LoadPeers: to AddDial", "nodeID", myNode.NodeID)
		ptt.AddDial(myNode.NodeID, opKey.Hash, pkgservice.PeerTypeMe)
	}
	return nil
}
