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

package types

import (
	"testing"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/crypto"
)

const ()

var (
	tDefaultPttID = PttID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}

	tDefaultKey2, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")

	tDefaultPttID2 = PttID{13, 58, 177, 75, 186, 211, 217, 159, 66, 3, 189, 122, 17, 172, 185, 72, 130, 5, 14, 126, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}

	tDefaultPttID3 = PttID{13, 58, 177, 75, 186, 211, 217, 159, 66, 3, 189, 122, 17, 172, 185, 72, 130, 5, 14, 126, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}

	origRandRead func(b []byte) (int, error) = nil

	origHandler log.Handler
)

func setupTest(t *testing.T) {
	origRandRead = RandRead
	RandRead = func(b []byte) (int, error) {
		lenB := len(b)
		for i := 0; i < lenB; i++ {
			b[i] = uint8(i % 0xff)
		}
		return lenB, nil
	}

	origHandler = log.Root().GetHandler()
	log.Root().SetHandler(log.Must.FileHandler("log.tmp.txt", log.TerminalFormat(true)))
}

func teardownTest(t *testing.T) {
	RandRead = origRandRead

	log.Root().SetHandler(origHandler)
}
