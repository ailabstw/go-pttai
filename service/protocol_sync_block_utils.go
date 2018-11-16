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

import "github.com/ailabstw/go-pttai/common/types"

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
		if blockInfo.IsGood[blockID][subBlockID] {
			continue
		}

		newBlocks = append(newBlocks, block)
	}

	return newBlocks
}

func verifyBlocks(blocks []*Block, blockInfo *BlockInfo, creatorID *types.PttID) error {
	var err error
	var blockID uint32
	var subBlockID uint8

	for _, block := range blocks {
		err = block.Verify(blockInfo.Hashs[blockID][subBlockID], creatorID)
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
