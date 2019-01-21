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
	AddUserOplogMsg //42
	AddUserOplogsMsg

	AddPendingUserOplogMsg
	AddPendingUserOplogsMsg

	SyncUserOplogMsg
	ForceSyncUserOplogMsg
	ForceSyncUserOplogAckMsg
	InvalidSyncUserOplogMsg
	SyncUserOplogAckMsg
	SyncUserOplogNewOplogsMsg
	SyncUserOplogNewOplogsAckMsg

	SyncPendingUserOplogMsg // 50
	SyncPendingUserOplogAckMsg

	// user-node
	SyncAddUserNodeMsg
	SyncAddUserNodeAckMsg

	ForceSyncUserNodeMsg
	ForceSyncUserNodeAckMsg

	// user-name
	SyncCreateUserNameMsg
	SyncCreateUserNameAckMsg

	SyncUpdateUserNameMsg // 56
	SyncUpdateUserNameAckMsg

	ForceSyncUserNameMsg
	ForceSyncUserNameAckMsg

	// user-img
	SyncCreateUserImgMsg
	SyncCreateUserImgAckMsg

	SyncUpdateUserImgMsg // 60
	SyncUpdateUserImgAckMsg

	ForceSyncUserImgMsg
	ForceSyncUserImgAckMsg

	// name-card
	SyncCreateNameCardMsg
	SyncCreateNameCardAckMsg

	SyncUpdateNameCardMsg // 56
	SyncUpdateNameCardAckMsg

	ForceSyncNameCardMsg
	ForceSyncNameCardAckMsg
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

	DBUserNamePrefix    = []byte(".umdb")
	DBUserNameIdxPrefix = []byte(".umix")
	DBUserImgPrefix     = []byte(".uidb")
	DBUserImgIdxPrefix  = []byte(".uiix")
	DBNameCardPrefix    = []byte(".ncdb")
	DBNameCardIdxPrefix = []byte(".ncix")

	DBUserNodePrefix     = []byte(".undb")
	DBUserNodeIdxPrefix  = []byte(".unix")
	DBUserNodeInfoPrefix = []byte(".uidb")
)

// max-masters
const (
	MaxMasters = 1
)

// sync
const (
	MaxSyncRandomSeconds = 30
	MinSyncRandomSeconds = 15
)

// op-key
var (
	RenewOpKeySeconds  int64 = 86400
	ExpireOpKeySeconds int64 = 259200
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
