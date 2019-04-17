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

func (pm *BaseProtocolManager) HandleSyncUpdateBlockAck(
	dataBytes []byte,
	peer *PttPeer,

	obj Object,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,
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
		err = pm.handleSyncUpdateBlockAck(
			blocksByIDsByObj,
			peer,

			&objID,
			obj,

			merkle,

			setLogDB,
			postupdate,
			broadcastLog,
		)
		if err != nil {
			break
		}
	}

	return err
}

func (pm *BaseProtocolManager) handleSyncUpdateBlockAck(
	blocksByIDsByObj map[types.PttID][]*Block,
	peer *PttPeer,

	objID *types.PttID,
	origObj Object,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

	postupdate func(obj Object, oplog *BaseOplog) error,
	broadcastLog func(oplog *BaseOplog) error,
) error {

	var err error

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
	syncInfo := origObj.GetSyncInfo()
	if syncInfo == nil {
		return nil
	}

	if syncInfo.GetIsAllGood() {
		return nil
	}

	blockInfo := syncInfo.GetBlockInfo()
	if blockInfo == nil {
		return nil
	}

	pm.SetBlockInfoDB(blockInfo, objID)

	blocks, ok := blocksByIDsByObj[*blockInfo.ID]
	if !ok {
		return nil
	}

	// shrink
	blocks = shrinkBlocks(blockInfo, blocks)
	if len(blocks) == 0 {
		return nil
	}

	// verify

	creatorID := syncInfo.GetUpdaterID()
	err = verifyBlocks(blocks, blockInfo, creatorID)
	if err != nil {
		return err
	}

	// set
	isSet := saveBlocks(blocks, blockInfo)
	if !isSet {
		return nil
	}

	isAllGood := syncInfo.CheckIsAllGood()
	if !isAllGood {
		return origObj.Save(true)
	}

	// get oplog
	logID := syncInfo.GetLogID()
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

	err = pm.handleUpdateObjectSameLog(origObj, syncInfo, oplog, postupdate)
	if err != nil {
		return err
	}

	return pm.syncUpdateAckSaveOplog(
		oplog,
		syncInfo,
		origObj,

		merkle,

		broadcastLog,
		postupdate,
	)
}
