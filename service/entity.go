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
	"bytes"
	"reflect"
	"sort"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

/*
Entity is the fundamental interface of entities.
The reason why entity is not an object is because:
	1. Creating an entity does not require the consensus from other nodes.
	2. We may recover a deleted entity.
*/
type Entity interface {
	PrestartAndStart() error
	Prestart() error
	Start() error
	Stop() error

	GetUpdateTS() types.Timestamp
	SetUpdateTS(ts types.Timestamp)

	Save(isLocked bool) error

	Init(ptt Ptt, service Service, spm ServiceProtocolManager) error

	/**********
	 * implemented in BaseEntity
	 **********/

	GetID() *types.PttID
	SetID(id *types.PttID)

	GetCreateTS() types.Timestamp
	GetCreatorID() *types.PttID

	SetJoinTS(ts types.Timestamp)

	GetUpdaterID() *types.PttID
	SetUpdaterID(id *types.PttID)

	GetLogID() *types.PttID
	SetLogID(id *types.PttID)

	GetUpdateLogID() *types.PttID
	SetUpdateLogID(id *types.PttID)

	GetStatus() types.Status
	SetStatus(status types.Status)

	GetOwnerIDs() []*types.PttID
	AddOwnerID(id *types.PttID)
	RemoveOwnerID(id *types.PttID)
	IsOwner(id *types.PttID) bool

	GetEntityType() EntityType
	SetEntityType(t EntityType)

	PM() ProtocolManager
	Ptt() Ptt
	Service() Service

	Name() string
	SetName(name string)

	DB() *pttdb.LDBBatch
	DBLock() *types.LockMap
	SetDB(db *pttdb.LDBBatch, dbLock *types.LockMap)

	MustLock() error
	Lock() error
	Unlock() error
	RLock() error
	RUnlock() error

	SetSyncInfo(syncInfo SyncInfo)
	GetSyncInfo() SyncInfo
}

