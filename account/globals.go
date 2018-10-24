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
	"path/filepath"

	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/pttdb"
)

// config
var (
	DefaultConfig = Config{
		DataDir: filepath.Join(node.DefaultDataDir(), "account"),
	}
)

// user-profile
const (
	MaxProfileImgWidth  = 128
	MinProfileImgWidth  = 64
	MaxProfileImgHeight = 128
	MinProfileImgHeight = 64
	MaxProfileImgSize   = 65535

	MaxNameLength         = 25
	ProfileImageMaskRatio = 0.8
)

// db
var (
	dbAccount *pttdb.LDBDatabase = nil

	dbMeta *pttdb.LDBDatabase = nil

	DBUserNamePrefix = []byte(".urnm")
	DBUserImgPrefix  = []byte(".urim")

	DBUserNodePrefix    = []byte(".undb")
	DBUserNodeIdxPrefix = []byte(".unix")
)

func InitAccount(dataDir string) error {
	var err error

	dbAccount, err = pttdb.NewLDBDatabase("account", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbMeta, err = pttdb.NewLDBDatabase("accountmeta", dataDir, 0, 0)
	if err != nil {
		return err
	}

	return nil
}

func TeardownAccount() {
	if dbAccount != nil {
		dbAccount.Close()
		dbAccount = nil
	}

	if dbMeta != nil {
		dbMeta.Close()
		dbMeta = nil
	}
}
