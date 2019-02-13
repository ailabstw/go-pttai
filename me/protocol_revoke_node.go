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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
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

func (pm *ProtocolManager) HandleRevokeOtherNode(oplog *MasterOplog, node *MyNode, fromID *types.PttID) error {
	peer := node.Peer

	if peer != nil {
		pm.UnregisterPeer(peer, true, false, false)
	}

	myInfo := pm.Entity().(*MyInfo)
	myNodeSignID := myInfo.NodeSignID

	log.Debug("HandleRevokeOtherNode", "peer", peer, "fromID", fromID, "myNodeSignID", myNodeSignID, "myID", myInfo.ID)

	if reflect.DeepEqual(myNodeSignID, fromID) {
		myInfo.Profile.PM().(*account.ProtocolManager).RemoveUserNode(node.NodeID)
	}

	return nil
}

func (pm *ProtocolManager) HandleRevokeMyNode(oplog *MasterOplog, isLockedEntity bool, isLockedNode bool) error {

	log.Debug("HandleRevokeMyNode: start")

	// set entity (because the revoke is from master-oplog)
	pm.revokeMyNodeSetEntity(oplog, isLockedEntity)

	// clean raft
	pm.cleanRaft()

	// clean my nodes
	pm.revokeMyNodeCleanMyNodes(isLockedNode)

	// join-key
	pm.CleanJoinKey()

	// op-key
	pm.CleanOpKey()

	pm.CleanOpKeyOplog()

	// clean log0
	pm.CleanLog0()

	// peer
	pm.CleanPeers()

	// revoke-key
	myInfo := pm.Entity()
	myID := myInfo.GetID()

	pm.Entity().Service().(*Backend).Config.RevokeMyKey(myID)

	// entities

	myService := myInfo.Service()
	pttMyID := pm.Ptt().GetMyEntity().GetID()
	if reflect.DeepEqual(myID, pttMyID) {

		log.Debug("HandleRevokeMyNode: revoke pttMyID")

		entities := pm.myPtt.GetEntities()
		for _, entity := range entities {
			if entity == myInfo {
				continue
			}
			if entity.Service() == myService {
				continue
			}
			if entity.GetStatus() > types.StatusAlive {
				continue
			}
			entity.SetStatus(types.StatusRevoked)
			entity.SetUpdateTS(oplog.UpdateTS)
			entity.SetLogID(oplog.ID)
			entity.PM().PostdeleteEntity(nil, true)
		}

		pm.Entity().Service().(*Backend).Config.RevokeKey()

		pm.Entity().Service().(*Backend).Config.SetMyKey("", "", "", true)

		// stop
		pm.myPtt.NotifyNodeStop().PassChan(struct{}{})
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

func (pm *ProtocolManager) cleanRaft() {
	pm.StopRaft()
	pm.CleanRaftStorage(false)
}
