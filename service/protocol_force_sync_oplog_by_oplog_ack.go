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

type ForceSyncOplogByOplogAck struct {
	Oplogs []*BaseOplog `json:"O"`
}

func (pm *BaseProtocolManager) ForceSyncOplogByOplogAck(
	theirNewNodes []*MerkleNode,

	forceSyncOplogByOplogAckMsg OpType,

	setDB func(oplog *BaseOplog),
	setNewestOplog func(log *BaseOplog) error,

	peer *PttPeer,

	merkle *Merkle,
) error {

	keys := make([][]byte, 0, len(theirNewNodes))
	for _, node := range theirNewNodes {
		keys = append(keys, node.Key)
	}

	merkleName := GetMerkleName(merkle, pm)

	theirNewLogs, err := getOplogsFromKeys(setDB, keys)
	log.Debug("ForceSyncOplogByOplogAck: after getOplogsFromKeys", "keys", len(keys), "theirNewLogs", theirNewLogs, "e", err, "merkle", merkleName, "peer", peer)
	if err != nil {
		return err
	}

	if len(theirNewLogs) == 0 {
		return nil
	}

	if setNewestOplog != nil {
		for _, log := range theirNewLogs {
			setNewestOplog(log)
		}
	}

	data := &SyncOplogNewOplogs{
		Oplogs: theirNewLogs,
	}

	err = pm.SendDataToPeer(forceSyncOplogByOplogAckMsg, data, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleForceSyncOplogByOplogAck(
	dataBytes []byte,
	peer *PttPeer,

	handleOplogs func(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error,

	merkle *Merkle,

) error {

	ptt := pm.Ptt()
	myInfo := ptt.GetMyEntity()
	if myInfo.GetStatus() != types.StatusAlive {
		return nil
	}

	entity := pm.Entity()
	if entity.GetStatus() != types.StatusAlive {
		return nil
	}

	data := &ForceSyncOplogByOplogAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	merkleName := GetMerkleName(merkle, pm)
	log.Debug("HandleForceSyncOplogByOplogAck: to handleOplogs", "oplogs", data.Oplogs, "merkle", merkleName, "peer", peer)

	err = handleOplogs(data.Oplogs, peer, true)
	if err != nil {
		return err
	}

	return nil
}
