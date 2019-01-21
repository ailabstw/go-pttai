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

func (pm *BaseProtocolManager) ForceSyncOplogAck(
	fromTS types.Timestamp,
	toTS types.Timestamp,

	merkle *Merkle,

	forceSyncOplogAckMsg OpType,

	peer *PttPeer,
) error {

	nodes := make([]*MerkleNode, 0, MaxSyncOplogAck)

	var err error
	offsetHourTS, _ := fromTS.ToHRTimestamp()
	startHourTS := offsetHourTS
	nextHourTS := offsetHourTS

	for currentHourTS := offsetHourTS; currentHourTS.IsLess(toTS); currentHourTS = nextHourTS {
		nodes, startHourTS, nextHourTS, err = pm.syncOplogAckCore(
			merkle,
			peer,
			forceSyncOplogAckMsg,
			nodes,
			currentHourTS,
			offsetHourTS,
			startHourTS,
			toTS,
		)
		log.Debug("ForceSyncOplogAck: (in-for-loop) after syncOplogAckCore", "nodes", nodes, "currentHour", currentHourTS, "nextHour", nextHourTS, "e", err, "entity", pm.Entity().GetID())
		if err != nil {
			return err
		}
	}

	log.Debug("ForceSyncOplogAck: after for-loop", "nodes", nodes, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	// deal with the last part of the nodes.
	if len(nodes) == 0 {
		return nil
	}

	syncOplogAck := &SyncOplogAck{
		TS:          offsetHourTS,
		Nodes:       nodes,
		StartHourTS: startHourTS,
		EndHourTS:   toTS,
		StartTS:     startHourTS,
		EndTS:       toTS,
	}

	err = pm.SendDataToPeer(forceSyncOplogAckMsg, syncOplogAck, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleForceSyncOplogAck(
	dataBytes []byte,
	peer *PttPeer,

	merkle *Merkle,
	info ProcessInfo,

	setDB func(oplog *BaseOplog),
	handleFailedValidOplog func(oplog *BaseOplog, info ProcessInfo) error,
	setNewestOplog func(log *BaseOplog) error,
	postprocessLogs func(info ProcessInfo, peer *PttPeer) error,

	newLogsMsg OpType,
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

	data := &SyncOplogAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	myNodes, err := merkle.GetMerkleTreeListByLevel(MerkleTreeLevelNow, data.StartHourTS, data.EndHourTS)
	if err != nil {
		return err
	}

	myNodes = pm.handleSyncOplogAckFilterTS(myNodes, data.StartTS, data.EndTS)

	myNewKeys, theirNewKeys, err := MergeMerkleNodeKeys(myNodes, data.Nodes)
	if err != nil {
		return err
	}

	theirNewLogs, err := getOplogsFromKeys(setDB, theirNewKeys)
	if err != nil {
		return err
	}

	err = HandleFailedValidOplogs(
		theirNewLogs,
		peer,

		info,

		setDB,
		handleFailedValidOplog,
		postprocessLogs,
	)
	if err != nil {
		return err
	}

	emptyKeys := make([][]byte, 0)
	return pm.SyncOplogNewOplogs(
		data,
		myNewKeys,
		emptyKeys,
		peer,

		nil,
		nil,

		setDB,
		setNewestOplog,
		nil,
		newLogsMsg,
	)
}
