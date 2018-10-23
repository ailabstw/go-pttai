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
	"net"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

type BackendCountPeers struct {
	MyPeers        int `json:"M"`
	ImportantPeers int `json:"I"`
	MemberPeers    int `json:"E"`
	RandomPeers    int `json:"R"`
}

type BackendPeer struct {
	NodeID   *discover.NodeID `json:"ID"`
	PeerType PeerType         `json:"T"`
	UserID   *types.PttID     `json:"UID"`
	Addr     net.Addr
}

func PeerToBackendPeer(peer *PttPeer) *BackendPeer {
	return &BackendPeer{
		NodeID:   peer.GetID(),
		PeerType: peer.PeerType,
		UserID:   peer.UserID,
		Addr:     peer.RemoteAddr(),
	}
}
