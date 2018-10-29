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

package service

import (
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

// default config
var (
	DefaultConfig = Config{
		MaxPeers:          350,
		MaxImportantPeers: 100,
		MaxMemberPeers:    200,
		MaxRandomPeers:    50,
	}
)

// protocol
const (
	_ uint = iota
	Ptt1
)

var (
	ProtocolVersions = [1]uint{Ptt1}
	ProtocolName     = "ptt1"
	ProtocolLengths  = [1]uint64{uint64(NCodeType)}
)

// ptt-layer
const (
	ProtocolMaxMsgSize = 10 * 1024 * 1024 // 4MB for video-streaming

	SizeOpType   = 4 // optype uint32
	SizeCodeType = 8 // codetype uint64

	SizeChallenge = 16

	HandshakeTimeout    = 60 * time.Second
	IdentifyPeerTimeout = 10 * time.Second
)

// join
const (
	IntRenewJoinKeySeconds = 86400 // 1 day for now
	RenewJoinKeySeconds    = time.Duration(IntRenewJoinKeySeconds) * time.Second
)

// op
const (
	_ OpType = iota

	// join

	JoinMsg
	JoinAckChallengeMsg

	JoinEntityMsg
	ApproveJoinMsg

	JoinAlreadyRegisteredMsg
	JoinAckAlreadyRegistedMsg

	// op-key
	AddOpKeyOplogMsg
	AddOpKeyOplogsMsg

	AddPendingOpKeyOplogMsg
	AddPendingOpKeyOplogsMsg

	RevokeOpKeyInfoMsg

	// peer
	IdentifyPeerMsg
	IdentifyPeerAckMsg

	// me
	AddMeOplogMsg
	AddMeOplogsMsg

	AddPendingMeOplogMsg
	AddPendingMeOplogsMsg

	BoardLastSeenMsg
	ArticleLastSeenMsg
	NMsg
)

// op-key
const (
	MaxIterDeriveKeyBIP32 = 10

	SleepTimeOpKeyLock = 10

	ExpireOpKeySeconds = 259200
)

var (
	DBOpKeyIdxOplogPrefix    = []byte(".okig")
	DBOpKeyOplogPrefix       = []byte(".oklg")
	DBOpKeyMerkleOplogPrefix = []byte(".oklg")
	DBOpKeyPrefix            = []byte(".okdb")
	DBOpKeyIdxPrefix         = []byte(".okix")
	DBOpKeyIdx2Prefix        = []byte(".oki2")
)

// db
const (
	SleepTimeMasterLock = 10

	SleepTimeMeLock = 10

	SleepTimePttLock = 10

	MaxCountPttOplog = 2000
	PPttOplog        = 12 // 2^12 = 4096
)

var (
	dbOplog     *pttdb.LDBBatch
	dbOplogCore *pttdb.LDBDatabase

	dbMeta *pttdb.LDBDatabase

	DBNewestMasterLogIDPrefix = []byte(".nmld")

	DBMasterOplogPrefix       = []byte(".malg")
	DBMasterIdxOplogPrefix    = []byte(".maig")
	DBMasterMerkleOplogPrefix = []byte(".mamk")
	DBMasterLockMap           *types.LockMap

	DBMeOplogPrefix       = []byte(".melg")
	DBMeIdxOplogPrefix    = []byte(".meig")
	DBMeMerkleOplogPrefix = []byte(".memk")
	DBMeLockMap           *types.LockMap

	DBCountPttOplogPrefix = []byte(".ptct")

	DBPttOplogPrefix       = []byte(".ptlg") // .ptlm, .ptli is used as well
	DBPttIdxOplogPrefix    = []byte(".ptig")
	DBPttMerkleOplogPrefix = []byte(".ptmk")
	DBPttLockMap           *types.LockMap

	DBLocalePrefix     = []byte(".locl")
	DBPttLogSeenPrefix = []byte(".ptsn")
)

// oplog
const (
	ExpireOplogSeconds       = 300 // expire oplog circulation as 5 minutes for now.
	OffsetMasterOplogRaftIdx = 12
)

// oplog-merkle-tree
var (
	SizeMerkleTreeLevel     = 1 // uint8
	SizeMerkleTreeNChildren = 4 // uint32
	NMerkleTreeMagicAlloc   = 50
	MerkleTreeOffsetAddr    = SizeMerkleTreeLevel
	MerkleTreeOffsetTS      = MerkleTreeOffsetAddr + common.AddressLength

	DBMerkleGenerateTimePrefix = []byte(".mtgt")
	DBMerkleSyncTimePrefix     = []byte(".mtst")
	DBMerkleFailSyncTimePrefix = []byte(".mtft")

	OffsetMerkleSyncTime uint64 = 3600 // validate until 2-hr ago, and sync with data starting 2-hr ago.

	GenerateOplogMerkleTreeSeconds              = 900 * time.Second // 15 mins
	ExpireGenerateOplogMerkleTreeSeconds uint64 = 450               // 7.5 mins
)

var (
	ExpireDialHistorySeconds = uint64(30)
	DialHistoryLoopInterval  = 30 * time.Second
)

func InitService(dataDir string) error {
	dbOplogCore, err := pttdb.NewLDBDatabase("oplog", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbOplog, err = pttdb.NewLDBBatch(dbOplogCore)
	if err != nil {
		return err
	}

	dbMeta, err = pttdb.NewLDBDatabase("meta", dataDir, 0, 0)
	if err != nil {
		return err
	}

	DBMasterLockMap, err = types.NewLockMap(SleepTimeMasterLock)
	if err != nil {
		return err
	}

	DBMeLockMap, err = types.NewLockMap(SleepTimeMeLock)
	if err != nil {
		return err
	}

	DBPttLockMap, err = types.NewLockMap(SleepTimePttLock)
	if err != nil {
		return err
	}

	return nil
}

func TeardownService() {
	if dbOplog != nil {
		dbOplog = nil
	}

	if dbOplogCore != nil {
		dbOplogCore.Close()
		dbOplogCore = nil
	}

	if dbMeta != nil {
		dbMeta.Close()
		dbMeta = nil
	}

	if DBMasterLockMap != nil {
		DBMasterLockMap = nil
	}

	if DBMeLockMap != nil {
		DBMeLockMap = nil
	}

	if DBPttLockMap != nil {
		DBPttLockMap = nil
	}
}
