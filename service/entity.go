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

/*
Entity is the fundamental interface of entities.
The reason why entity is not an object is because:
	1. Creating an entity does not require the consensus from other nodes.
	2. We may recover a deleted entity.
*/
type Entity interface {
	Start() error
	Stop() error

	GetUpdateTS() types.Timestamp
	SetUpdateTS(ts types.Timestamp)

	// implemented in BaseEntity
	GetID() *types.PttID
	SetID(id *types.PttID)

	GetCreateTS() types.Timestamp
	GetCreatorID() *types.PttID

	GetUpdaterID() *types.PttID
	SetUpdaterID(id *types.PttID)

	GetLogID() *types.PttID
	SetLogID(id *types.PttID)

	GetUpdateLogID() *types.PttID
	SetUpdateLogID(id *types.PttID)

	GetStatus() types.Status
	SetStatus(status types.Status)

	GetOwnerID() *types.PttID
	SetOwnerID(id *types.PttID)

	PM() ProtocolManager
	Ptt() Ptt
	Service() Service

	Name() string
	SetName(name string)

	DB() *pttdb.LDBBatch
}

type BaseEntity struct {
	V         types.Version
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`

	LogID       *types.PttID `json:"l,omitempty"`
	UpdateLogID *types.PttID `json:"u,omitempty"`

	Status types.Status `json:"S"`

	OwnerID *types.PttID `json:"o,omitempty"`

	pm      ProtocolManager
	name    string
	ptt     Ptt
	service Service

	db *pttdb.LDBBatch
}

func NewBaseEntity(id *types.PttID, createTS types.Timestamp, creatorID *types.PttID, status types.Status, ownerID *types.PttID, db *pttdb.LDBBatch) *BaseEntity {

	e := &BaseEntity{
		V:         types.CurrentVersion,
		ID:        id,
		CreateTS:  createTS,
		CreatorID: creatorID,
		UpdaterID: creatorID,
		Status:    status,
		OwnerID:   ownerID,
		db:        db,
	}

	return e
}

func (e *BaseEntity) Init(pm ProtocolManager, name string, ptt Ptt, service Service) {
	e.pm = pm
	e.name = ProtocolName
	e.ptt = ptt
	e.service = service
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

func (e *BaseEntity) GetOwnerID() *types.PttID {
	return e.OwnerID
}

func (e *BaseEntity) SetOwnerID(id *types.PttID) {
	e.OwnerID = id
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

func (e *BaseEntity) DB() *pttdb.LDBBatch {
	return e.db
}
