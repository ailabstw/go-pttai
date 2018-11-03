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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type Object interface {
	/**********
	 * Require obj implementation
	 **********/

	Save(isLocked bool) error
	Delete(isLocked bool) error

	GetByID(isLocked bool) error
	GetNewObjByID(id *types.PttID, isLocked bool) (Object, error)

	SetUpdateTS(ts types.Timestamp)
	GetUpdateTS() types.Timestamp

	// data
	GetBlockInfo() BlockInfo
	RemoveBlock(blockInfo BlockInfo, info ProcessInfo, isRemoveDB bool) error
	RemoveMeta()

	// sync-info
	GetSyncInfo() SyncInfo
	SetSyncInfo(syncInfo SyncInfo) error
	RemoveSyncInfo(oplog *BaseOplog, opData OpData, syncInfo SyncInfo, info ProcessInfo) error

	// create

	UpdateCreateInfo(oplog *BaseOplog, opData OpData, info ProcessInfo) error
	UpdateCreateObject(obj Object) error

	NewObjWithOplog(oplog *BaseOplog, opData OpData) error

	// delete

	UpdateDeleteInfo(oplog *BaseOplog, info ProcessInfo) error
	SetPendingDeleteSyncInfo(oplog *BaseOplog) error

	/**********
	 * implemented in BaseObject
	 **********/

	SetDB(db *pttdb.LDBBatch, dbLock *types.LockMap, dbPrefix []byte)
	Lock() error
	Unlock() error
	RLock() error
	RUnlock() error

	SetVersion(v types.Version)

	SetCreateTS(ts types.Timestamp)

	SetCreatorID(id *types.PttID)
	GetCreatorID() *types.PttID

	SetUpdaterID(id *types.PttID)

	SetID(id *types.PttID)
	GetID() *types.PttID

	SetLogID(id *types.PttID)
	GetLogID() *types.PttID

	SetUpdateLogID(id *types.PttID)
	GetUpdateLogID() *types.PttID

	SetStatus(status types.Status)
	GetStatus() types.Status

	SetEntityID(id *types.PttID)
	GetEntityID() *types.PttID
}

type BaseObject struct {
	V         types.Version
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`
	EntityID  *types.PttID    `json:"e,omitempty"`

	LogID       *types.PttID `json:"l,omitempty"`
	UpdateLogID *types.PttID `json:"u,omitempty"`

	Status types.Status `json:"S"`

	db       *pttdb.LDBBatch
	dbLock   *types.LockMap
	dbPrefix []byte
}

func NewObject(
	id *types.PttID,
	createTS types.Timestamp,
	creatorID *types.PttID,
	updaterID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	db *pttdb.LDBBatch,
	dbLock *types.LockMap,
) *BaseObject {
	return &BaseObject{
		V:         types.CurrentVersion,
		ID:        id,
		CreateTS:  createTS,
		CreatorID: creatorID,
		UpdaterID: updaterID,
		EntityID:  entityID,

		LogID: logID,

		Status: status,

		db:     db,
		dbLock: dbLock,
	}
}

func (o *BaseObject) SetDB(db *pttdb.LDBBatch, dbLock *types.LockMap, dbPrefix []byte) {
	o.db = db
	o.dbLock = dbLock
	o.dbPrefix = dbPrefix

	if o.EntityID != nil {
		o.dbPrefix = append(o.dbPrefix, o.EntityID[:]...)
	}
}

func (o *BaseObject) Lock() error {
	return o.dbLock.Lock(o.ID)
}

func (o *BaseObject) Unlock() error {
	return o.dbLock.Unlock(o.ID)
}

func (o *BaseObject) RLock() error {
	return o.dbLock.Lock(o.ID)
}

func (o *BaseObject) RUnlock() error {
	return o.dbLock.Unlock(o.ID)
}

func (o *BaseObject) SetVersion(v types.Version) {
	o.V = v
}

func (o *BaseObject) SetCreateTS(ts types.Timestamp) {
	o.CreateTS = ts
}

func (o *BaseObject) SetCreatorID(id *types.PttID) {
	o.CreatorID = id
}

func (o *BaseObject) GetCreatorID() *types.PttID {
	return o.CreatorID
}

func (o *BaseObject) SetUpdaterID(id *types.PttID) {
	o.UpdaterID = id
}

func (o *BaseObject) SetID(id *types.PttID) {
	o.ID = id
}

func (o *BaseObject) GetID() *types.PttID {
	return o.ID
}

func (o *BaseObject) SetLogID(id *types.PttID) {
	o.LogID = id
}

func (o *BaseObject) GetLogID() *types.PttID {
	return o.LogID
}

func (o *BaseObject) SetUpdateLogID(id *types.PttID) {
	o.UpdateLogID = id
}

func (o *BaseObject) GetUpdateLogID() *types.PttID {
	return o.UpdateLogID
}

func (o *BaseObject) SetStatus(status types.Status) {
	o.Status = status
}

func (o *BaseObject) GetStatus() types.Status {
	return o.Status
}

func (o *BaseObject) SetEntityID(id *types.PttID) {
	o.EntityID = id
}

func (o *BaseObject) GetEntityID() *types.PttID {
	return o.EntityID
}
