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

type SyncBlockAck struct {
	Blocks []*Block `json:"B"`
}

func (pm *BaseProtocolManager) SyncBlockAck(ackMsg OpType, blocks []*Block, peer *PttPeer) error {
	if len(blocks) == 0 {
		return nil
	}

	return pm.SendDataToPeer(ackMsg, &SyncBlockAck{Blocks: blocks}, peer)
}

func (pm *BaseProtocolManager) HandleSyncCreateBlockAck(
	dataBytes []byte,
	peer *PttPeer,

	obj Object,

	setLogDB func(oplog *BaseOplog),
	postcreate func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
) error {

	data := &SyncBlockAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	blocks := data.Blocks

	if len(blocks) == 0 {
		return nil
	}

	blocksByIDsByObjs := blocksToBlocksByIDsByObjs(blocks)

	for objID, blocksByIDsByObj := range blocksByIDsByObjs {
		err = pm.handleSyncCreateBlockAck(blocksByIDsByObj, peer, &objID, obj, setLogDB, postcreate, broadcastLog)
		if err != nil {
			break
		}
	}

	return err
}

func (pm *BaseProtocolManager) handleSyncCreateBlockAck(
	blocksByIDsByObj map[types.PttID][]*Block,
	peer *PttPeer,

	objID *types.PttID,
	origObj Object,

	setLogDB func(oplog *BaseOplog),
	postcreate func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
) error {

	var err error

	log.Debug("HandleSyncCreateBlockAck: start", "obj", objID)

	// orig-obj
	origObj.SetID(objID)
	err = origObj.Lock()
	if err != nil {
		return err
	}
	defer origObj.Unlock()

	err = origObj.GetByID(true)
	if err != nil {
		return err
	}

	// validate obj
	log.Debug("HandleSyncCreateBlockAck: to GetIsAllGood", "obj", objID)
	if origObj.GetIsAllGood() {
		return nil
	}

	log.Debug("HandleSyncCreateBlockAck: to get blockInfo", "obj", objID)
	blockInfo := origObj.GetBlockInfo()
	if blockInfo == nil {
		return nil
	}

	pm.SetBlockInfoDB(blockInfo, objID)

	log.Debug("HandleSyncCreateBlockAck: to get blocks", "obj", objID)
	blocks, ok := blocksByIDsByObj[*blockInfo.ID]
	if !ok {
		return nil
	}

	// shrink
	log.Debug("HandleSyncCreateBlockAck: to shrink", "obj", objID)
	blocks = shrinkBlocks(blockInfo, blocks)
	if len(blocks) == 0 {
		return nil
	}

	// verify
	log.Debug("HandleSyncCreateBlockAck: to verify", "obj", objID)
	creatorID := origObj.GetCreatorID()
	err = verifyBlocks(blocks, blockInfo, creatorID)
	if err != nil {
		return err
	}

	// set
	log.Debug("HandleSyncCreateBlockAck: to set", "obj", objID)
	isSet := saveBlocks(blocks, blockInfo)
	if !isSet {
		return nil
	}

	isAllGood := origObj.CheckIsAllGood()
	log.Debug("HandleSyncCreateBlockAck: after CheckIsAllGood", "obj", objID, "isAllGood", isAllGood)
	if !isAllGood {
		return origObj.Save(true)
	}

	// get oplog
	logID := origObj.GetLogID()
	oplog := &BaseOplog{ID: logID}
	setLogDB(oplog)
	err = oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	err = oplog.Get(logID, true)
	if err != nil {
		return err
	}

	err = pm.saveNewObjectWithOplog(origObj, oplog, true, false, postcreate)
	if err != nil {
		return err
	}

	return pm.syncCreateAckSaveOplog(oplog, origObj, broadcastLog, postcreate)
}
