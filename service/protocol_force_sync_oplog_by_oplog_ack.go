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

type ForceSyncOplogByOplogAck struct {
	Oplogs []*BaseOplog `json:"O"`
}

func (pm *BaseProtocolManager) ForceSyncOplogByOplogAck(
	theirNewKeys [][]byte,

	forceSyncOplogByOplogAckMsg OpType,

	setDB func(oplog *BaseOplog),
	setNewestOplog func(log *BaseOplog) error,

	peer *PttPeer,

	merkle *Merkle,
) error {

	theirNewLogs, err := getOplogsFromKeys(setDB, theirNewKeys)
	log.Debug("ForceSyncOplogByOplogAck: after get theirNewLogs", "theirNewLogs", len(theirNewLogs), "e", err, "merkle", merkle.Name)
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

	log.Debug("HandleForceSyncOplogByOplogAck: to handleOplogs", "oplogs", data.Oplogs, "merkle", merkle.Name)

	err = handleOplogs(data.Oplogs, peer, true)
	if err != nil {
		return err
	}

	return nil
}
