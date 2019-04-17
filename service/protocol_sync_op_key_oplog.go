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
	"github.com/ailabstw/go-pttai/pttdb"
)

type SyncOpKeyOplog struct {
	Oplogs []*BaseOplog `json:"O"`
}

/**********
 * OpKeyOplog
 **********/

func (pm *BaseProtocolManager) SyncOpKeyOplog(peer *PttPeer, syncMsg OpType) error {
	oplogs, failedOplogs, err := pm.getOpKeyOplogs()
	log.Debug("SyncOpKeyOplog: after getOpKeyOplogs", "oplogs", oplogs, "failedOplogs", failedOplogs, "e", err, "entity", pm.Entity().IDString())
	err = HandleFailedOplogs(failedOplogs, pm.SetOpKeyDB, pm.HandleFailedOpKeyOplog)
	log.Debug("SyncOpKeyOplog: after HandleFailedOplogs", "e", err, "entity", pm.Entity().IDString())
	if err != nil {
		return err
	}

	if peer == nil {
		return nil
	}

	if !peer.IsRegistered {
		return nil
	}

	data := &SyncOpKeyOplog{
		Oplogs: oplogs,
	}

	err = pm.SendDataToPeer(syncMsg, data, peer)
	log.Debug("SyncOpKeyOplog: after send data", "e", err, "entity", pm.Entity().GetID())

	return nil
}

func (pm *BaseProtocolManager) HandleSyncOpKeyOplog(dataBytes []byte, peer *PttPeer, syncMsg OpType) error {
	data := &SyncOpKeyOplog{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	log.Debug("HandleSyncOpKeyOplog: to HandleOpKeyOplogs", "oplogs", data.Oplogs, "entity", pm.Entity().IDString())
	err = pm.HandleOpKeyOplogs(data.Oplogs, peer, false)
	if err != nil {
		return err
	}

	switch syncMsg {
	case SyncOpKeyOplogMsg:
		err = pm.SyncOpKeyOplog(peer, SyncOpKeyOplogAckMsg)
	case SyncOpKeyOplogAckMsg:
		err = pm.postsyncOpKeyOplogs(peer)
	}

	return err
}

func (pm *BaseProtocolManager) getOpKeyOplogs() ([]*BaseOplog, []*BaseOplog, error) {
	oplog := &BaseOplog{}
	pm.SetOpKeyDB(oplog)

	expireTS, err := pm.getExpireOpKeyTS()
	if err != nil {
		return nil, nil, err
	}

	oplogs, err := GetOplogList(oplog, nil, 0, pttdb.ListOrderNext, types.StatusAlive, false)
	if err != nil {
		return nil, nil, err
	}

	lenLogs := len(oplogs)
	logs := make([]*BaseOplog, 0, lenLogs)
	failedLogs := make([]*BaseOplog, 0, lenLogs)

	for _, log := range oplogs {
		if log.UpdateTS.IsLess(expireTS) {
			failedLogs = append(failedLogs, log)
		} else {
			logs = append(logs, log)
		}
	}

	return logs, failedLogs, nil
}

func (pm *BaseProtocolManager) SyncPendingOpKeyOplog(peer *PttPeer) error {
	return pm.SyncPendingOplog(peer, pm.SetOpKeyDB, pm.HandleFailedOpKeyOplog, SyncPendingOpKeyOplogMsg)
}
