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
)

type SyncOplogNewOplogsAck struct {
	TS     types.Timestamp
	Oplogs []*BaseOplog `json:"O"`
}

func (pm *BaseProtocolManager) SyncOplogNewOplogsAck(
	ts types.Timestamp,
	keys [][]byte,
	peer *PttPeer,
	setDB func(oplog *BaseOplog),
	setNewestOplog func(log *BaseOplog) error,
	newLogsAckMsg OpType,
) error {

	if len(keys) == 0 {
		return nil
	}

	theirNewLogs, err := getOplogsFromKeys(setDB, keys)
	if err != nil {
		return err
	}

	if setNewestOplog != nil {
		for _, log := range theirNewLogs {
			setNewestOplog(log)
		}
	}

	data := &SyncOplogNewOplogsAck{
		TS:     ts,
		Oplogs: theirNewLogs,
	}

	err = pm.SendDataToPeer(newLogsAckMsg, data, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleSyncOplogNewOplogsAck(
	dataBytes []byte,
	peer *PttPeer,
	setDB func(oplog *BaseOplog),
	handleOplogs func(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error,
	postsync func(peer *PttPeer) error,
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

	data := &SyncOplogNewOplogsAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	err = handleOplogs(data.Oplogs, peer, true)
	if err != nil {
		return err
	}

	if postsync != nil {
		return postsync(peer)
	}

	return nil
}
