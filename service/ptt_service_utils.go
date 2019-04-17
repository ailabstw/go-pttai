// Copyright 2019 The go-pttai Authors
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
	"github.com/ethereum/go-ethereum/rpc"
)

func (p *BasePtt) GenerateProtocols() []p2p.Protocol {
	subProtocols := make([]p2p.Protocol, 0, len(ProtocolVersions))

	for i, version := range ProtocolVersions {
		protocol := p2p.Protocol{
			Name:     ProtocolName,
			Version:  version,
			Length:   ProtocolLengths[i],
			Run:      p.GenerateRun(version),
			NodeInfo: p.GenerateNodeInfo(),
			PeerInfo: p.GeneratePeerInfo(),
		}

		subProtocols = append(subProtocols, protocol)
	}

	return subProtocols
}

/*
GenerateRun generates run in Protocol (PttService)
(No need to do sync in the ptt-layer for now, because there is no information needs to sync in the ptt-layer.)

    1. set up ptt-peer.
    2. peerWG.
    3. handle-peer.
*/
func (p *BasePtt) GenerateRun(version uint) func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	return func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
		// 1. pttPeer
		pttPeer, err := p.NewPeer(version, peer, rw)
		log.Debug("GenerateRun: get new peer", "peer", pttPeer, "e", err)
		if err != nil {
			return err
		}

		// 2. peerWG
		p.peerWG.Add(1)
		defer p.peerWG.Done()

		// 3. handle peer
		err = p.HandlePeer(pttPeer)
		log.Debug("GenerateRun: after HandlePeer", "peer", pttPeer, "e", err)

		return err
	}
}

func (p *BasePtt) GenerateNodeInfo() func() interface{} {
	return func() interface{} {
		return p.NodeInfo()
	}
}

func (p *BasePtt) GeneratePeerInfo() func(id discover.NodeID) interface{} {
	return func(id discover.NodeID) interface{} {
		p.peerLock.RLock()
		defer p.peerLock.RUnlock()

		peer := p.GetPeer(&id, true)
		if peer == nil {
			return nil
		}

		return peer.Info()
	}
}

func (p *BasePtt) PttAPIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "ptt",
			Version:   "1.0",
			Service:   NewPrivateAPI(p),

			Public: IsPrivateAsPublic,
		},
	}
}

func (p *BasePtt) NodeInfo() interface{} {
	peers := len(p.myPeers) + len(p.importantPeers) + len(p.memberPeers) + len(p.randomPeers)
	var userID *types.PttID
	if p.myEntity != nil {
		userID = p.myEntity.GetID()
	}

	return &PttNodeInfo{
		NodeID:   p.myNodeID,
		UserID:   userID,
		Peers:    peers,
		Entities: len(p.entities),
		Services: len(p.services),
	}
}
