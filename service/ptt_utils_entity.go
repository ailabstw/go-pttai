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

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
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

func (p *BasePtt) RegisterEntityPeerWithOtherUserID(e Entity, id *types.PttID, peerType PeerType, isLocked bool) error {

	if !isLocked {
		p.peerLock.RLock()
		defer p.peerLock.RUnlock()
	}

	myID := p.GetMyEntity().GetID()
	if reflect.DeepEqual(myID, id) {
		return nil
	}

	peer, err := p.GetPeerByUserID(id, true)
	if err != nil {
		return err
	}
	if peer == nil {
		return nil
	}

	return e.PM().RegisterPeer(peer, peerType)
}

func (p *BasePtt) RegisterEntity(e Entity, isLocked bool, isPeerLocked bool) error {
	if !isLocked {
		p.entityLock.Lock()
		defer p.entityLock.Unlock()
	}

	id := e.GetID()
	p.entities[*id] = e

	log.Debug("RegisterEntity: to registerEntityPeers")

	return p.registerEntityPeers(e, isPeerLocked)
}

func (p *BasePtt) registerEntityPeers(e Entity, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	log.Debug("registerEntityPeers: after lock")

	toMyPeers := make([]*PttPeer, 0)
	toImportantPeers := make([]*PttPeer, 0)
	toMemberPeers := make([]*PttPeer, 0)
	toPendingPeers := make([]*PttPeer, 0)

	pm := e.PM()

	// my-peers: always my-peer and register the entity
	var peer *PttPeer
	log.Debug("registerEntityPeers: to myPeers")
	for _, peer = range p.myPeers {
		pm.RegisterPeer(peer, PeerTypeMe)
	}

	// hub-peers
	log.Debug("registerEntityPeers: to hubPeers")
	for _, peer = range p.hubPeers {
		if pm.IsMyDevice(peer) {
			pm.RegisterPeer(peer, PeerTypeMe)
		} else if pm.IsImportantPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeImportant)
		} else if pm.IsMemberPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeMember)
		} else if pm.IsPendingPeer(peer) {
			pm.RegisterPendingPeer(peer)
		}
	}

	// important-peers
	toRemovePeers := make([]*discover.NodeID, 0)
	log.Debug("registerEntityPeers: to importantPeers")
	for _, peer = range p.importantPeers {
		if pm.IsMyDevice(peer) {
			log.Debug("registerEntityPeers: important-to-me", "peer", peer)
			pm.RegisterPeer(peer, PeerTypeMe)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsImportantPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeImportant)
		} else if pm.IsMemberPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeMember)
		} else if pm.IsPendingPeer(peer) {
			pm.RegisterPendingPeer(peer)
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.importantPeers, *nodeID)
	}

	// member-peers
	toRemovePeers = make([]*discover.NodeID, 0)
	log.Debug("registerEntityPeers: to memberPeers")
	for _, peer = range p.memberPeers {
		if pm.IsMyDevice(peer) {
			log.Debug("registerEntityPeers: member-to-me", "peer", peer)
			pm.RegisterPeer(peer, PeerTypeMe)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsImportantPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeImportant)
			toImportantPeers = append(toImportantPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsMemberPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeMember)
		} else if pm.IsPendingPeer(peer) {
			pm.RegisterPendingPeer(peer)
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.memberPeers, *nodeID)
	}

	// pending-peers
	toRemovePeers = make([]*discover.NodeID, 0)
	log.Debug("registerEntityPeers: to pendingPeers")
	for _, peer = range p.pendingPeers {
		if pm.IsMyDevice(peer) {
			pm.RegisterPeer(peer, PeerTypeMe)
			log.Debug("registerEntityPeers: pending-to-me", "peer", peer)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsImportantPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeImportant)
			toImportantPeers = append(toImportantPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsMemberPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeMember)
			toMemberPeers = append(toMemberPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsPendingPeer(peer) {
			pm.RegisterPendingPeer(peer)
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.memberPeers, *nodeID)
	}

	// random-peers
	toRemovePeers = make([]*discover.NodeID, 0)
	log.Debug("registerEntityPeers: to randomPeers", "randomPeers", len(p.randomPeers))
	for _, peer = range p.randomPeers {
		if pm.IsMyDevice(peer) {
			log.Debug("registerEntityPeers: random-to-me", "peer", peer)
			pm.RegisterPeer(peer, PeerTypeMe)
			toMyPeers = append(toMyPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsImportantPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeImportant)
			toImportantPeers = append(toImportantPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsMemberPeer(peer) {
			pm.RegisterPeer(peer, PeerTypeMember)
			toMemberPeers = append(toMemberPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		} else if pm.IsPendingPeer(peer) {
			pm.RegisterPeer(peer, PeerTypePending)
			toPendingPeers = append(toPendingPeers, peer)
			toRemovePeers = append(toRemovePeers, peer.GetID())
		}
	}
	for _, nodeID := range toRemovePeers {
		delete(p.randomPeers, *nodeID)
	}

	// to my-peers
	log.Debug("registerEntityPeers", "toMyPeers", len(toMyPeers))
	for _, peer = range toMyPeers {
		//id := peer.ID()
		p.SetPeerType(peer, PeerTypeMe, false, true)
		// p.myPeers[id] = peer
	}

	// to important-peers
	log.Debug("registerEntityPeers", "toImportantPeers", len(toImportantPeers))
	for _, peer = range toImportantPeers {
		//id := peer.ID()
		p.SetPeerType(peer, PeerTypeImportant, false, true)
		// p.importantPeers[id] = peer
	}

	// to member
	log.Debug("registerEntityPeers", "toMemberPeers", len(toMemberPeers))
	for _, peer = range toMemberPeers {
		//id := peer.ID()
		p.SetPeerType(peer, PeerTypeMember, false, true)
		//p.memberPeers[id] = peer
	}

	// to pending
	log.Debug("registerEntityPeers", "toPendingPeers", len(toPendingPeers))
	for _, peer = range toPendingPeers {
		//id := peer.ID()
		p.SetPeerType(peer, PeerTypePending, false, true)
		// p.pendingPeers[id] = peer
	}

	log.Debug("registerEntityPeers: done")

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

func (p *BasePtt) GetEntities() map[types.PttID]Entity {
	return p.entities
}
