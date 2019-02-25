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
	"reflect"
	"sync"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/event"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
)

type ProtocolManager interface {
	Prestart() error
	Start() error
	Prestop() error
	Stop() error
	Poststop() error

	HandleMessage(op OpType, dataBytes []byte, peer *PttPeer) error

	Sync(peer *PttPeer) error

	LeaveEntity() (bool, error)
	Delete() error
	PostdeleteEntity(opData OpData, isForce bool) error

	FullCleanLog()

	// join
	ApproveJoin(
		joinEntity *JoinEntity,
		keyInfo *KeyInfo,
		peer *PttPeer,
	) (*KeyInfo, interface{}, error)

	GetJoinType(hash *common.Address) (JoinType, error)

	// master
	MasterLog0Hash() []byte
	SetMasterLog0Hash(theBytes []byte) error

	AddMaster(id *types.PttID, isForce bool) (*Master, *MasterOplog, error)
	MigrateMaster(id *types.PttID) error
	TransferMaster(id *types.PttID) error

	IsMaster(id *types.PttID, isLocked bool) bool

	RegisterMaster(master *Master, isLocked bool, isSkipPtt bool) error

	// member

	AddMember(id *types.PttID, isForce bool) (*Member, *MemberOplog, error)
	MigrateMember(fromID *types.PttID, toID *types.PttID) error
	DeleteMember(id *types.PttID) (bool, error)

	// owner-id
	SetOwnerID(ownerID *types.PttID, isLocked bool)
	GetOwnerID(isLocked bool) *types.PttID

	// oplog
	BroadcastOplog(oplog *BaseOplog, msg OpType, pendingMsg OpType) error
	BroadcastOplogs(oplogs []*BaseOplog, msg OpType, pendingMsg OpType) error

	SignOplog(oplog *BaseOplog) error
	ForceSignOplog(oplog *BaseOplog) error

	IntegrateOplog(oplog *BaseOplog, isLocked bool, merkle *Merkle) (bool, error)
	InternalSign(oplog *BaseOplog) (bool, error)

	// peers
	IsMyDevice(peer *PttPeer) bool
	IsImportantPeer(peer *PttPeer) bool
	IsMemberPeer(peer *PttPeer) bool
	IsPendingPeer(peer *PttPeer) bool

	IsSuspiciousID(id *types.PttID, nodeID *discover.NodeID) bool
	IsGoodID(id *types.PttID, nodeID *discover.NodeID) bool

	LoadPeers() error

	/**********
	 * implemented in base-protocol-manager
	 **********/

	// event-mux
	EventMux() *event.TypeMux

	// master
	SetNewestMasterLogID(id *types.PttID) error
	GetNewestMasterLogID() *types.PttID

	GetMasterListFromCache(isLocked bool) ([]*Master, error)
	GetMasterList(startID *types.PttID, limit int, listOrder pttdb.ListOrder, isLocked bool) ([]*Master, error)

	// master-oplog
	BroadcastMasterOplog(log *MasterOplog) error

	HandleAddMasterOplog(dataBytes []byte, peer *PttPeer) error
	HandleAddMasterOplogs(dataBytes []byte, peer *PttPeer) error
	HandleAddPendingMasterOplog(dataBytes []byte, peer *PttPeer) error
	HandleAddPendingMasterOplogs(dataBytes []byte, peer *PttPeer) error

	HandleSyncMasterOplog(dataBytes []byte, peer *PttPeer) error

	HandleForceSyncMasterOplog(dataBytes []byte, peer *PttPeer) error
	HandleForceSyncMasterOplogAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncMasterOplogInvalidAck(dataBytes []byte, peer *PttPeer) error

	HandleSyncMasterOplogAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncNewMasterOplog(dataBytes []byte, peer *PttPeer) error
	HandleSyncNewMasterOplogAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncPendingMasterOplog(dataBytes []byte, peer *PttPeer) error
	HandleSyncPendingMasterOplogAck(dataBytes []byte, peer *PttPeer) error

	HandleMasterOplogs(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error

	GetMasterOplogList(logID *types.PttID, limit int, listOrder pttdb.ListOrder, status types.Status) ([]*MasterOplog, error)

	GetMasterOplogMerkleNodeList(level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*MerkleNode, error)

	// member

	GetMemberList(startID *types.PttID, limit int, listOrder pttdb.ListOrder, isLocked bool) ([]*Member, error)

	// member-oplog
	BroadcastMemberOplog(log *MemberOplog) error

	HandleAddMemberOplog(dataBytes []byte, peer *PttPeer) error
	HandleAddMemberOplogs(dataBytes []byte, peer *PttPeer) error
	HandleAddPendingMemberOplog(dataBytes []byte, peer *PttPeer) error
	HandleAddPendingMemberOplogs(dataBytes []byte, peer *PttPeer) error

	HandleSyncMemberOplog(dataBytes []byte, peer *PttPeer) error

	HandleForceSyncMemberOplog(dataBytes []byte, peer *PttPeer) error
	HandleForceSyncMemberOplogAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncMemberOplogInvalidAck(dataBytes []byte, peer *PttPeer) error

	HandleSyncMemberOplogAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncNewMemberOplog(dataBytes []byte, peer *PttPeer) error
	HandleSyncNewMemberOplogAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncPendingMemberOplog(dataBytes []byte, peer *PttPeer) error
	HandleSyncPendingMemberOplogAck(dataBytes []byte, peer *PttPeer) error

	HandleMemberOplogs(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error
	SetMemberSyncTime(ts types.Timestamp) error

	GetMemberOplogList(logID *types.PttID, limit int, listOrder pttdb.ListOrder, status types.Status) ([]*MemberOplog, error)
	GetMemberOplogMerkleNodeList(level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*MerkleNode, error)

	GetMemberLogByMemberID(id *types.PttID, isLocked bool) (*MemberOplog, error)

	MyMemberLog() *MemberOplog

	// log0
	SetLog0DB(oplog *BaseOplog)
	Log0Merkle() *Merkle

	HandleLog0s(logs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error

	// join
	GetJoinKeyFromHash(hash *common.Address) (*KeyInfo, error)
	GetJoinKey() (*KeyInfo, error)

	CreateJoinKeyLoop() error

	JoinKeyList() []*KeyInfo

	// op

	GetOpKeyFromHash(hash *common.Address, isLocked bool) (*KeyInfo, error)
	GetNewestOpKey(isLocked bool) (*KeyInfo, error)
	GetOldestOpKey(isLocked bool) (*KeyInfo, error)

	RegisterOpKey(keyInfo *KeyInfo, isLocked bool) error

	RevokeOpKey(keyID *types.PttID) (bool, error)
	RemoveOpKeyFromHash(hash *common.Address, isLocked bool, isDeleteOplog bool, isDeleteDB bool) error

	OpKeys() map[common.Address]*KeyInfo
	OpKeyList() []*KeyInfo

	RenewOpKeySeconds() int64
	ExpireOpKeySeconds() int64
	GetToRenewOpKeySeconds() int
	ToRenewOpKeyTS() (types.Timestamp, error)

	DBOpKeyLock() *types.LockMap
	DBOpKey() *pttdb.LDBBatch
	DBOpKeyPrefix() []byte
	DBOpKeyIdxPrefix() []byte

	SetOpKeyDB(oplog *BaseOplog)

	SetOpKeyObjDB(opKey *KeyInfo)

	GetOpKeyListFromDB() ([]*KeyInfo, error)

	CreateOpKeyLoop() error

	CreateOpKey() error
	ForceCreateOpKey() error

	ForceOpKey() chan struct{}

	// op-key-oplog

	BroadcastOpKeyOplog(log *OpKeyOplog) error

	HandleAddOpKeyOplog(dataBytes []byte, peer *PttPeer) error
	HandleAddOpKeyOplogs(dataBytes []byte, peer *PttPeer) error
	HandleAddPendingOpKeyOplog(dataBytes []byte, peer *PttPeer) error
	HandleAddPendingOpKeyOplogs(dataBytes []byte, peer *PttPeer) error

	SyncOpKeyOplog(peer *PttPeer, syncMsg OpType) error
	HandleSyncOpKeyOplog(dataBytes []byte, peer *PttPeer, syncMsg OpType) error
	HandleSyncPendingOpKeyOplog(dataBytes []byte, peer *PttPeer) error
	HandleSyncPendingOpKeyOplogAck(dataBytes []byte, peer *PttPeer) error

	HandleOpKeyOplogs(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error

	HandleSyncCreateOpKey(dataBytes []byte, peer *PttPeer) error
	HandleSyncCreateOpKeyAck(dataBytes []byte, peer *PttPeer) error
	HandleSyncCreateOpKeyAckObj(opKey *KeyInfo, peer *PttPeer) error

	GetOpKeyOplogList(logID *types.PttID, limit int, listOrder pttdb.ListOrder, status types.Status) ([]*OpKeyOplog, error)

	// peers
	Peers() *PttPeerSet

	NewPeerCh() chan *PttPeer

	NoMorePeers() chan struct{}
	SetNoMorePeers(noMorePeers chan struct{})

	RegisterPeer(peer *PttPeer, peerType PeerType, isLocked bool) error
	RegisterPendingPeer(peer *PttPeer, isLocked bool) error
	UnregisterPeer(peer *PttPeer, isForceReset bool, isForceNotReset bool, isPttLocked bool) error

	GetPeerType(peer *PttPeer) PeerType

	IdentifyPeer(peer *PttPeer)
	HandleIdentifyPeer(dataBytes []byte, peer *PttPeer) error

	IdentifyPeerAck(data *IdentifyPeer, peer *PttPeer) error
	HandleIdentifyPeerAck(dataBytes []byte, peer *PttPeer) error

	SendDataToPeer(op OpType, data interface{}, peer *PttPeer) error
	SendDataToPeers(op OpType, data interface{}, peerList []*PttPeer) error

	CountPeers() (int, error)
	GetPeers() ([]*PttPeer, error)

	// sync
	ForceSyncCycle() time.Duration

	ForceSync() chan struct{}
	QuitSync() chan struct{}

	SyncWG() *sync.WaitGroup

	// entity
	Entity() Entity

	SaveNewEntityWithOplog(oplog *BaseOplog, isLocked bool, isForce bool) error

	MaybePostcreateEntity(
		oplog *BaseOplog,
		isForce bool,
		postcreateEntity func(entity Entity) error,
	) error

	GetEntityLog() (*BaseOplog, error)

	HandleEntityTerminal(status types.Status, entityLog *BaseOplog, peer *PttPeer) error

	// ptt
	Ptt() Ptt

	// db
	DB() *pttdb.LDBBatch
	DBObjLock() *types.LockMap
}

type MyProtocolManager interface {
	ProtocolManager

	SetMeDB(log *BaseOplog)
	SetMasterDB(log *BaseOplog)
}

type PttProtocolManager interface {
	ProtocolManager
}

type BaseProtocolManager struct {
	// eventMux
	eventMux *event.TypeMux

	// master
	newestMasterLogID *types.PttID

	isMaster          func(id *types.PttID, isLocked bool) bool
	dbMasterPrefix    []byte
	dbMasterIdxPrefix []byte

	masterLog0Hash []byte

	lockMaster sync.RWMutex
	masters    map[types.PttID]*Master
	maxMasters int

	inposttransferMaster func(theMaster Object, theNewMaster Object, oplog *BaseOplog) error

	// master-oplog
	dbMasterLock *types.LockMap
	masterMerkle *Merkle

	// member
	isMember          func(id *types.PttID, isLocked bool) bool
	dbMemberPrefix    []byte
	dbMemberIdxPrefix []byte

	inpostdeleteMember func(id *types.PttID, oplog *BaseOplog, origObj Object, opData OpData) error

	// member-oplog
	dbMemberLock *types.LockMap
	memberMerkle *Merkle
	myMemberLog  *MemberOplog

	// peer
	getPeerType func(peer *PttPeer) PeerType

	isMyDevice      func(peer *PttPeer) bool
	isImportantPeer func(peer *PttPeer) bool
	isMemberPeer    func(peer *PttPeer) bool
	isPendingPeer   func(peer *PttPeer) bool

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

	renewOpKeySeconds  int64
	expireOpKeySeconds int64

	dbOpKeyPrefix    []byte
	dbOpKeyIdxPrefix []byte

	forceOpKey chan struct{}

	// op-key-oplog
	dbOpKeyLock *types.LockMap

	// oplog
	internalSign          func(oplog *BaseOplog) (bool, error)
	forceSignOplog        func(oplog *BaseOplog) error
	isValidOplog          func(signInfos []*SignInfo) (*types.PttID, uint32, bool)
	validateIntegrateSign func(oplog *BaseOplog, isLocked bool) error

	oplog0      *BaseOplog
	log0Merkle  *Merkle
	setLog0DB   func(oplog *BaseOplog)
	handleLog0s func(logs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error

	// peer
	peers       *PttPeerSet
	newPeerCh   chan *PttPeer
	noMorePeers chan struct{}

	sendDataToPeersSub        *event.TypeMuxSubscription
	sendDataToPeerWithCodeSub *event.TypeMuxSubscription

	// sync
	maxSyncRandomSeconds int
	minSyncRandomSeconds int

	forceSync chan struct{}
	quitSync  chan struct{}
	syncWG    sync.WaitGroup

	postsyncMemberOplog func(peer *PttPeer) error

	// entity
	entity Entity

	theDelete  func() error
	postdelete func(opData OpData, isForce bool) error

	// ptt
	ptt Ptt

	// db
	db     *pttdb.LDBBatch
	dbLock *types.LockMap

	// block
	dbBlockPrefix []byte

	// media
	dbMediaPrefix    []byte
	dbMediaIdxPrefix []byte

	// is-start
	isStart    bool
	isPrestart bool
}

func NewBaseProtocolManager(
	ptt Ptt,
	renewOpKeySeconds int64,
	expireOpKeySeconds int64,
	maxSyncRandomSeconds int,
	minSyncRandomSeconds int,

	maxMasters int,

	log0Merkle *Merkle,

	internalSign func(oplog *BaseOplog) (bool, error),
	forceSignOplog func(oplog *BaseOplog) error,
	isValidOplog func(signInfos []*SignInfo) (*types.PttID, uint32, bool),
	validateIntegrateSign func(oplog *BaseOplog, isLocked bool) error,

	setLog0DB func(oplog *BaseOplog),
	handleLog0s func(logs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error,

	isMaster func(id *types.PttID, isLocked bool) bool,
	isMember func(id *types.PttID, isLocked bool) bool,

	getPeerType func(peer *PttPeer) PeerType,
	isMyDevice func(peer *PttPeer) bool,
	isImportantPeer func(peer *PttPeer) bool,
	isMemberPeer func(peer *PttPeer) bool,
	isPendingPeer func(peer *PttPeer) bool,

	postsyncMemberOplog func(peer *PttPeer) error,

	theDelete func() error,
	postdelete func(opData OpData, isForce bool) error,

	// entity
	e Entity,

	// db
	db *pttdb.LDBBatch,

) (*BaseProtocolManager, error) {

	entityID := e.GetID()

	peers, err := NewPttPeerSet()
	if err != nil {
		return nil, err
	}

	// db-lock
	dbLock, err := types.NewLockMap(SleepTimeLock)
	if err != nil {
		return nil, err
	}

	// master
	dbMasterLock, err := types.NewLockMap(SleepTimeLock)
	if err != nil {
		return nil, err
	}
	dbMasterPrefix := append(DBMasterPrefix, entityID[:]...)
	dbMasterIdxPrefix := append(DBMasterIdxPrefix, entityID[:]...)
	masterMerkle, err := NewMerkle(DBMasterOplogPrefix, DBMasterMerkleOplogPrefix, entityID, db)
	if err != nil {
		return nil, err
	}

	// member
	dbMemberLock, err := types.NewLockMap(SleepTimeLock)
	if err != nil {
		return nil, err
	}
	dbMemberPrefix := append(DBMemberPrefix, entityID[:]...)
	dbMemberIdxPrefix := append(DBMemberIdxPrefix, entityID[:]...)
	memberMerkle, err := NewMerkle(DBMemberOplogPrefix, DBMemberMerkleOplogPrefix, entityID, db)
	if err != nil {
		return nil, err
	}

	// op-key
	dbOpKeyLock, err := types.NewLockMap(SleepTimeOpKeyLock)
	if err != nil {
		return nil, err
	}
	dbOpKeyPrefix := append(DBOpKeyPrefix, entityID[:]...)
	dbOpKeyIdxPrefix := append(DBOpKeyIdxPrefix, entityID[:]...)

	// block
	dbBlockPrefix := append(DBBlockInfoPrefix, entityID[:]...)

	// media
	dbMediaPrefix := append(DBMediaPrefix, entityID[:]...)
	dbMediaIdxPrefix := append(DBMediaIdxPrefix, entityID[:]...)

	pm := &BaseProtocolManager{
		eventMux: new(event.TypeMux),

		// join
		joinKeyInfos: make([]*KeyInfo, 0),

		// master
		isMaster:          isMaster,
		dbMasterPrefix:    dbMasterPrefix,
		dbMasterIdxPrefix: dbMasterIdxPrefix,
		masters:           make(map[types.PttID]*Master),
		maxMasters:        maxMasters,

		// master-oplog
		dbMasterLock: dbMasterLock,
		masterMerkle: masterMerkle,

		// member
		isMember:          isMember,
		dbMemberPrefix:    dbMemberPrefix,
		dbMemberIdxPrefix: dbMemberIdxPrefix,

		// member-oplog
		dbMemberLock: dbMemberLock,
		memberMerkle: memberMerkle,

		// op
		renewOpKeySeconds:  renewOpKeySeconds,
		expireOpKeySeconds: expireOpKeySeconds,

		opKeyInfos: make(map[common.Address]*KeyInfo),

		dbOpKeyPrefix:    dbOpKeyPrefix,
		dbOpKeyIdxPrefix: dbOpKeyIdxPrefix,

		forceOpKey: make(chan struct{}),

		// op-key-oplog
		dbOpKeyLock: dbOpKeyLock,

		// oplog
		internalSign:          internalSign,
		forceSignOplog:        forceSignOplog,
		isValidOplog:          isValidOplog,
		validateIntegrateSign: validateIntegrateSign,

		log0Merkle:  log0Merkle,
		setLog0DB:   setLog0DB,
		handleLog0s: handleLog0s,

		// peers
		newPeerCh: make(chan *PttPeer),
		peers:     peers,

		getPeerType:     getPeerType,
		isMyDevice:      isMyDevice,
		isImportantPeer: isImportantPeer,
		isMemberPeer:    isMemberPeer,
		isPendingPeer:   isPendingPeer,

		// sync
		maxSyncRandomSeconds: maxSyncRandomSeconds,
		minSyncRandomSeconds: minSyncRandomSeconds,

		forceSync: make(chan struct{}),
		quitSync:  make(chan struct{}),

		postsyncMemberOplog: postsyncMemberOplog,

		// entity
		entity: e,

		theDelete:  theDelete,
		postdelete: postdelete,

		// ptt
		ptt: ptt,

		// db
		db:     db,
		dbLock: dbLock,

		// block
		dbBlockPrefix: dbBlockPrefix,

		// media
		dbMediaPrefix:    dbMediaPrefix,
		dbMediaIdxPrefix: dbMediaIdxPrefix,
	}
	if pm.internalSign == nil {
		pm.internalSign = pm.defaultInternalSign
	}
	if pm.forceSignOplog == nil {
		pm.forceSignOplog = pm.defaultForceSignOplog
	}
	if pm.isValidOplog == nil {
		pm.isValidOplog = pm.defaultIsValidOplog
	}
	if pm.validateIntegrateSign == nil {
		pm.validateIntegrateSign = pm.defaultValidateIntegrateSign
	}
	if pm.isMaster == nil {
		pm.isMaster = pm.defaultIsMaster
	}
	if pm.isMember == nil {
		pm.isMember = pm.defaultIsMember
	}
	if pm.getPeerType == nil {
		pm.getPeerType = pm.defaultGetPeerType
	}
	if pm.isMyDevice == nil {
		pm.isMyDevice = pm.defaultIsMyDevice
	}
	if pm.isImportantPeer == nil {
		pm.isImportantPeer = pm.defaultIsImportantPeer
	}
	if pm.isMemberPeer == nil {
		pm.isMemberPeer = pm.defaultIsMemberPeer
	}
	if pm.isPendingPeer == nil {
		pm.isPendingPeer = pm.defaultIsPendingPeer
	}
	if pm.theDelete == nil {
		pm.theDelete = pm.defaultDelete
	}
	if pm.postdelete == nil {
		pm.postdelete = pm.DefaultPostdeleteEntity
	}

	return pm, nil

}

func (pm *BaseProtocolManager) HandleMessage(op OpType, dataBytes []byte, peer *PttPeer) error {
	return types.ErrNotImplemented
}

func (pm *BaseProtocolManager) Prestart() error {
	if pm.isPrestart {
		log.Warn("Prestart: already prestarted", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
		return ErrAlreadyPrestarted
	}
	pm.isPrestart = true

	// send-data-to-peer sub
	pm.sendDataToPeersSub = pm.eventMux.Subscribe(&SendDataToPeersEvent{})
	go pm.sendDataToPeersLoop()

	pm.sendDataToPeerWithCodeSub = pm.eventMux.Subscribe(&SendDataToPeerWithCodeEvent{})
	go pm.sendDataToPeerWithCodeLoop()

	// master
	log.Debug("Prestart: to loadMasters", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	masters, err := pm.loadMasters()
	if err != nil {
		return err
	}

	log.Debug("Prestart: after loadMasters", "entity", pm.Entity().GetID(), "masters", len(masters))
	for _, master := range masters {
		log.Debug("Prestart (in-for-loop)", "master", master.ID)
	}

	pm.registerMasters(masters, false)

	// master-log-id
	newestMasterLogID, err := pm.loadNewestMasterLogID()
	if err != nil {
		return err
	}

	pm.newestMasterLogID = newestMasterLogID

	// master0hash
	masterLog0hash, err := pm.loadMasterLog0Hash()
	if err == nil {
		pm.masterLog0Hash = masterLog0hash
	}

	// op-key
	opKeyInfos, err := pm.loadOpKeyInfos()
	if err != nil {
		return err
	}

	log.Debug("Prestart: after loadOpKeyInfos", "opKeyInfos", opKeyInfos)

	pm.registerOpKeys(opKeyInfos, false)

	log.Debug("Prestart: after registerOpKeys")

	// to register
	if len(opKeyInfos) == 0 {
		pm.CreateOpKey()
	}

	// load myMemberLog
	entity := pm.Entity()
	service := entity.Service()
	myService := pm.Ptt().GetMyService()
	myEntity := pm.Ptt().GetMyEntity().(Entity)

	if service != myService {
		log.Debug("Prestart: to loadMyMemberLog")
		err = pm.loadMyMemberLog()
		log.Debug("Prestart: after loadMyMemberLog", "e", err, "service", pm.Entity().Service().Name())
		if err != nil {
			return err
		}
	}

	// load oplog0
	oplog := &BaseOplog{}
	pm.SetLog0DB(oplog)
	oplogs, err := GetOplogList(oplog, nil, 1, pttdb.ListOrderNext, types.StatusAlive, false)
	log.Debug("Prestart: after GetOplogList", "entity", pm.Entity().GetID(), "e", err, "oplogs", len(oplogs))
	if entity == myEntity && len(oplogs) == 0 {
		return nil
	}
	if len(oplogs) != 1 {
		return ErrInvalidOplog
	}
	pm.oplog0 = oplogs[0]

	return nil
}

func (pm *BaseProtocolManager) Start() error {
	if pm.isStart {
		log.Warn("Start: already started", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
		return ErrAlreadyStarted
	}

	pm.isStart = true

	log.Info("Start: to master merkle-tree", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

	syncWG := pm.SyncWG()
	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		PMOplogMerkleTreeLoop(pm, pm.masterMerkle)
	}()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		PMOplogMerkleTreeLoop(pm, pm.memberMerkle)
	}()

	myID := pm.Ptt().GetMyEntity().GetID()
	// check owner
	entity := pm.Entity()

	var err error
	if !entity.IsOwner(myID) {
		owners := entity.GetOwnerIDs()
		log.Warn("Start: I am not the owner", "myID", myID, "entityID", entity.GetID(), "owners", owners)
		for _, ownerID := range owners {
			if !reflect.DeepEqual(myID, ownerID) {
				log.Debug("Start: to MigrateMember", "entity", entity.GetID(), "ownerID", ownerID, "myID", myID)
				err = pm.MigrateMember(ownerID, myID)
				if err == nil {
					break
				}
			}
		}
		if err != nil {
			log.Warn("Start: I am still not the owner, resetting as new me", "entity", entity.GetID(), "e", err)
			entity.ResetOwnerIDs()
			entity.AddOwnerID(myID)
			return entity.Save(false)
		}
	}

	return nil
}

func (pm *BaseProtocolManager) Prestop() error {
	log.Debug("Prestop: start", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name(), "isStart", pm.isStart)
	close(pm.quitSync)

	return nil
}

func (pm *BaseProtocolManager) Stop() error {
	return nil
}

func (pm *BaseProtocolManager) Poststop() error {
	if pm.isStart {
		pm.syncWG.Wait()
	}

	if pm.isPrestart {
		pm.sendDataToPeersSub.Unsubscribe()
		pm.sendDataToPeerWithCodeSub.Unsubscribe()
	}

	pm.eventMux.Stop()

	log.Debug("Stopped", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

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

func (pm *BaseProtocolManager) DBObjLock() *types.LockMap {
	return pm.dbLock
}

func (pm *BaseProtocolManager) MasterMerkle() *Merkle {
	return pm.masterMerkle
}

func (pm *BaseProtocolManager) MemberMerkle() *Merkle {
	return pm.memberMerkle
}

func (pm *BaseProtocolManager) IsStart() bool {
	return pm.isStart
}

func (pm *BaseProtocolManager) LoadPeers() error {
	return nil
}

func (pm *BaseProtocolManager) MyMemberLog() *MemberOplog {
	return pm.myMemberLog
}
