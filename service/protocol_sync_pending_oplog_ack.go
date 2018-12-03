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

type SyncPendingOplogAck struct {
	Oplogs []*BaseOplog
}

func (pm *BaseProtocolManager) SyncPendingOplogAck(
	data *SyncPendingOplog,
	peer *PttPeer,

	syncAckMsg OpType,

	handlePendingOplogs func(oplogs []*BaseOplog, peer *PttPeer) error,
	setDB func(oplog *BaseOplog),
	handleFailedOplog func(oplog *BaseOplog) error,
) error {

	err := handlePendingOplogs(data.Oplogs, peer)
	if err != nil {
		return err
	}

	oplogs, failedOplogs, err := pm.GetPendingOplogs(setDB, peer, false)
	if err != nil {
		return nil
	}

	err = HandleFailedOplogs(failedOplogs, setDB, handleFailedOplog)
	if err != nil {
		return nil
	}

	opData := &SyncPendingOplogAck{
		Oplogs: oplogs,
	}

	pm.SendDataToPeer(syncAckMsg, opData, peer)

	return nil
}

func (pm *BaseProtocolManager) HandleSyncPendingOplogAck(
	dataBytes []byte,
	peer *PttPeer,

	handlePendingOplogs func(oplogs []*BaseOplog, peer *PttPeer) error,
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

	data := &SyncPendingOplogAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	err = handlePendingOplogs(data.Oplogs, peer)
	if err != nil {
		return err
	}

	return nil
}