type BaseEntity struct {
	V         types.Version
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`

	JoinTS types.Timestamp `json:"JT"`

	LogID       *types.PttID `json:"l,omitempty"`
	UpdateLogID *types.PttID `json:"u,omitempty"`

	Status types.Status `json:"S"`

	OwnerIDs []*types.PttID `json:"o,omitempty"`

	EntityType EntityType `json:"e"`

	pm      ProtocolManager
	name    string
	ptt     Ptt
	service Service

	db     *pttdb.LDBBatch
	dbLock *types.LockMap

	SyncInfo SyncInfo
}

func NewBaseEntity(id *types.PttID, createTS types.Timestamp, creatorID *types.PttID, status types.Status, db *pttdb.LDBBatch, dbLock *types.LockMap) *BaseEntity {

	e := &BaseEntity{
		V:         types.CurrentVersion,
		ID:        id,
		CreateTS:  createTS,
		JoinTS:    createTS,
		CreatorID: creatorID,
		UpdaterID: creatorID,
		Status:    status,
		OwnerIDs:  make([]*types.PttID, 0),
		db:        db,
		dbLock:    dbLock,
	}
	e.OwnerIDs = append(e.OwnerIDs, creatorID)

	return e
}

func (e *BaseEntity) Init(pm ProtocolManager, ptt Ptt, service Service) {
	e.pm = pm
	e.ptt = ptt
	e.service = service
}

func (e *BaseEntity) SetDB(db *pttdb.LDBBatch, dbLock *types.LockMap) {
	e.db = db
	e.dbLock = dbLock
}

func (e *BaseEntity) PrestartAndStart() error {
	err := e.Prestart()
	log.Debug("PrestartAndStart: after Prestart", "e", err, "entity", e.GetID(), "service", e.Service().Name())
	if err != nil {
		return err
	}
	return e.Start()
}

func (e *BaseEntity) Prestart() error {
	return PrestartPM(e.pm)
}

func (e *BaseEntity) Start() error {
	return StartPM(e.pm)
}

func (e *BaseEntity) Stop() error {
	return StopPM(e.pm)
}

// implemented in BaseEntity
func (e *BaseEntity) GetID() *types.PttID {
	return e.ID
}

func (e *BaseEntity) SetID(id *types.PttID) {
	e.ID = id
}

func (e *BaseEntity) GetCreateTS() types.Timestamp {
	return e.CreateTS
}

func (e *BaseEntity) GetCreatorID() *types.PttID {
	return e.CreatorID
}

func (e *BaseEntity) GetUpdaterID() *types.PttID {
	return e.UpdaterID
}

func (e *BaseEntity) SetUpdaterID(id *types.PttID) {
	e.UpdaterID = id
}

func (e *BaseEntity) GetLogID() *types.PttID {
	return e.LogID
}

func (e *BaseEntity) SetLogID(id *types.PttID) {
	e.LogID = id
}

func (e *BaseEntity) GetUpdateLogID() *types.PttID {
	return e.UpdateLogID
}

func (e *BaseEntity) SetUpdateLogID(id *types.PttID) {
	e.UpdateLogID = id
}

func (e *BaseEntity) GetStatus() types.Status {
	return e.Status
}

func (e *BaseEntity) SetStatus(status types.Status) {
	e.Status = status
}

func (e *BaseEntity) GetOwnerIDs() []*types.PttID {
	return e.OwnerIDs
}

func (e *BaseEntity) AddOwnerID(id *types.PttID) {
	ownerIDs := e.OwnerIDs
	idx := sort.Search(len(ownerIDs), func(i int) bool {
		return bytes.Compare(ownerIDs[i][:], id[:]) >= 0
	})

	log.Debug("AddOwnerID: after search", "idx", idx, "id", id)

	if idx == len(ownerIDs) {
		e.OwnerIDs = append(ownerIDs, id)
		return
	}

	if reflect.DeepEqual(id, ownerIDs[idx]) {
		return
	}

	// insert-into-slice
	ownerIDs = append(ownerIDs, nil)
	copy(ownerIDs[(idx+1):], ownerIDs[idx:])
	ownerIDs[idx] = id

	log.Debug("AddOwnerID: after append-append", "ownerIDs", ownerIDs)

	e.OwnerIDs = ownerIDs
}

func (e *BaseEntity) RemoveOwnerID(id *types.PttID) {
	ownerIDs := e.OwnerIDs
	idx := sort.Search(len(ownerIDs), func(i int) bool {
		return bytes.Compare(ownerIDs[i][:], id[:]) >= 0
	})

	if idx == len(ownerIDs) {
		return
	}

	if !reflect.DeepEqual(id, ownerIDs[idx]) {
		return
	}

	e.OwnerIDs = append(ownerIDs[:idx], ownerIDs[(idx+1):]...)
}

func (e *BaseEntity) IsOwner(id *types.PttID) bool {
	ownerIDs := e.OwnerIDs
	idx := sort.Search(len(ownerIDs), func(i int) bool {
		return bytes.Compare(ownerIDs[i][:], id[:]) >= 0
	})

	log.Debug("IsOwner: after Search idx", "id", id, "idx", idx, "ownerIDs", ownerIDs)

	if idx == len(ownerIDs) {
		return false
	}

	log.Debug("IsOwner: to DeepEqual", "id", id, "owerID", ownerIDs[idx])

	return reflect.DeepEqual(id, ownerIDs[idx])
}

func (e *BaseEntity) PM() ProtocolManager {
	return e.pm
}

func (e *BaseEntity) Ptt() Ptt {
	return e.ptt
}

func (e *BaseEntity) Service() Service {
	return e.service
}

func (e *BaseEntity) Name() string {
	return e.name
}

func (e *BaseEntity) SetName(name string) {
	e.name = name
}

func (e *BaseEntity) GetEntityType() EntityType {
	return e.EntityType
}

func (e *BaseEntity) SetEntityType(t EntityType) {
	e.EntityType = t
}
func (e *BaseEntity) DB() *pttdb.LDBBatch {
	return e.db
}

func (e *BaseEntity) DBLock() *types.LockMap {
	return e.dbLock
}

func (e *BaseEntity) Lock() error {
	return e.dbLock.Lock(e.GetID())
}

func (e *BaseEntity) MustLock() error {
	return e.dbLock.MustLock(e.GetID())
}

func (e *BaseEntity) Unlock() error {
	return e.dbLock.Unlock(e.GetID())
}

func (e *BaseEntity) RLock() error {
	return e.dbLock.RLock(e.GetID())
}

func (e *BaseEntity) RUnlock() error {
	return e.dbLock.RUnlock(e.GetID())
}

func (e *BaseEntity) SetSyncInfo(syncInfo SyncInfo) {
	if syncInfo == nil {
		e.SyncInfo = nil
		return
	}

	e.SyncInfo = syncInfo
}

func (e *BaseEntity) GetSyncInfo() SyncInfo {
	if e.SyncInfo == nil {
		return nil
	}
	return e.SyncInfo
}

func (e *BaseEntity) ResetJoinMeta() {}

func (e *BaseEntity) SetJoinTS(ts types.Timestamp) {
	e.JoinTS = ts
}
