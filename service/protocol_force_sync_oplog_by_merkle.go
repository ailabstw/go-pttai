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

type ForceSyncOplogByMerkle struct {
	UpdateTS types.Timestamp `json:"UT"`
	Level    MerkleTreeLevel `json:"L"`
	Nodes    []*MerkleNode   `json:"K"`
}

func (pm *BaseProtocolManager) ForceSyncOplogByMerkle(
	myNewNode *MerkleNode,

	forceSyncOplogMsg OpType,

	merkle *Merkle,

	peer *PttPeer,
) error {

	myNode, _ := merkle.GetNodeByLevelTS(myNewNode.Level, myNewNode.UpdateTS)

	nodes, err := merkle.GetChildNodes(myNewNode.Level, myNewNode.UpdateTS)
	if err != nil {
		return err
	}

	merkleName := GetMerkleName(merkle, pm)
	if myNode != nil && len(nodes) != int(myNode.NChildren) {
		log.Warn("ForceSyncOplogByMerkle: len != NChildren", "len", len(nodes), "children", myNode.NChildren, "merkle", merkleName)
		err = merkle.TryForceSync(pm)
		if err != nil {
			log.Error("ForceSyncOplogByMerkle: len != NChildren (unable to sync)", "len", len(nodes), "children", myNode.NChildren, "merkle", merkleName)
			return err
		}
		return nil
	}

	if myNode == nil && len(nodes) != 0 {
		log.Warn("ForceSyncOplogByMerkle: len != 0", "len", len(nodes), "level", myNewNode.Level, "ts", myNewNode.UpdateTS, "merkle", merkleName)
		err = merkle.TryForceSync(pm)
		if err != nil {
			log.Error("ForceSyncOplogByMerkle: len != 0 (unable to sync)", "len", len(nodes), "level", myNewNode.Level, "ts", myNewNode.UpdateTS, "merkle", merkleName)
			return err
		}
		return nil
	}

	data := &ForceSyncOplogByMerkle{
		UpdateTS: myNewNode.UpdateTS,
		Level:    myNewNode.Level,
		Nodes:    nodes,
	}

	log.Debug("ForceSyncOplogByMerkle: to SendDataToPeer", "children", len(nodes), "level", myNewNode.Level, "ts", myNewNode.UpdateTS, "merkle", merkleName)

	err = pm.SendDataToPeer(forceSyncOplogMsg, data, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleForceSyncOplogByMerkle(
	dataBytes []byte,
	peer *PttPeer,

	forceSyncOplogAckMsg OpType,
	forceSyncOplogByOplogAckMsg OpType,

	setDB func(oplog *BaseOplog),

	setNewestOplog func(log *BaseOplog) error,

	merkle *Merkle,
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

	data := &ForceSyncOplogByMerkle{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	merkleName := GetMerkleName(merkle, pm)

	nodes, err := merkle.GetChildNodes(data.Level, data.UpdateTS)
	log.Debug("HandleForceSyncOplogByMerkle: after GetChildKeys", "children", len(nodes), "they", len(data.Nodes), "level", data.Level, "ts", data.UpdateTS, "merkle", merkleName)
	if err != nil {
		return err
	}

	myNewNodes, theirNewNodes, err := DiffMerkleTree(nodes, data.Nodes, types.ZeroTimestamp, pm, merkle)
	log.Debug("HandleForceSyncOplogByMerkle: after DiffMerkleKeys", "myNewKeys", myNewNodes, "theirNewNodes", theirNewNodes, "level", data.Level, "ts", data.UpdateTS, "merkle", merkleName)
	if err != nil {
		return err
	}

	if len(theirNewNodes) == 0 {
		return nil
	}

	if data.Level == MerkleTreeLevelHR {
		return pm.ForceSyncOplogByOplogAck(
			theirNewNodes,
			forceSyncOplogByOplogAckMsg,
			setDB,
			setNewestOplog,
			peer,
			merkle,
		)
	}

	return pm.ForceSyncOplogByMerkleAck(
		theirNewNodes,
		forceSyncOplogAckMsg,
		merkle,
		peer,
	)
}
