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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type SyncOplogAck struct {
	TS    types.Timestamp
	Nodes []*MerkleNode `json:"N"`

	StartHourTS types.Timestamp `json:"STS"`
	EndHourTS   types.Timestamp `json:"ETS"`
	StartTS     types.Timestamp `json:"sTS"`
	EndTS       types.Timestamp `json:"eTS"`
}

/*
SyncOplogAck: sending SyncOplogAck. Passed validating the oplogs until toSyncTime.
	=> get the merkleNodes with level as MerkleTreeLevelNow from offsetHoursTS to now (blocked).
	=> Send the merkleNodes to the peer.
*/
func (pm *BaseProtocolManager) SyncOplogAck(
	toSyncTime types.Timestamp,
	merkle *Merkle,

	syncOplogAckMsg OpType,

	peer *PttPeer,
) error {

	now, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	nodes := make([]*MerkleNode, 0, MaxSyncOplogAck)

	if toSyncTime.IsEqual(types.ZeroTimestamp) {
		toSyncTime = pm.Entity().GetCreateTS()
	}

	offsetHourTS, _ := toSyncTime.ToHRTimestamp()
	startHourTS := offsetHourTS
	nextHourTS := offsetHourTS

	for currentHourTS := offsetHourTS; currentHourTS.IsLess(now); currentHourTS = nextHourTS {
		nodes, startHourTS, nextHourTS, err = pm.syncOplogAckCore(
			merkle,
			peer,
			syncOplogAckMsg,
			nodes,
			currentHourTS,
			offsetHourTS,
			startHourTS,
			now,
		)
		log.Debug("SyncOplogAck: (in-for-loop) after syncOplogAckCore", "offsetHourTS", offsetHourTS, "currentHourTS", currentHourTS, "now", now, "nodes", len(nodes))
		if err != nil {
			return err
		}
	}

	// deal with the last part of the nodes.
	if len(nodes) == 0 {
		return nil
	}

	syncOplogAck := &SyncOplogAck{
		TS:          offsetHourTS,
		Nodes:       nodes,
		StartHourTS: startHourTS,
		EndHourTS:   now,
		StartTS:     startHourTS,
		EndTS:       now,
	}

	err = pm.SendDataToPeer(syncOplogAckMsg, syncOplogAck, peer)
	if err != nil {
		return err
	}

	return nil
}

