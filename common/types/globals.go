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

package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shengdoushi/base58"
)

// protocol-versions
const (
	_ uint = iota
	Account1
	Me1
	Friend1
)

const (
	SizeTimestamp = 12 // uint64 + uint32 (Sizeof may align with 64bit)
	SizePostfix   = common.AddressLength
	SizePttID     = common.AddressLength + SizePostfix // address-length (integrate with user-id)

	SizeSalt   = 32
	OffsetHash = 12
)

const (
	HRSeconds  = 3600
	DaySeconds = 86400
)

const (
	MaxImageWidth  = 10000
	MaxImageHeight = 10000
)

// id / timestamp
var (
	EmptyID = PttID{}
	MaxID   = PttID{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	}
	EmptyIDBytes = EmptyID[:]
	quoteBytes   = []byte{'"'}

	ZeroTimestamp = Timestamp{0, 0}

	//                       12345678901234567  123456789
	MaxTimestamp = Timestamp{99999999999999999, 999999999}
	pttIDDataDir = ""

	// unable to use this because http-url would use percent encoding for this
	// pronunciation:
	//
	//     zero, one, two, three, four, five, six, seven, eight, nine, ten,
	//     https://en.wikipedia.org/wiki/Bopomofo
	//     jia3, i3, bim3, ding1, wu4, chi3, gen1, shin1, jen2, kue3
	// myAlphabet = base58.NewAlphabet("零壹二三四五六七八九十ㄅㄆㄇㄈㄉㄊㄋㄌㄍㄎㄏㄐㄑㄒㄓㄔㄕㄖㄗㄘㄙㄧㄨㄩㄚㄛㄜㄝㄞㄟㄠㄡㄢㄣㄤㄥㄦ甲乙丙丁戊己庚辛壬癸")
	// BitcoinAlphabet:
	// (0123456789012345678901234567890123456789012345678901234567)
	//  123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
	myAlphabet = base58.BitcoinAlphabet

	NIterLock = 100

	OffsetSecond int64 = 0
)
