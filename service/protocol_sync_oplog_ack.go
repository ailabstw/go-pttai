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
}

func (pm *BaseProtocolManager) SyncOplogAck(toSyncTime types.Timestamp, merkle *Merkle, op OpType, peer *PttPeer) error {
	now, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	offsetHourTS, _ := toSyncTime.ToHRTimestamp()
	nodes, err := merkle.GetMerkleTreeListByLevel(MerkleTreeLevelNow, offsetHourTS, now)
	log.Debug("SyncOplogAck: after GetMerkleTreeListByLevel", "offsetHourTS", offsetHourTS, "nodes", nodes, "e", err, "entity", pm.Entity().GetID())
	if err != nil {
		return err
	}

	syncOplogAck := &SyncOplogAck{
		TS:    offsetHourTS,
		Nodes: nodes,
	}

	log.Debug("SyncOplogAck: to SendDataToPeer", "nodes", nodes, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	err = pm.SendDataToPeer(op, syncOplogAck, peer)
	if err != nil {
		return err
	}

	return nil
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

	now, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	myNodes, err := merkle.GetMerkleTreeListByLevel(MerkleTreeLevelNow, data.TS, now)
	if err != nil {
		return err
	}

	myNewKeys, theirNewKeys, err := MergeMerkleNodeKeys(myNodes, data.Nodes)
	if err != nil {
		return err
	}

	log.Debug("HandleSyncOplogAck: to SyncOplogNewOplogs", "myNodes", myNodes, "myNewKeys", len(myNewKeys), "theirNewKeys", len(theirNewKeys), "newLogsMsg", newLogsMsg, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	return pm.SyncOplogNewOplogs(data, myNewKeys, theirNewKeys, peer, setDB, setNewestOplog, postsync, newLogsMsg)
}
