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

package account

import (
	"crypto/ecdsa"
	"os"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
)

const ()

var (
	tKeyA      *ecdsa.PrivateKey = nil
	tUserIDA   *types.PttID      = nil
	tTsA       types.Timestamp   = types.Timestamp{}
	tUserNameA *UserName         = nil

	tKeyB      *ecdsa.PrivateKey = nil
	tUserIDB   *types.PttID      = nil
	tTsB       types.Timestamp   = types.Timestamp{}
	tUserNameB *UserName         = nil

	tKeyC      *ecdsa.PrivateKey = nil
	tUserIDC   *types.PttID      = nil
	tTsC       types.Timestamp   = types.Timestamp{}
	tUserNameC *UserName         = nil

	tKeyD      *ecdsa.PrivateKey = nil
	tUserIDD   *types.PttID      = nil
	tTsD       types.Timestamp   = types.Timestamp{}
	tUserNameD *UserName         = nil

	origRandRead func(b []byte) (int, error) = nil
)

func setupTest(t *testing.T) {
	origRandRead = types.RandRead
	types.RandRead = func(b []byte) (int, error) {
		lenB := len(b)
		for i := 0; i < lenB; i++ {
			b[i] = uint8(i % 0xff)
		}
		return lenB, nil
	}

	tKeyA, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	tUserIDA, _ = types.NewPttIDFromKey(tKeyA)
	tTsA = types.Timestamp{Ts: 1, NanoTs: 5}
	tUserNameA, _ = NewUserName(tUserIDA, tTsA)

	tKeyB, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	tUserIDB, _ = types.NewPttIDFromKey(tKeyB)
	tTsB = types.Timestamp{Ts: 2, NanoTs: 6}
	tUserNameB, _ = NewUserName(tUserIDB, tTsB)

	tKeyC, _ = crypto.HexToECDSA("869d6ecf5211f1cc60418a13b9d870b22959d0c16f02bec714c960dd2298a32d")
	tUserIDC, _ = types.NewPttIDFromKey(tKeyC)
	tTsC = types.Timestamp{Ts: 3, NanoTs: 7}
	tUserNameC, _ = NewUserName(tUserIDC, tTsC)

	tKeyD, _ = crypto.HexToECDSA("e238eb8e04fee6511ab04c6dd3c89ce097b11f25d584863ac2b6d5b35b1847e4")
	tUserIDD, _ = types.NewPttIDFromKey(tKeyD)
	tTsD = types.Timestamp{Ts: 4, NanoTs: 8}
	tUserNameD, _ = NewUserName(tUserIDD, tTsD)

	InitAccount("./test.out")

}

func teardownTest(t *testing.T) {
	types.RandRead = origRandRead

	TeardownAccount()

	os.RemoveAll("./test.out")
}
