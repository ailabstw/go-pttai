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
	"crypto/ecdsa"
	"path/filepath"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

// default config
var (
	DefaultConfig = Config{
		DataDir:  filepath.Join(node.DefaultDataDir(), "me"),
		NodeType: pkgservice.NodeTypeDesktop,
	}
)

// MyInfo
var (
	MyID       *types.PttID
	MyNodeID   *discover.NodeID
	MyKey      *ecdsa.PrivateKey
	MyNodeType pkgservice.NodeType = pkgservice.NodeTypeDesktop

	MyNodeSignID *types.PttID
	MyRaftID     uint64
)

// defaults
var (
	DataDirPrivateKey = "mykey"

	DefaultTitle = []byte("")
)

// db
var (
	DBMePrefix                    = []byte(".medb")
	dbMe       *pttdb.LDBDatabase = nil
	dbMeBatch  *pttdb.LDBBatch    = nil

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

// raft

const (
	RaftTickTime        = 100 * time.Millisecond
	RaftElectionTick    = 50
	RaftHeartbeatTick   = 5
	RaftMaxSizePerMsg   = 1024 * 1024
	RaftMaxInflightMsgs = 16
)

func InitMe(dataDir string) error {
	var err error

	// db
	dbMe, err = pttdb.NewLDBDatabase("me", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbMeBatch, err = pttdb.NewLDBBatch(dbMe)
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

func initMyInfo(id *types.PttID, nodeID *discover.NodeID, key *ecdsa.PrivateKey, nodeType pkgservice.NodeType) error {
	MyID = id
	MyNodeID = nodeID
	MyKey = key
	MyNodeType = nodeType

	nodeIDPubkey, err := MyNodeID.Pubkey()
	if err != nil {
		return err
	}

	MyNodeSignID, err = types.NewPttIDWithPubkeyAndRefID(nodeIDPubkey, MyID)
	if err != nil {
		return err
	}

	MyRaftID, err = MyNodeID.ToRaftID()
	if err != nil {
		return err
	}

	return nil
}

func TeardownMe() {
	if dbMe != nil {
		dbMe.Close()
		dbMe = nil
	}

	if dbMeBatch != nil {
		dbMeBatch = nil
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
