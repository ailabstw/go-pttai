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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

// sync-oplog
type SyncOplog struct {
	ToSyncTime  types.Timestamp `json:"LT"`
	ToSyncNodes []*MerkleNode   `json:"LN"`
}

/*
SyncOplog: I initiate sync-oplog.

Expected merkle-tree-list length: 24 (hour) + 31 (day) + 12 (month) + n (year)
(should be within the packet-limit)
*/
func (pm *BaseProtocolManager) SyncOplog(peer *PttPeer, merkle *Merkle, op OpType) error {

	if peer == nil {
		return nil
	}

	if !peer.IsRegistered {
		return nil
	}

	ptt := pm.Ptt()
	myInfo := ptt.GetMyEntity()
	if myInfo.GetStatus() != types.StatusAlive {
		return nil
	}

	entity := pm.Entity()
	if entity.GetStatus() != types.StatusAlive {
		return nil
	}

	_, err := pm.GetOldestOpKey(false)
	if err != nil {
		return pm.Ptt().RequestOpKeyByEntity(entity, peer)
	}

	toSyncTime, err := merkle.ToSyncTime()
	if err != nil {
		return err
	}

	toSyncNodes, _, err := merkle.GetMerkleTreeList(toSyncTime, false)
	if err != nil {
		return err
	}

	syncOplog := &SyncOplog{
		ToSyncTime:  toSyncTime,
		ToSyncNodes: toSyncNodes,
	}

	err = pm.SendDataToPeer(op, syncOplog, peer)
	if err != nil {
		return err
	}

	return nil
}

/*
HandleSyncOplog: I received sync-oplog. (MerkleTreeList should be within the packet-limit.)

	1. get my merkle-tree-list.
	2. validate merkle tree
	3. SyncOplogAck
*/
func (pm *BaseProtocolManager) HandleSyncOplog(
	dataBytes []byte,
	peer *PttPeer,
	merkle *Merkle,

	forceSyncOplogMsg OpType,
	forceSyncOplogAckMsg OpType,
	invalidOplogMsg OpType,
	syncOplogAckMsg OpType,
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

	data := &SyncOplog{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	merkleName := GetMerkleName(merkle, pm)

	myToSyncTime, err := merkle.ToSyncTime()
	log.Debug("HandleSyncOplog: after get myToSyncTime", "e", err, "myToSyncTime", myToSyncTime, "data.ToSyncTime", data.ToSyncTime, "merkle", merkleName)
	if err != nil {
		return err
	}

	if myToSyncTime.IsLess(data.ToSyncTime) {
		return nil
	}

	toSyncTime := myToSyncTime
	if data.ToSyncTime.IsLess(toSyncTime) {
		toSyncTime = data.ToSyncTime
	}

	// get my merkle-tree-list.
	myToSyncNodes, _, err := merkle.GetMerkleTreeList(toSyncTime, false)
	log.Debug("HandleSyncOplog: after getMerkleTreeList", "myToSyncNodes", myToSyncNodes, "data.ToSyncNodes", data.ToSyncNodes, "merkle", merkleName)
	if err != nil {
		return err
	}

	// 2. validate merkle tree
	myNewNodes, theirNewNodes, err := DiffMerkleTree(myToSyncNodes, data.ToSyncNodes, toSyncTime, pm, merkle)
	if err != nil {
		return err
	}

	if len(myNewNodes) > 0 || len(theirNewNodes) > 0 {
		log.Warn("HandleSyncOplog: invalid merkle", "myNewNodes", len(myNewNodes), "theirNewNodes", len(theirNewNodes), "merkle", merkleName, "peer", peer)
		return pm.SyncOplogInvalidByMerkle(
			myNewNodes,
			theirNewNodes,

			forceSyncOplogMsg,
			forceSyncOplogAckMsg,

			merkle,

			peer,
		)
	}

	// 3. SyncOplogAck
	return pm.SyncOplogAck(toSyncTime, myToSyncTime, merkle, forceSyncOplogAckMsg, syncOplogAckMsg, peer)
}
