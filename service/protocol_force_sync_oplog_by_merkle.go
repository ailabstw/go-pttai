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

type ForceSyncOplogByMerkle struct {
	UpdateTS types.Timestamp `json:"UT"`
	Level    MerkleTreeLevel `json:"L"`
	Keys     [][]byte        `json:"K"`
}

func (pm *BaseProtocolManager) ForceSyncOplogByMerkle(
	myNewNode *MerkleNode,

	forceSyncOplogMsg OpType,

	merkle *Merkle,

	peer *PttPeer,
) error {

	myNode, _ := merkle.GetNodeByLevelTS(myNewNode.Level, myNewNode.UpdateTS)

	log.Debug("ForceSyncOplogByMerkle: to GetChildKeys", "myNode", myNode, "myNewNode", myNewNode, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	keys, err := merkle.GetChildKeys(myNewNode.Level, myNewNode.UpdateTS)
	log.Debug("ForceSyncOplogByMerkle: after GetChildKeys", "e", err, "level", myNewNode.Level, "TS", myNewNode.UpdateTS, "keys", keys, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	if err != nil {
		return err
	}

	if myNode != nil && len(keys) != int(myNode.NChildren) {
		log.Debug("ForceSyncOplogByMerkle: len != NChildren", "len", len(keys), "children", myNode.NChildren, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
		return merkle.TryForceSync(pm)
	}

	if myNode == nil && len(keys) != 0 {
		log.Debug("ForceSyncOplogByMerkle: len != 0", "len", len(keys), "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
		return merkle.TryForceSync(pm)
	}

	data := &ForceSyncOplogByMerkle{
		UpdateTS: myNewNode.UpdateTS,
		Level:    myNewNode.Level,
		Keys:     keys,
	}

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

	log.Debug("HandleForceSyncOplogByMerkle: to GetChildKeys", "level", data.Level, "TS", data.UpdateTS, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	keys, err := merkle.GetChildKeys(data.Level, data.UpdateTS)
	log.Debug("HandleForceSyncOplogByMerkle: after GetChildKeys", "level", data.Level, "TS", data.UpdateTS, "keys", keys, "e", err, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	if err != nil {
		return err
	}

	myNewKeys, theirNewKeys, err := DiffMerkleKeys(keys, data.Keys)
	log.Debug("HandleForceSyncOplogByMerkle: after DiffMerkleKeys", "myNewKeys", myNewKeys, "theirNewKeys", theirNewKeys, "e", err, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	if err != nil {
		return err
	}

	if data.Level == MerkleTreeLevelHR {
		return pm.ForceSyncOplogByOplogAck(
			theirNewKeys,
			forceSyncOplogByOplogAckMsg,
			setDB,
			setNewestOplog,
			peer,
			merkle,
		)
	}

	return pm.ForceSyncOplogByMerkleAck(
		theirNewKeys,
		forceSyncOplogAckMsg,
		merkle,
		peer,
	)
}
