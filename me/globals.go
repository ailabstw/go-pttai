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

	SyncPendingMeOplogMsg
	SyncPendingMeOplogAckMsg

	// raft

	SendRaftMsgsMsg // 53
	RequestRaftLeadMsg

	// init-me-info
	InitMeInfoMsg
	InitMeInfoAckMsg
	InitMeInfoSyncMsg
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

// sync
const (
	MaxSyncRandomSeconds = 8
	MinSyncRandomSeconds = 5
)

// join
const (
	SyncJoinSeconds = 10 * time.Second

	RenewJoinFriendKeySeconds = pkgservice.RenewJoinKeySeconds
)

// op-key
var (
	RenewOpKeySeconds  int64 = 86400
	ExpireOpKeySeconds int64 = 259200
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
	dbMeCore, err = pttdb.NewLDBDatabase("me", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbMe, err = pttdb.NewLDBBatch(dbMeCore)
	if err != nil {
		return err
	}

	dbMyNodes, err = pttdb.NewLDBDatabase("mynodes", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbRaft, err = pttdb.NewLDBDatabase("raft", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbMeta, err = pttdb.NewLDBDatabase("memeta", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbKeyCore, err = pttdb.NewLDBDatabase("signkey", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbKey, err = pttdb.NewLDBBatch(dbKeyCore)
	if err != nil {
		return err
	}

	return nil
}

func TeardownMe() {
	if dbMeCore != nil {
		dbMeCore.Close()
		dbMeCore = nil
	}

	if dbMe != nil {
		dbMe = nil
	}

	if dbMyNodes != nil {
		dbMyNodes.Close()
		dbMyNodes = nil
	}

	if dbRaft != nil {
		dbRaft.Close()
		dbRaft = nil
	}

	if dbMeta != nil {
		dbMeta.Close()
		dbMeta = nil
	}

	if dbKeyCore != nil {
		dbKeyCore.Close()
		dbKeyCore = nil
	}

	if dbKey != nil {
		dbKey = nil
	}
}
