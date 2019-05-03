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

package me

import (
	"path/filepath"
	"time"

	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

// default config
var (
	DefaultConfig = Config{
		DataDir: filepath.Join(node.DefaultDataDir(), "me"),
	}
)

// defaults
var (
	DataDirPrivateKey = "mykey"

	DefaultTitle = []byte("")
)

// protocol
const (
	_ pkgservice.OpType = iota + pkgservice.NMsg

	JoinFriendMsg

	// me-oplog
	AddMeOplogMsg
	AddMeOplogsMsg

	AddPendingMeOplogMsg
	AddPendingMeOplogsMsg

	SyncMeOplogMsg // 47
	SyncMeOplogAckMsg
	SyncMeOplogNewOplogsMsg
	SyncMeOplogNewOplogsAckMsg

	InvalidSyncMeOplogMsg

	ForceSyncMeOplogMsg
	ForceSyncMeOplogAckMsg
	ForceSyncMeOplogByMerkleMsg
	ForceSyncMeOplogByMerkleAckMsg
	ForceSyncMeOplogByOplogAckMsg

	SyncPendingMeOplogMsg
	SyncPendingMeOplogAckMsg

	// raft

	SendRaftMsgsMsg // 53
	RequestRaftLeadMsg

	// init-me-info
	InitMeInfoMsg
	InitMeInfoAckMsg
	InitMeInfoSyncMsg

	// sync-board
	InternalSyncBoardMsg
	InternalSyncBoardAckMsg

	// sync-friend
	InternalSyncFriendMsg
	InternalSyncFriendAckMsg
)

// db
var (
	SleepTimeLock = 10

	DBMePrefix                    = []byte(".medb")
	dbMeCore   *pttdb.LDBDatabase = nil
	dbMe       *pttdb.LDBBatch    = nil

	DBMyNodePrefix = []byte(".mndb")

	DBRaftPrefix                    = []byte(".rfdb")
	dbRaft       *pttdb.LDBDatabase = nil

	dbMyNodes *pttdb.LDBDatabase = nil

	dbMeta *pttdb.LDBDatabase = nil

	dbKeyCore *pttdb.LDBDatabase = nil
	dbKey     *pttdb.LDBBatch    = nil

	DBKeyRaftHardState = []byte(".rfhs")
	DBKeyRaftSnapshot  = []byte(".rfsn")

	// raft in me
	DBKeyRaftLastIndex     = []byte(".rfli")
	DBKeyRaftAppliedIndex  = []byte(".rfai")
	DBKeyRaftSnapshotIndex = []byte(".rfsi")
	DBKeyRaftConfState     = []byte(".rfcs")
	DBKeyRaftLead          = []byte(".rfld")
)

// max-masters

const (
	MaxMasters = 0
)

// sync
const (
	MaxSyncRandomSeconds = 8
	MinSyncRandomSeconds = 5
)

// op-key
var (
	RenewOpKeySeconds  int64 = 86400
	ExpireOpKeySeconds int64 = 259200
)

// join
const (
	SyncJoinSeconds = 10 * time.Second

	RenewJoinFriendKeySeconds = pkgservice.RenewJoinKeySeconds
)

// sign-key
const (
	NRenewSignKey = 100
)

// oplog
var (
	DBMeOplogPrefix       = []byte(".melg")
	DBMeIdxOplogPrefix    = []byte(".meig")
	DBMeMerkleOplogPrefix = []byte(".memk")

	DBMasterOplogPrefix    = []byte(".malg")
	DBMasterIdxOplogPrefix = []byte(".maig")
)

// me-oplog
const (
	GenerateMeOplogMerkleTreeSeconds = 10 * time.Second

	ExpireGenerateMeOplogMerkleTreeSeconds = 60
	OffsetGenerateMeOplogMerkleTreeSeconds = 7200

	SleepTimeMeLock = 10
)

// master-oplog
const (
	OffsetMasterOplogRaftIdx = 12

	SleepTimeMasterLock = 10
)

var (
	MasterIDZeros = make([]byte, OffsetMasterOplogRaftIdx)
)

// raft

const (
	RaftTickTime        = 100 * time.Millisecond
	RaftElectionTick    = 50
	RaftHeartbeatTick   = 5
	RaftMaxSizePerMsg   = 1024 * 1024
	RaftMaxInflightMsgs = 16

	NRequestRaftLead = 10
)

// weight
const (
	WeightServer  = 2000000
	WeightDesktop = 2000
	WeightMobile  = 2
)

// init-me-info

const (
	InitMeInfoTickTime = 3 * time.Second
)

func InitMe(dataDir string) error {
	var err error

	// db
	if dbMeCore == nil {
		dbMeCore, err = pttdb.NewLDBDatabase("me", dataDir, 0, 0)
		if err != nil {
			return err
		}
	}

	if dbMe == nil {
		dbMe, err = pttdb.NewLDBBatch(dbMeCore)
		if err != nil {
			return err
		}
	}

	if dbMyNodes == nil {
		dbMyNodes, err = pttdb.NewLDBDatabase("mynodes", dataDir, 0, 0)
		if err != nil {
			return err
		}
	}

	if dbRaft == nil {
		dbRaft, err = pttdb.NewLDBDatabase("raft", dataDir, 0, 0)
		if err != nil {
			return err
		}
	}

	if dbMeta == nil {
		dbMeta, err = pttdb.NewLDBDatabase("memeta", dataDir, 0, 0)
		if err != nil {
			return err
		}
	}

	if dbKeyCore == nil {
		dbKeyCore, err = pttdb.NewLDBDatabase("signkey", dataDir, 0, 0)
		if err != nil {
			return err
		}
	}

	if dbKey == nil {
		dbKey, err = pttdb.NewLDBBatch(dbKeyCore)
		if err != nil {
			return err
		}
	}

	return nil
}

func TeardownMe() {
}
