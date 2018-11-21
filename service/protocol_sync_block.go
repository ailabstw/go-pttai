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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type SyncBlock struct {
	IDs []*SyncBlockID
}

func (pm *BaseProtocolManager) SyncBlock(op OpType, syncBlockIDs []*SyncBlockID, peer *PttPeer) error {
	if len(syncBlockIDs) == 0 {
		return nil
	}

	return pm.SendDataToPeer(op, &SyncBlock{IDs: syncBlockIDs}, peer)
}

func (pm *BaseProtocolManager) HandleSyncBlock(
	dataBytes []byte,
	peer *PttPeer,
	obj Object,
	syncAckMsg OpType,
) error {

	data := &SyncBlock{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	lenObjs := len(data.IDs)
	if lenObjs == 0 {
		return nil
	}

	blocks := make([]*Block, 0, lenObjs*NSubBlock)

	var blockInfo *BlockInfo
	var newBlocks []*Block
	for _, syncBlockID := range data.IDs {
		newObj, err := obj.GetNewObjByID(syncBlockID.ObjID, false)
		if err != nil {
			continue
		}

		if newObj.GetStatus() == types.StatusInternalSync {
			continue
		}

		blockInfo = newObj.GetBlockInfo()
		if blockInfo == nil || !reflect.DeepEqual(blockInfo.ID, syncBlockID.ID) {
			continue
		}
		pm.SetBlockInfoDB(blockInfo, syncBlockID.ObjID)

		newBlocks, err = GetBlockList(blockInfo, 0, false)
		if err != nil {
			continue
		}
		log.Debug("HandleSyncCreateBlock: (in-for-loop)", "blockInfo", blockInfo, "newBlocks", newBlocks)
		blocks = append(blocks, newBlocks...)
	}

	log.Debug("HandleSyncCreateBlock: to syncBlockAck", "blocks", blocks)

	return pm.SyncBlockAck(syncAckMsg, blocks, peer)
}

func blocksToBlocksByIDsByObjs(blocks []*Block) map[types.PttID]map[types.PttID][]*Block {
	nBlocksByIDsByObjs := make(map[types.PttID]map[types.PttID]int)

	blocksByIDsByObjs := make(map[types.PttID]map[types.PttID][]*Block)

	// count
	var block *Block
	var nBlocksByIDsByObj map[types.PttID]int
	ok := false

	for _, block = range blocks {
		nBlocksByIDsByObj, ok = nBlocksByIDsByObjs[*block.ObjID]
		if !ok {
			nBlocksByIDsByObj = make(map[types.PttID]int)
			nBlocksByIDsByObjs[*block.ObjID] = nBlocksByIDsByObj
		}
		_, ok = nBlocksByIDsByObj[*block.ID]
		if !ok {
			nBlocksByIDsByObj[*block.ID] = 0
		}
		nBlocksByIDsByObj[*block.ID]++
	}

	// allocate

	for id, nBlocksByIDsByObj := range nBlocksByIDsByObjs {
		blocksByIDsByObjs[id] = make(map[types.PttID][]*Block)
		for id2, nBlocksByIDByObj := range nBlocksByIDsByObj {
			blocksByIDsByObjs[id][id2] = make([]*Block, 0, nBlocksByIDByObj)
		}
	}

	// set
	var blocksByIDsByObj map[types.PttID][]*Block
	var blocksByIDByObj []*Block

	for _, block = range blocks {
		blocksByIDsByObj, ok = blocksByIDsByObjs[*block.ObjID]
		if !ok {
			continue
		}
		blocksByIDByObj, ok = blocksByIDsByObj[*block.ID]
		if !ok {
			continue
		}
		blocksByIDByObj = append(blocksByIDByObj, block)
		blocksByIDsByObj[*block.ID] = blocksByIDByObj
	}

	return blocksByIDsByObjs
}

func shrinkBlocks(blockInfo *BlockInfo, blocks []*Block) []*Block {
	newBlocks := make([]*Block, 0, len(blocks))
	var blockID uint32
	var subBlockID uint8

	for _, block := range blocks {
		blockID = block.BlockID
		subBlockID = block.SubBlockID
		if blockInfo.GetIsGood(blockID, subBlockID) {
			continue
		}

		newBlocks = append(newBlocks, block)
	}

	return newBlocks
}

func verifyBlocks(blocks []*Block, blockInfo *BlockInfo, creatorID *types.PttID) error {
	var err error

	for _, block := range blocks {
		err = block.Verify(blockInfo.Hashs[block.BlockID][block.SubBlockID], creatorID)
		log.Debug("verifyBlocks: after Verify", "blockID", block.BlockID, "subBlockID", block.SubBlockID, "e", err)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveBlocks(blocks []*Block, blockInfo *BlockInfo) bool {
	var err error
	var blockID uint32
	var subBlockID uint8

	isSet := false
	for _, block := range blocks {
		blockID = block.BlockID
		subBlockID = block.SubBlockID

		blockInfo.SetBlockDB(block)
		err = block.Save()
		if err != nil {
			continue
		}
		blockInfo.SetIsGood(blockID, subBlockID, true)
		isSet = true
	}

	return isSet
}
