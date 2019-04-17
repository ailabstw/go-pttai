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

// MIT License
//
// Copyright (c) 2017 bmkessler
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// https://github.com/bmkessler/streamstats/blob/master/bitvector.go

package types

import (
	"encoding/binary"
	"strconv"
)

type BitVector []uint64

// NewBitVector returns a new BitVector of length L
func NewBitVector(L uint64) BitVector {
	// length of backing slice is # of 64-bit words, lower 6 bits index index inside the word
	return BitVector(make([]uint64, 1+((L-1)>>6)))
}

// Set sets the bit at position N
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) Set(N uint64) {
	b[N>>6] |= 1 << (N & 63)
}

// Set sets the bit at position N
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) SetWithIsNew(N uint64) bool {
	if b[N>>6]&(1<<(N&63)) == 1 {
		return false
	}

	b[N>>6] |= 1 << (N & 63)
	return true
}

// Get returns the bit at position N as a uint64
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) Get(N uint64) uint64 {
	return (b[N>>6] >> (N & 63)) & 1
}

// Clear clears the bit at position N
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) Clear(N uint64) {
	b[N>>6] = b[N>>6] &^ (1 << (N & 63))
}

// String outputs a string representation of the binary string with the first bit at the left
// note that any padding zeros are present on the right hand side
func (b BitVector) String() string {
	buff := make([]byte, 0, 64*len(b))
	for _, word := range b {
		bits := []byte(strconv.FormatUint(word, 2))
		for i := len(bits) - 1; i >= 0; i-- {
			buff = append(buff, bits[i]) // append the bits in reverse order
		}
		for j := len(bits); j < 64; j++ {
			buff = append(buff, '0') // add any leading zeros
		}
	}
	return string(buff)
}

// PopCount returns the nubmer of set bits in the bit vector
// the algorithm for PopCount on a single 64-bit word is from
// 1957 due to Donald B. Gillies and Jeffrey C. P. Miller
// and referenced by Donald Knuth
func (b BitVector) PopCount() uint64 {
	var total uint64
	for _, word := range b {
		word = word - ((word) >> 1 & 0x5555555555555555)
		word = (word & 0x3333333333333333) + ((word >> 2) & 0x3333333333333333)
		word = (word + (word >> 4)) & 0x0F0F0F0F0F0F0F0F
		word += (word >> 8)
		word += (word >> 16)
		word += (word >> 32)
		total += word & 255
	}
	return total
}

func (b BitVector) Marshal() ([]byte, error) {
	theBytes := make([]byte, len(b)*8)
	for pb, pByte := []uint64(b), theBytes; len(pb) != 0; pb, pByte = pb[1:], pByte[8:] {
		binary.BigEndian.PutUint64(pByte, pb[0])
	}
	return theBytes, nil
}

func UnmarshalBitVector(theBytes []byte) (BitVector, error) {
	lenBytes := len(theBytes)
	if lenBytes%8 != 0 {
		return nil, ErrInvalidBitVector
	}
	lenBitVector := lenBytes / 8
	b := make([]uint64, lenBitVector)
	for pb, pByte := b, theBytes; len(pb) != 0; pb, pByte = pb[1:], pByte[8:] {
		pb[0] = binary.BigEndian.Uint64(pByte)
	}

	return BitVector(b), nil
}
