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
	"github.com/ailabstw/go-pttai/p2p/discover"
)

func (p *BasePtt) getEntityFromHash(hash *common.Address, lock *sync.RWMutex, hashMap map[common.Address]*types.PttID) (Entity, error) {
	lock.RLock()
	defer lock.RUnlock()

	hashVal := *hash
	entityID, ok := hashMap[hashVal]
	if !ok {
		return nil, ErrInvalidData
	}
	idVal := *entityID
	entity, ok := p.entities[idVal]
	if !ok {
		return nil, ErrInvalidData
	}

	return entity, nil
}

func (p *BasePtt) RegisterEntity(e Entity, isLocked bool) error {
	if !isLocked {
		p.entityLock.Lock()
		defer p.entityLock.Unlock()
	}

	id := e.GetID()
	p.entities[*id] = e

	return p.registerEntityPeers(e)
}

func (p *BasePtt) registerEntityPeers(e Entity) error {
	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	toMyPeers := make([]*PttPeer, 0)
	toImportantPeers := make([]*PttPeer, 0)
	toMemberPeers := make([]*PttPeer, 0)

	// my-peers: always my-peer and register the entity
	for _, peer := range p.myPeers {
		e.PM().RegisterPeer(peer)
	}

	// important-peers
	toRemovePeers := make([]*discover.NodeID, 0)
	for nodeID, peer := range p.importantPeers {
		if e.PM().IsMyDevice(peer) {
			e.PM().RegisterPeer(peer)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, &nodeID)
		} else if e.PM().IsImportantPeer(peer) {
			e.PM().RegisterPeer(peer)
		} else if e.PM().IsMemberPeer(peer) {
			e.PM().RegisterPeer(peer)
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.importantPeers, *nodeID)
	}

	// member-peers
	toRemovePeers = make([]*discover.NodeID, 0)
	for nodeID, peer := range p.memberPeers {
		if e.PM().IsMyDevice(peer) {
			//e.PM().RegisterPeer(peer)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, &nodeID)
		} else if e.PM().IsImportantPeer(peer) {
			//e.PM().RegisterPeer(peer)
			toImportantPeers = append(toImportantPeers, peer)
			toRemovePeers = append(toRemovePeers, &nodeID)
		} else if e.PM().IsMemberPeer(peer) {
			//e.PM().RegisterPeer(peer)
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.memberPeers, *nodeID)
	}

	// random-peers
	toRemovePeers = make([]*discover.NodeID, 0)
	for nodeID, peer := range p.randomPeers {
		if e.PM().IsMyDevice(peer) {
			e.PM().RegisterPeer(peer)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, &nodeID)
		} else if e.PM().IsImportantPeer(peer) {
			e.PM().RegisterPeer(peer)
			toImportantPeers = append(toImportantPeers, peer)
			toRemovePeers = append(toRemovePeers, &nodeID)
		} else if e.PM().IsMemberPeer(peer) {
			e.PM().RegisterPeer(peer)
			toImportantPeers = append(toMemberPeers, peer)
			toRemovePeers = append(toRemovePeers, &nodeID)
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.randomPeers, *nodeID)
	}

	// to my-peers
	for _, peer := range toMyPeers {
		id := peer.ID()
		p.myPeers[id] = peer
	}

	// to important-peers
	for _, peer := range toImportantPeers {
		id := peer.ID()
		p.importantPeers[id] = peer
	}

	// to member
	for _, peer := range toMemberPeers {
		id := peer.ID()
		p.memberPeers[id] = peer
	}

	return nil
}

func (p *BasePtt) UnregisterEntity(e Entity, isLocked bool) error {
	if !isLocked {
		p.entityLock.Lock()
		defer p.entityLock.Unlock()
	}

	id := e.GetID()
	delete(p.entities, *id)

	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	return nil
}
