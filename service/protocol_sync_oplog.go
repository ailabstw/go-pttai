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

// sync-oplog
type SyncOplog struct {
	ToSyncTime  types.Timestamp `json:"LT"`
	ToSyncNodes []*MerkleNode   `json:"LN"`
}

func (pm *BaseProtocolManager) SyncOplog(peer *PttPeer, merkle *Merkle, op OpType) error {
	ptt := pm.Ptt()
	myInfo := ptt.GetMyEntity()
	if myInfo.GetStatus() != types.StatusAlive {
		log.Warn("SyncOplog: I am not alive", "status", myInfo.GetStatus())
		return nil
	}

	e := pm.Entity()
	if e.GetStatus() != types.StatusAlive {
		return nil
	}

	toSyncTime, err := merkle.ToSyncTime()
	log.Debug("SyncOplog: after GetSyncTime", "e", pm.Entity().GetID(), "toSyncTime", toSyncTime, "e", err)
	if err != nil {
		return err
	}

	toSyncNodes, _, err := merkle.GetMerkleTreeList(toSyncTime)
	log.Debug("SyncOplog: after GetMerkleTreeList", "e", pm.Entity().GetID(), "toSyncNodes", toSyncNodes)
	if err != nil {
		return err
	}

	syncOplog := &SyncOplog{
		ToSyncTime:  toSyncTime,
		ToSyncNodes: toSyncNodes,
	}

	log.Debug("SyncOplog: to SendDataToPeer", "e", pm.Entity().GetID(), "op", op, "toSyncTime", toSyncTime, "toSyncNodes", len(toSyncNodes))

	err = pm.SendDataToPeer(op, syncOplog, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleSyncOplog(
	dataBytes []byte,
	peer *PttPeer,
	merkle *Merkle,
	op OpType,
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

	myToSyncTime, err := merkle.ToSyncTime()
	if err != nil {
		return err
	}

	toSyncTime := myToSyncTime
	if data.ToSyncTime.IsLess(toSyncTime) {
		toSyncTime = data.ToSyncTime
	}

	log.Debug("HandleSyncOplog: to GetMerkleTreeList", "op", op, "myToSyncTime", myToSyncTime, "toSyncTime", toSyncTime, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	myToSyncNodes, _, err := merkle.GetMerkleTreeList(toSyncTime)
	log.Debug("HandleSyncOplog: GetMerkleTreeList", "op", op, "err", err, "myToSyncNodes", len(myToSyncNodes), "entity", pm.Entity().GetID())
	if err != nil {
		return err
	}

	isValid := ValidateMerkleTree(myToSyncNodes, data.ToSyncNodes, toSyncTime)
	log.Debug("HandleSyncOplog: after ValidateMerkleTree", "op", op, "isValid", isValid)
	if !isValid {
		return ErrInvalidOplog
	}

	return pm.SyncOplogAck(toSyncTime, merkle, op, peer)
}
