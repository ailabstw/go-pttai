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

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/klauspost/reedsolomon"
)

func (pm *BaseProtocolManager) FullBlockDBPrefix(prefix []byte) ([]byte, error) {
	if prefix == nil {
		return pm.dbBlockPrefix, nil
	}

	return common.Concat([][]byte{pm.dbBlockPrefix, prefix})
}

func (pm *BaseProtocolManager) SplitContentBlocks(prefix []byte, objID *types.PttID, buf [][]byte, nFirstLineInBlock int) (*types.PttID, [][][]byte, error) {

	myEntity := pm.Ptt().GetMyEntity()

	blockInfoID, err := types.NewPttID()
	if err != nil {
		return nil, nil, err
	}

	hashs := make([][][]byte, 0)

	fullDBPrefix, err := pm.FullBlockDBPrefix(prefix)
	if err != nil {
		return nil, nil, err
	}

	nLineInBlock := nFirstLineInBlock
	lenCurrentBuf := 0
	var eachBlock *Block
	var eachHashs [][]byte
	for blockID, currentBuf := 0, buf[0:]; len(currentBuf) != 0; blockID, currentBuf = blockID+1, currentBuf[lenCurrentBuf:] {
		lenCurrentBuf = common.MinInt(nLineInBlock, len(currentBuf))
		scrambledBufs, err := ScrambleBuf(currentBuf[:lenCurrentBuf])
		if err != nil {
			return nil, nil, err
		}

		eachHashs = make([][]byte, len(scrambledBufs))
		for subBlockID, scrambledBuf := range scrambledBufs {
			eachBlock, err = NewBlock(uint32(blockID), uint8(subBlockID), scrambledBuf)
			if err != nil {
				return nil, nil, err
			}
			eachBlock.SetDB(pm.DB(), fullDBPrefix, objID, blockInfoID)
			err = myEntity.SignBlock(eachBlock)
			if err != nil {
				return nil, nil, err
			}

			err = eachBlock.Save()
			if err != nil {
				return nil, nil, err
			}
			eachHashs[subBlockID] = eachBlock.Hash
		}

		hashs = append(hashs, eachHashs)

		nLineInBlock = NLineInBlock
	}

	return blockInfoID, hashs, nil
}

func (pm *BaseProtocolManager) SplitMediaBlocks(objID *types.PttID, buf []byte) (*types.PttID, [][][]byte, error) {

	myEntity := pm.Ptt().GetMyEntity()

	blockInfoID, err := types.NewPttID()
	if err != nil {
		return nil, nil, err
	}

	hashs := make([][][]byte, 0)

	fullDBPrefix, err := pm.FullBlockDBPrefix(nil)
	if err != nil {
		return nil, nil, err
	}

	lenCurrentBuf := 0
	halfLenCurrentBuf := 0
	realLenCurrentBuf := 0
	var eachBlock *Block
	var bufs [][]byte
	var eachHashs [][]byte
	var firstHalfBuf []byte
	var secondHalfBuf []byte
	var scrambledBufs [][]byte

	for blockID, currentBuf := 0, buf[0:]; len(currentBuf) != 0; blockID, currentBuf = blockID+1, currentBuf[lenCurrentBuf:] {
		// 1. Unless there is only 1 char, we hope that both blocks contains at least 1 char. Squeezing the last-char to the block.
		realLenCurrentBuf = len(currentBuf)
		if realLenCurrentBuf == NByteInBlock+1 {
			lenCurrentBuf = realLenCurrentBuf
		} else {
			lenCurrentBuf = common.MinInt(NByteInBlock, len(currentBuf))
		}

		// 2. construct the bufs
		halfLenCurrentBuf = (lenCurrentBuf + 1) / 2
		firstHalfBuf = currentBuf[:halfLenCurrentBuf]
		secondHalfBuf = currentBuf[halfLenCurrentBuf:lenCurrentBuf]

		bufs = [][]byte{firstHalfBuf, secondHalfBuf}

		// 3. scramble the buf
		scrambledBufs, err = ScrambleBuf(bufs)
		if err != nil {
			return nil, nil, err
		}

		// 4. construct the hash
		eachHashs = make([][]byte, len(scrambledBufs))
		for subBlockID, scrambledBuf := range scrambledBufs {
			eachBlock, err = NewBlock(uint32(blockID), uint8(subBlockID), scrambledBuf)
			if err != nil {
				return nil, nil, err
			}
			eachBlock.SetDB(pm.DB(), fullDBPrefix, objID, blockInfoID)
			err = myEntity.SignBlock(eachBlock)
			if err != nil {
				return nil, nil, err
			}

			err = eachBlock.Save()
			if err != nil {
				return nil, nil, err
			}
			eachHashs[subBlockID] = eachBlock.Hash
		}

		// 5. append hash
		hashs = append(hashs, eachHashs)
	}

	return blockInfoID, hashs, nil
}

