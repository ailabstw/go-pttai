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
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

/*
GetBlockList gets the block list based on the information of block-info.

Always getting all blocks with pttdb.ListOrderNext
*/
func GetBlockList(blockInfo *BlockInfo, limit uint32, isLocked bool) ([]*Block, error) {
	if !blockInfo.GetIsAllGood() {
		return nil, ErrInvalidBlock
	}

	iter, err := blockInfo.GetBlockIterWithBlockInfo(isLocked)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	// for-loop
	var each *Block
	var v []byte

	uint32NBlock := uint32(blockInfo.NBlock)
	if limit == 0 {
		limit = uint32NBlock
	}
	if limit > uint32NBlock {
		limit = uint32NBlock
	}

	blocks := make([]*Block, 0, limit*NSubBlock)

	for iter.Next() {
		v = iter.Value()
		each = NewEmptyBlock()
		err = each.Unmarshal(v)
		if err != nil {
			continue
		}

		if each.BlockID >= limit {
			break
		}

		blocks = append(blocks, each)
	}

	return blocks, nil
}

/*
GetContentBlockList gets the block list based on the information of block-info.

Always getting all blocks with pttdb.ListOrderNext

This is used for the main-content (article).
*/
func GetContentBlockList(blockInfo *BlockInfo, limit uint32, isLocked bool) ([]*ContentBlock, error) {
	if !blockInfo.GetIsAllGood() {
		return nil, ErrInvalidBlock
	}

	iter, err := blockInfo.GetBlockIterWithBlockInfo(isLocked)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	// for-loop
	var each *Block
	var v []byte

	uint32NBlock := uint32(blockInfo.NBlock)
	if limit == 0 {
		limit = uint32NBlock
	}
	if limit > uint32NBlock {
		limit = uint32NBlock
	}

	bufs := make([][][]byte, limit)
	for i := 0; i < len(bufs); i++ {
		bufs[i] = make([][]byte, NSubBlock)
	}

	for iter.Next() {
		v = iter.Value()
		each = NewEmptyBlock()
		blockInfo.SetBlockDB(each)
		err = each.Unmarshal(v)
		if err != nil {
			continue
		}
		log.Debug("GetBlockList: (in-for-loop)", "blockID", each.BlockID, "subBlockID", each.SubBlockID, "limit", limit)

		if each.BlockID >= limit {
			break
		}

		bufs[each.BlockID][each.SubBlockID] = each.Buf

	}

	// check
	for i := 0; i < len(bufs); i++ {
		for j := 0; j < NSubBlock; j++ {
			if bufs[i][j] == nil {
				return nil, ErrInvalidBlock
			}
		}
	}

	// unscrambedBufs
	unscrambledBufs := make([][][]byte, len(bufs))
	var unscrambledBuf [][]byte
	for i, eachBuf := range bufs {
		unscrambledBuf, err = UnscrambleBuf(eachBuf)
		log.Debug("GetBlockList: (unscramble)", "i", i, "e", err)
		if err != nil {
			return nil, ErrInvalidBlock
		}
		unscrambledBufs[i] = unscrambledBuf
	}

	// content-blocks
	contentBlocks := make([]*ContentBlock, len(unscrambledBufs))
	for i, each := range unscrambledBufs {
		contentBlocks[i] = NewContentBlock(uint32(i), each)
	}

	return contentBlocks, nil
}

func (b *BlockInfo) GetBlockIterWithBlockInfo(isLocked bool) (iterator.Iterator, error) {
	block := NewEmptyBlock()
	b.SetBlockDB(block)

	return block.GetIter(pttdb.ListOrderNext, isLocked)
}
