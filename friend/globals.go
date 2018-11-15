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

package friend

import (
	"path/filepath"

	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

// config
var (
	DefaultConfig = Config{
		DataDir: filepath.Join(node.DefaultDataDir(), "friend"),
	}
)

// db
var (
	dbFriendCore *pttdb.LDBDatabase = nil
	dbFriend     *pttdb.LDBBatch    = nil
	dbKey        *pttdb.LDBDatabase = nil

	dbMeta *pttdb.LDBDatabase = nil

	DBFriendIdxPrefix         = []byte(".frix")
	DBFriendIdx2Prefix        = []byte(".fri2")
	DBFriendPrefix            = []byte(".frdb")
	DBFriendOplogPrefix       = []byte(".frlg")
	DBFriendIdxOplogPrefix    = []byte(".frig")
	DBFriendMerkleOplogPrefix = []byte(".frmk")
)

// protocol
const (
	_ pkgservice.OpType = iota + pkgservice.NMsg
	// friend-oplog
	AddFriendOplogMsg //30
	AddFriendOplogsMsg

	AddPendingFriendOplogMsg
	AddPendingFriendOplogsMsg

	SyncFriendOplogMsg
	SyncFriendOplogAckMsg
	SyncFriendOplogNewOplogsMsg
	SyncFriendOplogNewOplogsAckMsg

	SyncPendingFriendOplogMsg
	SyncPendingFriendOplogAckMsg

	SyncAddFriendNodeMsg
	SyncAddFriendNodeAckMsg

	// init friend info
	InitFriendInfoMsg
	InitFriendInfoAckMsg
)

// max-masters
const (
	MaxMasters = 2
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

// message
const (
	NFirstLineInBlock = 20
)

func InitFriend(dataDir string) error {
	var err error

	dbFriendCore, err = pttdb.NewLDBDatabase("friend", dataDir, 0, 0)
	if err != nil {
		return err
	}
	dbFriend, err = pttdb.NewLDBBatch(dbFriendCore)
	if err != nil {
		return err
	}

	dbMeta, err = pttdb.NewLDBDatabase("friendmeta", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbKey, err = pttdb.NewLDBDatabase("friendkey", dataDir, 0, 0)
	if err != nil {
		return err
	}

	return nil
}

func TeardownFriend() {
	if dbKey != nil {
		dbKey.Close()
		dbKey = nil
	}

	if dbFriendCore != nil {
		dbFriendCore.Close()
		dbFriendCore = nil
	}
	if dbFriend != nil {
		dbFriend = nil
	}

	if dbMeta != nil {
		dbMeta.Close()
		dbMeta = nil
	}
}
