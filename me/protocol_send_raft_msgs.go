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

package me

import (
	"context"
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pb "github.com/ailabstw/etcd/raft/raftpb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SendRaftMsgs struct {
	Msgs []pb.Message `json:"M"`
}

func (pm *ProtocolManager) SendRaftMsgs(msgs []pb.Message) error {
	log.Debug("SendRaftMsgs: start", "msgs", msgs)
	myRaftID := pm.myPtt.MyRaftID()
	msgByPeers := make(map[uint64][]pb.Message)
	var origMsgByPeers []pb.Message
	for _, msg := range msgs {
		if msg.To == myRaftID {
			continue
		}
		origMsgByPeers = msgByPeers[msg.To]
		msgByPeers[msg.To] = append(origMsgByPeers, msg)
	}

	pm.lockMyNodes.RLock()
	defer pm.lockMyNodes.RUnlock()

	peers := pm.Peers()
	log.Debug("SendRaftMsgs: to for-loop", "peers", peers, "msgByPeers", msgByPeers)
	var data *SendRaftMsgs
	for raftID, eachMsgs := range msgByPeers {
		myNode := pm.MyNodes[raftID]
		if myNode == nil {
			log.Warn("SendRaftMsgs: unable to send peer not myNode", "raftID", raftID)
			continue
		}

		if myNode.Status == types.StatusInit || myNode.Status == types.StatusInternalPending {
			log.Warn("SendRaftMsgs: myNode status invalid", "raftID", raftID, "status", myNode.Status)
			continue
		}

		peer := peers.Peer(myNode.NodeID, false)
		log.Debug("SendRaftMsgs: after get Peer", "peer", peer)
		if peer == nil {
			continue
		}

		data = &SendRaftMsgs{
			Msgs: eachMsgs,
		}

		log.Debug("SendRaftMsgs: to send to Peer", "peer", peer)

		pm.SendDataToPeer(SendRaftMsgsMsg, data, peer)
	}

	return nil
}

func (pm *ProtocolManager) HandleSendRaftMsgs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	myInfo := pm.Entity().(*MyInfo)
	// defensive-programming
	if myInfo.Status == types.StatusInit || myInfo.Status == types.StatusInternalPending {
		return nil
	}

	if pm.raftNode == nil {
		return nil
	}

	// unmarshal
	data := &SendRaftMsgs{}
	err := json.Unmarshal(dataBytes, data)
	log.Debug("HandleSendRaftMsgs: start", "myID", myInfo.ID, "peer", peer.GetID(), "msgs", data.Msgs)
	if err != nil {
		return err
	}

	// step
	for _, msg := range data.Msgs {
		pm.raftNode.Step(context.TODO(), msg)
	}

	return nil
}