/*
ScrambleBuf scrambles the buf.

	1. obtain the json-str to ensure that the 1st-char and the last-char are not 0
	2. do scramble on the json-str.

XXX TODO: better scrambleBuf

*/
func ScrambleBuf(buf [][]byte) ([][]byte, error) {
	theBytes, err := json.Marshal(buf)
	if err != nil {
		return nil, err
	}

	enc, err := reedsolomon.New(NScrambleInBlock, NScrambleInBlock)
	if err != nil {
		return nil, err
	}

	if len(theBytes)%NScrambleInBlock != 0 {
		nAddBytes := NScrambleInBlock - len(theBytes)%NScrambleInBlock
		theBytes = append(theBytes, make([]byte, nAddBytes)...)
	}

	nBytes := len(theBytes)
	nBytesInScramble := nBytes / NScrambleInBlock

	newBufs := make([][]byte, NScrambleInBlock*2)
	for i, pBytes := 0, theBytes; i < NScrambleInBlock; i, pBytes = i+1, pBytes[nBytesInScramble:] {
		newBufs[i] = pBytes[:nBytesInScramble]
	}
	for i := NScrambleInBlock; i < 2*NScrambleInBlock; i++ {
		newBufs[i] = make([]byte, nBytesInScramble)
	}

	err = enc.Encode(newBufs)
	if err != nil {
		return nil, err
	}

	return newBufs[NScrambleInBlock:], nil
}

/*
UnscrambleBuf unscramles the buf.

	1. unscramble the buf.
	2. do json.unmarshal.
*/
func UnscrambleBuf(buf [][]byte) ([][]byte, error) {
	if len(buf) != NScrambleInBlock {
		return nil, ErrInvalidBlock
	}

	enc, err := reedsolomon.New(NScrambleInBlock, NScrambleInBlock)
	if err != nil {
		return nil, err
	}

	nBytesInScramble := len(buf[0])
	theBuf := make([]byte, NScrambleInBlock*nBytesInScramble)
	newBufs := make([][]byte, NScrambleInBlock*2)
	for i := 0; i < NScrambleInBlock; i++ {
		newBufs[i+NScrambleInBlock] = buf[i]
	}

	err = enc.Reconstruct(newBufs)
	if err != nil {
		return nil, err
	}

	for i, pBuf := 0, theBuf; i < NScrambleInBlock; i, pBuf = i+1, pBuf[nBytesInScramble:] {
		copy(pBuf[:nBytesInScramble], newBufs[i])
	}

	// log.Debug("UnscrambleBuf: after Reconstruct", "theBuf", theBuf)

	theBuf = unscrambleBufUnpad(theBuf)

	// log.Debug("UnscrambleBuf: to unmarshal", "theBuf", theBuf)

	result := &[][]byte{}
	err = json.Unmarshal(theBuf, result)
	if err != nil {
		return nil, err
	}
	// log.Debug("UnscrambleBuf: after unmarshal", "result", *result)

	return *result, nil
}

func unscrambleBufUnpad(buf []byte) []byte {
	lenBuf := len(buf)
	i := 0
	for i = lenBuf - 1; i >= 0 && buf[i] == 0; i-- {
	}
	return buf[:(i + 1)]
}

func unscrambleContentBlocksToBufs(contentBlocks []*Block) ([][]byte, error) {
	if len(contentBlocks) != NScrambleInBlock {
		return nil, ErrInvalidBlock
	}
	bufs := make([][]byte, NScrambleInBlock)
	for i, contentBlock := range contentBlocks {
		bufs[i] = contentBlock.Buf
	}

	return UnscrambleBuf(bufs)
}
