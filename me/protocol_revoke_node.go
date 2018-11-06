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
	"github.com/ailabstw/go-pttai/p2p/discover"
)

/*
RevokeNode intends to revoke a specific node
*/
func (pm *ProtocolManager) RevokeNode(nodeID *discover.NodeID) error {
	pm.lockMyNodes.RLock()
	defer pm.lockMyNodes.RUnlock()

	raftID, err := nodeID.ToRaftID()
	if err != nil {
		return err
	}

	_, ok := pm.MyNodes[raftID]
	if !ok {
		return ErrInvalidNode
	}

	err = pm.EnsureRaftLead()
	if err != nil {
		return err
	}

	return pm.ProposeRaftRemoveNode(nodeID)
}

func (pm *ProtocolManager) HandleRevokeOtherNode(oplog *MasterOplog, node *MyNode) error {
	peer := node.Peer
	if peer != nil {
		pm.UnregisterPeer(peer, true, false)
	}

	return nil
}

func (pm *ProtocolManager) HandleRevokeMyNode(oplog *MasterOplog, isLockedEntity bool, isLockedNode bool) error {

	// set entity (because the revoke is from master-oplog)
	pm.revokeMyNodeSetEntity(oplog, isLockedEntity)

	// clean raft
	pm.cleanRaft()

	// clean my nodes
	pm.revokeMyNodeCleanMyNodes(isLockedNode)

	// peer
	pm.cleanPeers()

	// me-oplog
	pm.CleanMeOplog()

	// op-key
	pm.CleanOpKey()

	pm.CleanOpKeyOplog()

	// revoke-key
	myID := pm.Entity().GetID()
	pttMyID := pm.Ptt().GetMyEntity().GetID()

	pm.Entity().Service().(*Backend).Config.RevokeMyKey(myID)

	if reflect.DeepEqual(myID, pttMyID) {
		pm.Entity().Service().(*Backend).Config.RevokeKey()
	}

	return nil
}

func (pm *ProtocolManager) revokeMyNodeSetEntity(oplog *MasterOplog, isLocked bool) {
	entity := pm.Entity().(*MyInfo)

	if !isLocked {
		entity.MustLock()
		defer entity.Unlock()
	}

	entity.LogID = oplog.ID
	entity.UpdateTS = oplog.UpdateTS
	entity.Status = types.StatusRevoked

	entity.MustSave(true)
}

func (pm *ProtocolManager) revokeMyNodeCleanMyNodes(isLocked bool) {
	if !isLocked {
		pm.lockMyNodes.Lock()
		defer pm.lockMyNodes.Unlock()
	}

	myID := pm.Entity().GetID()

	var nodeSignID *types.PttID
	for raftID, node := range pm.MyNodes {
		nodeSignID, _ = setNodeSignID(node.NodeID, myID)

		delete(pm.MyNodes, raftID)
		delete(pm.MyNodeByNodeSignIDs, *nodeSignID)
		node.Delete(true)
	}

}

func (pm *ProtocolManager) cleanPeers() {
	peerList := pm.Peers().PeerList(false)
	for _, peer := range peerList {
		pm.UnregisterPeer(peer, true, false)
	}
}

func (pm *ProtocolManager) cleanRaft() {
	pm.StopRaft()
	pm.CleanRaftStorage(false)
}
