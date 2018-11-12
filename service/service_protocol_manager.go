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

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
)

/*
ServiceProtocolManager manage service-level operations.

ServiceProtocolManager includes peers-of-services and the corresponding entities.
both are dynamically allocated / deallocated.

When there is a new peer: have all the existing entities trying to register the peer.
When a peer disappear: have all the existing entities trying to unregister the peer.

When there is a new entity: trying to register all the peers.
When a peer disappear: trying to unregister all the peers.
*/
type ServiceProtocolManager interface {
	Prestart() error
	Start() error
	Stop() error

	// entities
	Entities() map[types.PttID]Entity
	Entity(id *types.PttID) Entity

	RegisterEntity(id *types.PttID, e Entity) error
	UnregisterEntity(id *types.PttID) error

	Ptt() Ptt
	Service() Service

	GetDBLock() *types.LockMap

	NewEmptyEntity() Entity
}

type BaseServiceProtocolManager struct {
	lock     sync.RWMutex
	entities map[types.PttID]Entity

	noMorePeers chan struct{}

	ptt     Ptt
	service Service

	dbLock    *types.LockMap
	dbLogLock *types.LockMap

	lockJoinRequest sync.RWMutex
	joinRequests    map[common.Address]*JoinRequest
}

func NewBaseServiceProtocolManager(ptt Ptt, service Service) (*BaseServiceProtocolManager, error) {
	dbLock, err := types.NewLockMap(SleepTimeLock)
	if err != nil {
		return nil, err
	}

	dbLogLock, err := types.NewLockMap(SleepTimeLock)
	if err != nil {
		return nil, err
	}

	spm := &BaseServiceProtocolManager{
		entities: make(map[types.PttID]Entity),

		noMorePeers: ptt.NoMorePeers(),
		ptt:         ptt,

		service: service,

		dbLock:    dbLock,
		dbLogLock: dbLogLock,
	}

	return spm, nil
}

func (spm *BaseServiceProtocolManager) GetDBLock() *types.LockMap {
	return spm.dbLock
}

func (spm *BaseServiceProtocolManager) GetDBLogLock() *types.LockMap {
	return spm.dbLogLock
}

func (spm *BaseServiceProtocolManager) Prestart() error {
	return spm.PrestartEntities()
}

func (spm *BaseServiceProtocolManager) Start() error {
	return spm.StartEntities()
}

func (spm *BaseServiceProtocolManager) Stop() error {
	return spm.StopEntities()
}

func (spm *BaseServiceProtocolManager) Ptt() Ptt {
	return spm.ptt
}

func (spm *BaseServiceProtocolManager) Service() Service {
	return spm.service
}

func (spm *BaseServiceProtocolManager) NewEmptyEntity() Entity {
	return nil
}
