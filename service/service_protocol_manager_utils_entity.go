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
	"github.com/ailabstw/go-pttai/log"
)

func (spm *BaseServiceProtocolManager) Entity(id *types.PttID) Entity {
	spm.lock.RLock()
	defer spm.lock.RUnlock()

	entity, ok := spm.entities[*id]
	if !ok {
		return nil
	}
	return entity
}

func (spm *BaseServiceProtocolManager) Entities() map[types.PttID]Entity {
	return spm.entities
}

/*
RegisterEntity register the entity to the service

need to do lock in the beginning because need to update entitiesByPeerID
*/
func (spm *BaseServiceProtocolManager) RegisterEntity(id *types.PttID, e Entity) error {
	spm.lock.Lock()
	defer spm.lock.Unlock()

	_, ok := spm.entities[*id]
	if ok {
		return ErrEntityAlreadyRegistered
	}

	spm.entities[*id] = e
	e.PM().SetNoMorePeers(spm.noMorePeers)

	return nil
}

func (spm *BaseServiceProtocolManager) UnregisterEntity(id *types.PttID) error {
	spm.lock.Lock()
	defer spm.lock.Unlock()

	_, ok := spm.entities[*id]
	if !ok {
		return ErrEntityNotRegistered
	}

	delete(spm.entities, *id)

	return nil
}

func (b *BaseServiceProtocolManager) StartEntities() error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	log.Info("StartEntities", "entities", b.entities)
	for _, entity := range b.entities {
		err := entity.Start()
		if err != nil {
			return err
		}
	}

	return nil

}

func (b *BaseServiceProtocolManager) StopEntities() error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, entity := range b.entities {
		err := entity.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}
