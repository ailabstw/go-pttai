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

type SyncPendingOplog struct {
	Oplogs []*BaseOplog
}

func (pm *BaseProtocolManager) SyncPendingOplog(
	peer *PttPeer,

	setDB func(oplog *BaseOplog),
	handleFailedOplog func(oplog *BaseOplog) error,

	syncMsg OpType,
) error {

	oplogs, failedOplogs, err := pm.GetPendingOplogs(setDB, peer, false)
	if err != nil {
		return nil
	}

	err = HandleFailedOplogs(failedOplogs, setDB, handleFailedOplog)
	if err != nil {
		return nil
	}

	if peer == nil {
		return nil
	}

	data := &SyncPendingOplog{
		Oplogs: oplogs,
	}

	pm.SendDataToPeer(syncMsg, data, peer)

	return nil
}

func (pm *BaseProtocolManager) HandleSyncPendingOplog(
	dataBytes []byte,
	peer *PttPeer,
	handlePendingOplogs func(oplogs []*BaseOplog, peer *PttPeer) error,
	setDB func(oplog *BaseOplog),
	handleFailedOplog func(oplog *BaseOplog) error,
	syncAckMsg OpType,
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

	data := &SyncPendingOplog{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return pm.SyncPendingOplogAck(data, peer, syncAckMsg, handlePendingOplogs, setDB, handleFailedOplog)
}
