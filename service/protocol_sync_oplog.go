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
)

// sync-oplog
type SyncOplog struct {
	LastSyncTime  types.Timestamp `json:"LT"`
	LastSyncNodes []*MerkleNode   `json:"LN"`
}

func (pm *BaseProtocolManager) SyncOplog(peer *PttPeer, merkle *Merkle, op OpType) error {
	ptt := pm.Ptt()
	myInfo := ptt.GetMyEntity()
	if myInfo.GetStatus() != types.StatusAlive {
		return nil
	}

	e := pm.Entity()
	if e.GetStatus() != types.StatusAlive {
		return nil
	}

	lastSyncTime, err := merkle.GetSyncTime()
	if err != nil {
		return err
	}

	lastSyncNodes, _, err := merkle.GetMerkleTreeList(lastSyncTime)
	if err != nil {
		return err
	}

	syncOplog := &SyncOplog{
		LastSyncTime:  lastSyncTime,
		LastSyncNodes: lastSyncNodes,
	}

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

	myLastSyncTime, err := merkle.GetSyncTime()
	if err != nil {
		return err
	}

	lastSyncTime := myLastSyncTime
	if data.LastSyncTime.IsLess(lastSyncTime) {
		lastSyncTime = data.LastSyncTime
	}

	myLastSyncNodes, _, err := merkle.GetMerkleTreeList(lastSyncTime)
	if err != nil {
		return err
	}

	isValid := ValidateMerkleTree(myLastSyncNodes, data.LastSyncNodes, lastSyncTime)
	if !isValid {
		return ErrInvalidOplog
	}

	return pm.SyncOplogAck(lastSyncTime, merkle, op, peer)
}