/*
syncOplogAckCore: core of SyncOplogAck: (sync-node within 1 hour (currentHourTS))

	1. get the nodes of current hour.
	2. if not reaching the limit: concat and return.
	3. send the origin nodes and reset the nodes.
	4. if the new nodes not reaching the limit: concat and return.
	5. send blocks of the new-nodes.
*/
func (pm *BaseProtocolManager) syncOplogAckCore(
	merkle *Merkle,
	peer *PttPeer,

	syncOplogAckMsg OpType,

	nodes []*MerkleNode,

	currentHourTS types.Timestamp,
	offsetHourTS types.Timestamp,
	startHourTS types.Timestamp,
	now types.Timestamp,

) ([]*MerkleNode, types.Timestamp, types.Timestamp, error) {

	nextHourTS := currentHourTS.NextHourTS()
	if now.IsLess(nextHourTS) {
		nextHourTS = now
	}

	// 1. get the nodes of current hour.
	eachNodes, err := merkle.GetMerkleTreeListByLevel(MerkleTreeLevelNow, currentHourTS, nextHourTS)
	if err != nil {
		return nil, types.ZeroTimestamp, types.ZeroTimestamp, err
	}

	// 2. if not reaching the limit: concat and return.
	if len(nodes)+len(eachNodes) < MaxSyncOplogAck {
		nodes = append(nodes, eachNodes...)
		return nodes, startHourTS, nextHourTS, nil
	}

	// 3. send the origin nodes and reset the nodes.
	var syncOplogAck *SyncOplogAck
	if len(nodes) > 0 {
		syncOplogAck = &SyncOplogAck{
			TS:          offsetHourTS,
			Nodes:       nodes,
			StartHourTS: startHourTS,
			EndHourTS:   currentHourTS,
			StartTS:     startHourTS,
			EndTS:       currentHourTS,
		}

		log.Debug("syncOplogAckCore: to SendDataToPeer", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name(), "nodes", nodes, "offsetHourTS", offsetHourTS, "startHourTS", startHourTS, "currentHourTS", currentHourTS)

		err = pm.SendDataToPeer(syncOplogAckMsg, syncOplogAck, peer)
		if err != nil {
			return nil, types.ZeroTimestamp, types.ZeroTimestamp, err
		}

		nodes = make([]*MerkleNode, 0, MaxSyncObjectAck)
	}

	// 4. if not reaching the limit: concat and return.
	if len(eachNodes) < MaxSyncObjectAck {
		nodes = append(nodes, eachNodes...)
		return nodes, currentHourTS, nextHourTS, nil
	}

	// 5. send blocks of the new nodes.
	pNodes := eachNodes
	lenEachNodes := 0
	var eachEachNodes []*MerkleNode
	startTS := currentHourTS
	endTS := currentHourTS
	for len(pNodes) > 0 {
		lenEachNodes = MaxSyncOplogAck
		if lenEachNodes > len(pNodes) {
			lenEachNodes = len(pNodes)
		}

		eachEachNodes, pNodes = pNodes[:lenEachNodes], pNodes[lenEachNodes:]

		endTS = nextHourTS
		if len(pNodes) > 0 {
			endTS = pNodes[0].UpdateTS
		}

		syncOplogAck = &SyncOplogAck{
			TS:          offsetHourTS,
			Nodes:       eachEachNodes,
			StartHourTS: currentHourTS,
			EndHourTS:   nextHourTS,
			StartTS:     startTS,
			EndTS:       endTS,
		}

		log.Debug("syncOplogAckCore: to SendDataToPeer (in-for-loop)", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name(), "nodes", eachEachNodes, "offsetHourTS", offsetHourTS, "StartHourTS (currentHourTS)", currentHourTS, "EndHourTS (nextHourTS)", nextHourTS, "StartTS", startTS, "EndTS", endTS)

		err = pm.SendDataToPeer(syncOplogAckMsg, syncOplogAck, peer)
		if err != nil {
			return nil, types.ZeroTimestamp, types.ZeroTimestamp, err
		}

		startTS = endTS
	}

	return nodes, nextHourTS, nextHourTS, nil
}

func (pm *BaseProtocolManager) HandleSyncOplogAck(
	dataBytes []byte,
	peer *PttPeer,
	merkle *Merkle,
	setDB func(oplog *BaseOplog),
	setNewestOplog func(log *BaseOplog) error,
	postsync func(peer *PttPeer) error,
	newLogsMsg OpType,
) error {

	ptt := pm.Ptt()
	myInfo := ptt.GetMyEntity()
	if myInfo.GetStatus() != types.StatusAlive {
		return nil
	}

	e := pm.Entity()
	if e.GetStatus() != types.StatusAlive {
		return nil
	}

	data := &SyncOplogAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	myNodes, err := merkle.GetMerkleTreeListByLevel(MerkleTreeLevelNow, data.StartHourTS, data.EndHourTS)
	if err != nil {
		return err
	}

	myNodes = pm.handleSyncOplogAckFilterTS(myNodes, data.StartTS, data.EndTS)

	myNewKeys, theirNewKeys, err := MergeMerkleNodeKeys(myNodes, data.Nodes)
	if err != nil {
		return err
	}

	log.Debug("HandleSyncOplogAck: to SyncOplogNewOplogs", "myNodes", myNodes, "myNewKeys", len(myNewKeys), "theirNewKeys", len(theirNewKeys), "newLogsMsg", newLogsMsg, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	return pm.SyncOplogNewOplogs(data, myNewKeys, theirNewKeys, peer, setDB, setNewestOplog, postsync, newLogsMsg)
}

func (pm *BaseProtocolManager) handleSyncOplogAckFilterTS(nodes []*MerkleNode, startTS types.Timestamp, endTS types.Timestamp) []*MerkleNode {
	newNodes := make([]*MerkleNode, 0, len(nodes))

	for _, node := range nodes {
		if startTS.IsLessEqual(node.UpdateTS) && node.UpdateTS.IsLess(endTS) {
			newNodes = append(newNodes, node)
		}
	}

	return newNodes
}
