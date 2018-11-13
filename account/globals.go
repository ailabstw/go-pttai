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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

// config
var (
	DefaultConfig = Config{
		DataDir: filepath.Join(node.DefaultDataDir(), "account"),
	}
)

// protocol
const (
	_ pkgservice.OpType = iota + pkgservice.NMsg
	// user-oplog
	AddUserOplogMsg //30
	AddUserOplogsMsg

	AddPendingUserOplogMsg
	AddPendingUserOplogsMsg

	SyncUserOplogMsg
	SyncUserOplogAckMsg
	SyncUserOplogNewOplogsMsg
	SyncUserOplogNewOplogsAckMsg

	SyncPendingUserOplogMsg
	SyncPendingUserOplogAckMsg

	SyncAddUserNodeMsg
	SyncAddUserNodeAckMsg
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
	dbAccount     *pttdb.LDBBatch    = nil
	dbAccountCore *pttdb.LDBDatabase = nil

	dbMeta *pttdb.LDBDatabase = nil

	DBProfilePrefix = []byte(".pfdb")

	DBUserNamePrefix = []byte(".urnm")
	DBUserImgPrefix  = []byte(".urim")

	DBUserNodePrefix     = []byte(".undb")
	DBUserNodeIdxPrefix  = []byte(".unix")
	DBUserNodeInfoPrefix = []byte(".uidb")
)

// op-key
var (
	RenewOpKeySeconds  int64 = 86400
	ExpireOpKeySeconds int64 = 259200
)

// sync
const (
	MaxSyncRandomSeconds = 30
	MinSyncRandomSeconds = 15
)

// oplog
var (
	DBUserOplogPrefix       = []byte(".urlg")
	DBUserIdxOplogPrefix    = []byte(".urig")
	DBUserMerkleOplogPrefix = []byte(".urmk")
)

func InitAccount(dataDir string) error {
	var err error

	dbAccountCore, err = pttdb.NewLDBDatabase("account", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbAccount, err = pttdb.NewLDBBatch(dbAccountCore)
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
	if dbAccountCore != nil {
		dbAccountCore.Close()
		dbAccountCore = nil
	}

	if dbAccount != nil {
		dbAccount = nil
	}

	if dbMeta != nil {
		dbMeta.Close()
		dbMeta = nil
	}
}
