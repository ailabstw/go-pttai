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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type BlockInfo struct {
	V types.Version

	ID *types.PttID `json:"ID,omitempty"`

	NBlock    int                   `json:"N"`
	Hashs     [][][]byte            `json:"H,omitempty"`
	IsGood    types.BoolDoubleArray `json:"G,omitempty"`
	IsAllGood types.Bool            `json:"g"`

	MediaIDs []*types.PttID `json:"M,omitempty"`

	UpdaterID *types.PttID `json:"U"`

	db           *pttdb.LDBBatch
	dbLock       *types.LockMap
	fullDBPrefix []byte

	objID *types.PttID

	setMediaDB func(media *Media)
}

func NewBlockInfo(

	id *types.PttID,

	hashs [][][]byte,

	mediaIDs []*types.PttID,

	updaterID *types.PttID,

) (*BlockInfo, error) {

	nBlock := len(hashs)

	isGoodAry := make([][]types.Bool, nBlock)
	for i := 0; i < nBlock; i++ {
		isGoodAry[i] = make([]types.Bool, NSubBlock)
	}

	return &BlockInfo{
		V: types.CurrentVersion,

		ID: id,

		NBlock: nBlock,
		Hashs:  hashs,
		IsGood: types.BoolDoubleArray(isGoodAry),

		MediaIDs: mediaIDs,

		UpdaterID: updaterID,
	}, nil
}

func (pm *BaseProtocolManager) SetBlockInfoDB(blockInfo *BlockInfo, objID *types.PttID) {
	blockInfo.SetDB(pm.DB(), pm.DBObjLock(), pm.dbBlockPrefix, objID, pm.SetMediaDB)
}

func (b *BlockInfo) Init(nBlock int) {
	b.NBlock = nBlock

	b.Hashs = make([][][]byte, nBlock)
	for i := 0; i < nBlock; i++ {
		b.Hashs[i] = make([][]byte, 2)
	}

	b.ResetIsGood()
}

func (b *BlockInfo) SetDB(
	db *pttdb.LDBBatch,
	dbLock *types.LockMap,
	fullDBPrefix []byte,

	objID *types.PttID,

	setMediaDB func(media *Media),
) {
	b.V = types.CurrentVersion
	b.db = db
	b.dbLock = dbLock
	b.fullDBPrefix = fullDBPrefix
	b.objID = objID
	b.setMediaDB = setMediaDB
}

func (b *BlockInfo) SetBlockDB(block *Block) {
	block.SetDB(b.db, b.fullDBPrefix, b.objID, b.ID)
}

func (b *BlockInfo) VerifyBlocks(blocks []*Block) ([]*Block, []*Block) {
	goodBlocks := make([]*Block, len(blocks))
	badBlocks := make([]*Block, len(blocks))

	isGood := false
	for _, block := range blocks {
		isGood = b.VerifyBlock(block)
		if isGood {
			goodBlocks = append(goodBlocks, block)
		} else {
			badBlocks = append(badBlocks, block)
		}
	}

	return goodBlocks, badBlocks
}

func (b *BlockInfo) VerifyBlock(block *Block) bool {
	if int(block.BlockID) >= b.NBlock {
		return false
	}

	if block.SubBlockID >= NSubBlock {
		return false
	}

	marshaled, err := block.Marshal()
	if err != nil {
		return false
	}

	bytesWithSalt := append(marshaled, block.Salt[:]...)

	err = VerifyData(bytesWithSalt, block.Hash, block.Sig, block.Pub, b.UpdaterID, block.KeyExtra)
	if err != nil {
		return false
	}

	return true
}

func (b *BlockInfo) SaveBlocks(isLocked bool, blocks []*Block) error {
	var err error

	if !isLocked {
		err = b.dbLock.Lock(b.ID)
		if err != nil {
			return err
		}
		defer b.dbLock.Unlock(b.ID)
	}

	if b.IsGood == nil {
		b.InitIsGood()
	}
	for _, block := range blocks {
		if b.IsGood[block.BlockID][block.SubBlockID] {
			continue
		}

		if block.db == nil {
			b.SetBlockDB(block)
		}
		err = block.Save()
		if err != nil {
			break
		}
		b.IsGood[block.BlockID][block.SubBlockID] = true
	}

	return err
}

func (b *BlockInfo) Remove(
	isLocked bool,

) error {

	if !isLocked {
		err := b.dbLock.Lock(b.ID)
		if err != nil {
			return err
		}
		defer b.dbLock.Unlock(b.ID)
	}

	block := NewEmptyBlock()
	b.SetBlockDB(block)
	block.ID = b.ID

	err := block.RemoveAll()
	if err != nil {
		return err
	}

	if b.MediaIDs == nil {
		return nil
	}

	media := NewEmptyMedia()
	b.setMediaDB(media)

	for _, mediaID := range b.MediaIDs {
		media.SetID(mediaID)
		media.DeleteAll(false)
	}

	return nil
}

func (b *BlockInfo) GetIsAllGood() bool {
	if b.NBlock == 0 {
		return true
	}

	if b.IsAllGood {
		return true
	}

	if b.IsGood == nil {
		return false
	}

	for i := 0; i < b.NBlock; i++ {
		for j := 0; j < NSubBlock; j++ {
			if !b.IsGood[i][j] {
				return false
			}
		}
	}

	b.IsAllGood = true
	return true
}

func (b *BlockInfo) ResetIsGood() {
	b.IsGood = nil
	b.IsAllGood = false
}

func (b *BlockInfo) GetIsGood(blockID uint32, subBlockID uint8) types.Bool {
	if b.IsGood == nil {
		return false
	}

	if blockID >= uint32(b.NBlock) {
		return false
	}

	if subBlockID >= NSubBlock {
		return false
	}

	return b.IsGood[blockID][subBlockID]
}

func (b *BlockInfo) SetIsGood(blockID uint32, subBlockID uint8, isGood types.Bool) {
	if b.IsGood == nil {
		b.InitIsGood()
	}

	b.IsGood[blockID][subBlockID] = isGood
}

func (b *BlockInfo) InitIsGood() {
	isGood := make([][]types.Bool, b.NBlock)

	for i := 0; i < b.NBlock; i++ {
		isGood[i] = make([]types.Bool, NSubBlock)
	}
	b.IsGood = types.BoolDoubleArray(isGood)
	b.IsAllGood = false
}

func (b *BlockInfo) SetIsAllGood() {
	if b.IsGood == nil {
		b.InitIsGood()
	}

	pIsGood := b.IsGood[:]
	var ppIsGood []types.Bool
	for i := 0; i < b.NBlock; i++ {
		ppIsGood = pIsGood[0][:]
		for j := 0; j < NSubBlock; j++ {
			ppIsGood[0] = true
			ppIsGood = ppIsGood[1:]
		}
		pIsGood = pIsGood[1:]
	}
	b.IsAllGood = true
}
