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
		MaxHubPeers:       5,
		MaxImportantPeers: 100,
		MaxMemberPeers:    200,
		MaxPendingPeers:   50,
		MaxRandomPeers:    50,

		NodeType: NodeTypeDesktop,

		ExpireOplogSeconds: 180,

		IsE2E: false,

		IsPrivateAsPublic: false,
	}
)

// protocol
const (
	_ uint = iota + 1
	Ptt2
)

var (
	ProtocolVersions = [1]uint{Ptt2}
	ProtocolName     = "ptt2"
	ProtocolLengths  = [1]uint64{uint64(NCodeType)}
)

// ptt-layer
const (
	ProtocolMaxMsgSize = 20 * 1024 * 1024 // 20MB for video-streaming

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

// msg
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

	AddOpKeyOplogMsg // 7
	AddOpKeyOplogsMsg
	AddPendingOpKeyOplogMsg
	AddPendingOpKeyOplogsMsg

	SyncOpKeyOplogMsg
	ForceSyncOpKeyOplogMsg
	ForceSyncOpKeyOplogAckMsg
	InvalidSyncOpKeyOplogMsg
	SyncOpKeyOplogAckMsg
	SyncPendingOpKeyOplogMsg
	SyncPendingOpKeyOplogAckMsg

	SyncCreateOpKeyMsg
	SyncCreateOpKeyAckMsg

	// master
	AddMasterOplogMsg // 17
	AddMasterOplogsMsg

	AddPendingMasterOplogMsg
	AddPendingMasterOplogsMsg

	SyncMasterOplogMsg
	ForceSyncMasterOplogMsg
	ForceSyncMasterOplogAckMsg
	InvalidSyncMasterOplogMsg
	SyncMasterOplogAckMsg
	SyncMasterOplogNewOplogsMsg
	SyncMasterOplogNewOplogsAckMsg

	SyncPendingMasterOplogMsg
	SyncPendingMasterOplogAckMsg

	// member
	AddMemberOplogMsg // 27
	AddMemberOplogsMsg

	AddPendingMemberOplogMsg
	AddPendingMemberOplogsMsg

	SyncMemberOplogMsg
	ForceSyncMemberOplogMsg
	ForceSyncMemberOplogAckMsg
	InvalidSyncMemberOplogMsg
	SyncMemberOplogAckMsg
	SyncMemberOplogNewOplogsMsg
	SyncMemberOplogNewOplogsAckMsg

	SyncPendingMemberOplogMsg
	SyncPendingMemberOplogAckMsg

	// peer
	IdentifyPeerMsg // 37
	IdentifyPeerAckMsg

	BoardLastSeenMsg
	ArticleLastSeenMsg

	NMsg
)

// member
var (
	DBMasterPrefix    = []byte(".MAdb")
	DBMasterIdxPrefix = []byte(".MAix")

	DBMasterOplogPrefix       = []byte(".MAlg")
	DBMasterIdxOplogPrefix    = []byte(".MAig")
	DBMasterMerkleOplogPrefix = []byte(".MAmk")
)

// member
var (
	DBMemberPrefix    = []byte(".mbdb")
	DBMemberIdxPrefix = []byte(".mbix")

	DBMemberOplogPrefix       = []byte(".mblg")
	DBMemberIdxOplogPrefix    = []byte(".mbig")
	DBMemberMerkleOplogPrefix = []byte(".mbmk")
)

// op-key
const (
	MaxIterDeriveKeyBIP32 = 10

	SleepTimeOpKeyLock = 10
)

var (
	DBOpKeyPrefix     = []byte(".okdb")
	DBOpKeyIdxPrefix  = []byte(".okix")
	DBOpKeyIdx2Prefix = []byte(".oki2")

	DBOpKeyOplogPrefix    = []byte(".oklg")
	DBOpKeyIdxOplogPrefix = []byte(".okig")
)

// oplog
const (
	MaxSyncOplogAck = 200
)

// object
const (
	MaxSyncObjectAck = 50
	MaxSyncBlock     = 50
)

// block

const (
	NSubBlock        = 2
	NLineInBlock     = 20
	NScrambleInBlock = 2
)

var (
	DBBlockInfoPrefix    = []byte(".bidb")
	DBBlockInfoIdxPrefix = []byte(".biix")

	DBContentBlockPrefix = []byte(".bkdb")
)

// media
const (
	NByteInBlock = 65535

	MaxUploadMediaSize = 10485760 // 10MB

	MaxUploadImageWidth  = 8192
	MaxUploadImageHeight = 8192
)

var (
	DBMediaPrefix    = []byte(".mddb")
	DBMediaIdxPrefix = []byte(".mdix")
)

// db
const (
	SleepTimePttLock = 10

	SleepTimeLock = 10

	MaxCountPttOplog = 2000
	PPttOplog        = 12 // 2^12 = 4096
)

var (
	dbOplog     *pttdb.LDBBatch
	dbOplogCore *pttdb.LDBDatabase

	dbMeta *pttdb.LDBDatabase

	DBNewestMasterLogIDPrefix = []byte(".nmld")
	DBMasterLog0HashPrefix    = []byte(".ml0h")

	DBCountPttOplogPrefix = []byte(".ptct")

	DBPttOplogPrefix    = []byte(".ptlg") // .ptlm, .ptli is used as well
	DBPttIdxOplogPrefix = []byte(".ptig")
	DBPttLockMap        *types.LockMap

	DBLocalePrefix     = []byte(".locl")
	DBPttLogSeenPrefix = []byte(".ptsn")
)

// oplog
var (
	ExpireOplogSeconds = 300 // expire oplog circulation as 5 minutes for now.
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
	DBMerkleMetaPostfix        = []byte("mt")
	DBMerkleToUpdatePostfix    = []byte("Mt")
	DBMerkleUpdatingPostfix    = []byte("MT")

	OffsetMerkleSyncTime int64 = 3600 // validate until 2-hr ago, and sync with data starting 2-hr ago.

	GenerateOplogMerkleTreeSeconds             = 900 * time.Second // 15 mins
	ExpireGenerateOplogMerkleTreeSeconds int64 = 450               // 7.5 mins
)

// dial-history
var (
	ExpireDialHistorySeconds int64 = 30
	DialHistoryLoopInterval        = 30 * time.Second
)

// locale
var (
	DefaultLocale Locale = LocaleTW
	CurrentLocale Locale = DefaultLocale
)

// misc
var (
	IsE2E = false

	IsPrivateAsPublic = false
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

	DBPttLockMap, err = types.NewLockMap(SleepTimePttLock)
	if err != nil {
		return err
	}

	CurrentLocale = LoadLocale()

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

	if DBPttLockMap != nil {
		DBPttLockMap = nil
	}
}
