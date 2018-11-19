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

package account

import (
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) GetPeerType(peer *pkgservice.PttPeer) pkgservice.PeerType {
	switch {
	case pm.IsMyDevice(peer):
		return pkgservice.PeerTypeMe
	case pm.IsMemberPeer(peer):
		return pkgservice.PeerTypeMember
	case pm.IsPendingPeer(peer):
		return pkgservice.PeerTypePending
	}
	return pkgservice.PeerTypeRandom
}

func (pm *ProtocolManager) IsMyDevice(peer *pkgservice.PttPeer) bool {
	return peer.PeerType == pkgservice.PeerTypeMe
}

func (pm *ProtocolManager) IsImportantPeer(peer *pkgservice.PttPeer) bool {
	return false
}

func (pm *ProtocolManager) IsMemberPeer(peer *pkgservice.PttPeer) bool {
	if peer.UserID == nil {
		return false
	}

	return pm.IsMember(peer.UserID, false)
}

/*
func (pm *ProtocolManager) IsPendingPeer(peer *pkgservice.PttPeer) bool {
	if peer.UserID == nil {
		return false
	}

	return pm.IsPendingMember(peer.UserID, false)
}
*/

func (pm *ProtocolManager) IsFitPeer(peer *pkgservice.PttPeer) pkgservice.PeerType {
	return pm.GetPeerType(peer)
}

func (pm *ProtocolManager) RegisterPeer(peer *pkgservice.PttPeer, peerType pkgservice.PeerType) error {
	pm.BaseProtocolManager.RegisterPeer(peer, peerType)

	if peerType != pkgservice.PeerTypePending {
		return nil
	}
	pm.postRegisterPeer(peer)

	return nil
}

func (pm *ProtocolManager) RegisterPendingPeer(peer *pkgservice.PttPeer) error {
	pm.BaseProtocolManager.RegisterPendingPeer(peer)

	pm.postRegisterPeer(peer)

	return nil
}

func (pm *ProtocolManager) postRegisterPeer(peer *pkgservice.PttPeer) error {
	return nil
}
