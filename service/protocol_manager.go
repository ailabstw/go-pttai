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
	"sync"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/event"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
)

type ProtocolManager interface {
	Start() error
	Stop() error

	HandleMessage(op OpType, dataBytes []byte, peer *PttPeer) error

	Sync(peer *PttPeer) error

	// join
	ApproveJoin(joinEntity *JoinEntity, keyInfo *KeyInfo, peer *PttPeer) (*KeyInfo, interface{}, error)

	GetJoinType(hash *common.Address) (JoinType, error)

	// master
	Master0Hash() []byte
	IsMaster(id *types.PttID) bool

	// owner-id
	SetOwnerID(ownerID *types.PttID, isLocked bool)
	GetOwnerID(isLocked bool) *types.PttID

	// oplog
	isValidOplog(signInfos []*SignInfo) (*types.PttID, uint32, bool)

	// peers
	IsMyDevice(peer *PttPeer) bool
	IsImportantPeer(peer *PttPeer) bool
	IsMemberPeer(peer *PttPeer) bool

	IsSuspiciousID(id *types.PttID, nodeID *discover.NodeID) bool
	IsGoodID(id *types.PttID, nodeID *discover.NodeID) bool

	/**********
	 * implemented in base-protocol-manager
	 **********/

	// event-mux
	EventMux() *event.TypeMux

	// master
	SetNewestMasterLogID(id *types.PttID) error
	GetNewestMasterLogID() *types.PttID

	// join
	GetJoinKeyInfo(hash *common.Address) (*KeyInfo, error)
	GetJoinKey() (*KeyInfo, error)

	CreateJoinKeyInfoLoop() error

	JoinKeyInfos() []*KeyInfo

	// op
	GetOpKeyInfoFromHash(hash *common.Address) (*KeyInfo, error)
	GetNewestOpKey() (*KeyInfo, error)
	GetOldestOpKey() (*KeyInfo, error)

	RegisterOpKeyInfo(keyInfo *KeyInfo) error

	RemoveOpKeyInfoFromHash(hash *common.Address) error
	RemoveOpKeyInfo(keyInfo *KeyInfo) error

	OpKeyInfos() map[common.Address]*KeyInfo

	SaveOpKeyInfo(opKeyInfo *KeyInfo) error

	RevokeKeyChan() chan *KeyInfo

	RenewOpKeySeconds() uint64
	ExpireOpKeySeconds() uint64

	DBOpKeyLock() *types.LockMap
	DBOpKeyInfo() *pttdb.LDBBatch

	TryCreateOpKeyInfo() error

	// peers
	Peers() *PttPeerSet

	NewPeerCh() chan *PttPeer

	NoMorePeers() chan struct{}
	SetNoMorePeers(noMorePeers chan struct{})

	RegisterPeer(peer *PttPeer) error
	UnregisterPeer(peer *PttPeer) error

	GetPeerType(peer *PttPeer) PeerType

	IdentifyPeer(peer *PttPeer)
	HandleIdentifyPeer(dataBytes []byte, peer *PttPeer) error

	IdentifyPeerAck(data *IdentifyPeer, peer *PttPeer) error
	HandleIdentifyPeerAck(dataBytes []byte, peer *PttPeer) error

	// sync
	ForceSyncCycle() time.Duration

	QuitSync() chan struct{}
	SetQuitSync(quitSync chan struct{})

	SyncWG() *sync.WaitGroup

	// entity
	Entity() Entity

	// ptt
	Ptt() Ptt

	// db
	DB() *pttdb.LDBBatch
}

type BaseProtocolManager struct {
	// eventMux
	eventMux *event.TypeMux

	// master
	newestMasterLogID *types.PttID

	// owner-id
	lockOwnerID sync.RWMutex
	ownerID     *types.PttID

	// join
	lockJoinKeyInfo sync.RWMutex

	joinKeyInfos []*KeyInfo

	// op
	lockOpKeyInfo sync.RWMutex

	opKeyInfos      map[common.Address]*KeyInfo
	newestOpKeyInfo *KeyInfo
	oldestOpKeyInfo *KeyInfo

	renewOpKeySeconds  uint64
	expireOpKeySeconds uint64

	dbOpKeyLock *types.LockMap

	// oplog
	isValidOplog func(signInfos []*SignInfo) (*types.PttID, uint32, bool)

	// peer
	peers       *PttPeerSet
	newPeerCh   chan *PttPeer
	noMorePeers chan struct{}

	// sync
	maxSyncRandomSeconds int
	minSyncRandomSeconds int

	quitSync chan struct{}
	syncWG   *sync.WaitGroup

	// entity
	entity Entity

	// ptt
	ptt Ptt

	// db
	db *pttdb.LDBBatch

	// is-start
	isStart bool
}

func NewBaseProtocolManager(ptt Ptt, renewOpKeySeconds uint64, expireOpKeySeconds uint64, maxSyncRandomSeconds int, minSyncRandomSeconds int, isValidOplog func(signInfos []*SignInfo) (*types.PttID, uint32, bool), e Entity, db *pttdb.LDBBatch) (*BaseProtocolManager, error) {

	peers, err := NewPttPeerSet()
	if err != nil {
		return nil, err
	}

	dbOpKeyLock, err := types.NewLockMap(SleepTimeOpKeyLock)
	if err != nil {
		return nil, err
	}

	pm := &BaseProtocolManager{
		eventMux: new(event.TypeMux),

		// join
		joinKeyInfos: make([]*KeyInfo, 0),

		// op
		renewOpKeySeconds:  renewOpKeySeconds,
		expireOpKeySeconds: expireOpKeySeconds,

		dbOpKeyLock: dbOpKeyLock,

		// oplog
		isValidOplog: isValidOplog,

		// peers
		newPeerCh: make(chan *PttPeer),
		peers:     peers,

		// sync
		maxSyncRandomSeconds: maxSyncRandomSeconds,
		minSyncRandomSeconds: minSyncRandomSeconds,

		// entity
		entity: e,

		// ptt
		ptt: ptt,

		// db
		db: db,
	}

	// op-key
	opKeyInfos, err := pm.loadOpKeyInfos()
	if err != nil {
		return nil, err
	}

	pm.lockOpKeyInfo.Lock()
	defer pm.lockOpKeyInfo.Unlock()

	ptt.LockOps()
	defer ptt.UnlockOps()

	for _, keyInfo := range opKeyInfos {
		pm.RegisterOpKeyInfo(keyInfo, true, true)
	}

	// master-log-id
	newestMasterLogID, err := pm.loadNewestMasterLogID()
	if err != nil {
		return nil, err
	}

	pm.newestMasterLogID = newestMasterLogID

	return pm, nil

}

func (pm *BaseProtocolManager) Start() error {
	pm.isStart = true

	return nil
}

func (pm *BaseProtocolManager) Stop() error {
	pm.eventMux.Stop()

	return nil
}

func (pm *BaseProtocolManager) EventMux() *event.TypeMux {
	return pm.eventMux
}

func (pm *BaseProtocolManager) Entity() Entity {
	return pm.entity
}

func (pm *BaseProtocolManager) Ptt() Ptt {
	return pm.ptt
}

func (pm *BaseProtocolManager) DB() *pttdb.LDBBatch {
	return pm.db
}
