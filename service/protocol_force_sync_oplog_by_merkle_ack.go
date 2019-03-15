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

type ForceSyncOplogByMerkleAck struct {
	Keys [][]byte `json:"K"`
}

func (pm *BaseProtocolManager) ForceSyncOplogByMerkleAck(
	keys [][]byte,

	forceSyncOplogAckMsg OpType,
	merkle *Merkle,

	peer *PttPeer,

) error {

	data := &ForceSyncOplogByMerkleAck{
		Keys: keys,
	}

	err := pm.SendDataToPeer(forceSyncOplogAckMsg, data, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleForceSyncOplogByMerkleAck(
	dataBytes []byte,
	peer *PttPeer,

	forceSyncOplogMsg OpType,

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

	data := &ForceSyncOplogByMerkleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	log.Debug("HandleForceSyncOplogByMerkleAck: to for-loop", "keys", data.Keys, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	node := &MerkleNode{}
	for _, key := range data.Keys {
		node.ConstructUpdateTSAndLevelByKey(key)
		log.Debug("HandleForceSyncOplogByMerkleAck: (in-for-loop)", "node", node, "merkle", merkle.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
		pm.ForceSyncOplogByMerkle(node, forceSyncOplogMsg, merkle, peer)
	}

	return nil
}
