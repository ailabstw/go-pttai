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

package simulations

import (
	"errors"
	"testing"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	CodeTypeStatus = 0
	CodeTypeTest   = 1

	NetworkID          = 1
	Version            = 2
	ProtocolMaxMsgSize = 1 * 1024

	BUF_SIZE = ProtocolMaxMsgSize - 1
)

var (
	origHandler log.Handler

	tKey1, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	tKey2, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

	ErrInvalidMsgCode = errors.New("invalid msg code")
	ErrMsgTooLarge    = errors.New("msg too large")
	ErrInvalidData    = errors.New("invalid data")

	LargeFileFilename = "../../e2e/e2e-test.zip"
)

func setupTest(t *testing.T) {
	origHandler = log.Root().GetHandler()
	log.Root().SetHandler(log.Must.FileHandler("log.tmp.txt", log.TerminalFormat(true)))
	log.LogLevel = log.LvlDebug
}

func teardownTest(t *testing.T) {
	log.Root().SetHandler(origHandler)
}
