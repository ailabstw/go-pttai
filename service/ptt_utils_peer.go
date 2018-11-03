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
	mrand "math/rand"
	"reflect"
	"sync"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

/**********
 * Peer
 **********/

/*
NewPeer inits PttPeer
*/
func (p *BasePtt) NewPeer(version uint, peer *p2p.Peer, rw p2p.MsgReadWriter) (*PttPeer, error) {
	meteredMsgReadWriter, err := NewBaseMeteredMsgReadWriter(rw, version)
	if err != nil {
		return nil, err
	}
	return NewPttPeer(version, peer, meteredMsgReadWriter, p)
}

/*
HandlePeer handles peer
	1. Basic handshake
	2. AddNewPeer (defer RemovePeer)
	3. init read/write
	4. for-loop handle-message
*/
func (p *BasePtt) HandlePeer(peer *PttPeer) error {
	log.Debug("HandlePeer: start", "peer", peer)
	defer log.Debug("HandlePeer: done", "peer", peer)

	// 1. basic handshake
	err := peer.Handshake(p.networkID)
	if err != nil {
		return err
	}

	// 2. add new peer (defer remove-peer)
	err = p.AddNewPeer(peer)
	if err != nil {
		return err
	}
	defer p.RemovePeer(peer, false)

	// 3. init read-write
	p.RWInit(peer, peer.Version())

	// 4. for-loop handle-message
	log.Debug("HandlePeer: to for-loop", "peer", peer)
	for {
		err = p.HandleMessageWrapper(peer)
		if err != nil {
			log.Error("HandlePeer: message handling failed", "e", err)
			break
		}
	}
	log.Debug("HandlePeer: after for-loop", "peer", peer, "e", err)

	return err
}

/*
AddPeer adds a new peer. expected no user-id.
	1. validate peer as random.
	2. set peer type as random.
	3. check dial-entity
	4. if there is a corresponding entity for dial: identify peer.
*/
func (p *BasePtt) AddNewPeer(peer *PttPeer) error {
	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	// 1. validate peer as random.
	err := p.ValidatePeer(peer.GetID(), peer.UserID, PeerTypeRandom, true)
	if err != nil {
		return err
	}

	// 2. set peer type as random.
	err = p.SetPeerType(peer, PeerTypeRandom, false, true)
	if err != nil {
		return err
	}

	err = p.CheckDialEntityAndIdentifyPeer(peer)
	if err != nil {
		return err
	}

	return nil
}

func (p *BasePtt) FinishIdentifyPeer(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	if peer.UserID == nil {
		return ErrPeerUserID
	}

	peerType, err := p.determinePeerTypeFromAllEntities(peer, true)
	if err != nil {
		return err
	}

	return p.SetupPeer(peer, peerType, true)
}

func (p *BasePtt) determinePeerTypeFromAllEntities(peer *PttPeer, isLocked bool) (PeerType, error) {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	p.entityLock.RLock()
	defer p.entityLock.RUnlock()

	// me
	if p.myEntity != nil && p.myEntity.MyPM().IsMyDevice(peer) {
		return PeerTypeMe, nil
	}

	// important
	var pm ProtocolManager
	for _, entity := range p.entities {
		pm = entity.PM()
		if pm.IsImportantPeer(peer) {
			return PeerTypeImportant, nil
		}
	}

	// member
	for _, entity := range p.entities {
		pm = entity.PM()
		if pm.IsMemberPeer(peer) {
			return PeerTypeMember, nil
		}
	}

	// random
	return PeerTypeRandom, nil
}

/*
SetupPeer setup peer with known user-id and register to entities.
*/
func (p *BasePtt) SetupPeer(peer *PttPeer, peerType PeerType, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	if peer.UserID == nil {
		return ErrPeerUserID
	}

	err := p.addPeerKnownUserID(peer, peerType, true)
	if err != nil {
		return err
	}

	err = p.RegisterPeerToEntities(peer, true)
	if err != nil {
		return err
	}

	return nil
}

/*
AddPeerKnownUserID deals with peer already with user-id and the corresponding peer-type.
	1. validate-peer.
	2. setup peer.
*/
func (p *BasePtt) addPeerKnownUserID(peer *PttPeer, peerType PeerType, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	err := p.ValidatePeer(peer.GetID(), peer.UserID, peerType, true)
	if err != nil {
		return err
	}

	return p.SetPeerType(peer, peerType, false, true)
}

/*
RemovePeer removes peer
	1. get reigsteredPeer
	2. unregister peer from entities
	3. unset peer type
	4. disconnect
*/
func (p *BasePtt) RemovePeer(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	peerID := peer.GetID()

	registeredPeer := p.GetPeer(peerID, true)
	if registeredPeer == nil {
		return nil
	}

	err := p.UnregisterPeerFromEntities(peer, true)
	if err != nil {
		log.Error("unable to unregister peer from entities", "peer", peer, "e", err)
	}

	err = p.UnsetPeerType(registeredPeer, true)
	if err != nil {
		log.Error("unable to remove peer", "peer", peer, "e", err)
	}

	peer.Peer.Disconnect(p2p.DiscUselessPeer)

	return err
}

/*
ValidatePeer validates peer
	1. no need to do anything with my device
	2. check repeated user-id
	3. check
*/
func (p *BasePtt) ValidatePeer(nodeID *discover.NodeID, userID *types.PttID, peerType PeerType, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	// no need to do anything with peer-type-me
	if peerType == PeerTypeMe {
		return nil
	}

	// check repeated user-id
	if userID != nil {
		origNodeID, ok := p.userPeerMap[*userID]
		if ok && !reflect.DeepEqual(origNodeID, nodeID) {
			return ErrAlreadyRegistered
		}
	}

	// check max-peers
	lenMyPeers := len(p.myPeers)
	lenImportantPeers := len(p.importantPeers)
	lenMemberPeers := len(p.memberPeers)
	lenRandomPeers := len(p.randomPeers)

	if peerType == PeerTypeImportant && lenImportantPeers >= p.config.MaxImportantPeers {
		return p2p.DiscTooManyPeers
	}

	if peerType == PeerTypeMember && lenMemberPeers >= p.config.MaxMemberPeers {
		return p2p.DiscTooManyPeers
	}

	if peerType == PeerTypeRandom && lenRandomPeers >= p.config.MaxRandomPeers {
		return p2p.DiscTooManyPeers
	}

	lenPeers := lenMyPeers + lenImportantPeers + lenMemberPeers + lenRandomPeers
	if lenPeers >= p.config.MaxPeers {
		err := p.dropAnyPeer(peerType, true)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
SetPeerType sets the peer to the new peer-type and set in ptt peer-map.
*/
func (p *BasePtt) SetPeerType(peer *PttPeer, peerType PeerType, isForce bool, isLocked bool) error {
	if !isLocked {
		peer.LockPeerType.Lock()
		defer peer.LockPeerType.Unlock()

		p.peerLock.Lock()
		defer p.peerLock.Unlock()

	}

	origPeerType := peer.PeerType

	if !isForce && origPeerType >= peerType {
		return nil
	}

	peer.PeerType = peerType

	switch origPeerType {
	case PeerTypeMe:
		delete(p.myPeers, peer.ID())
	case PeerTypeImportant:
		delete(p.importantPeers, peer.ID())
	case PeerTypeMember:
		delete(p.memberPeers, peer.ID())
	case PeerTypeRandom:
		delete(p.randomPeers, peer.ID())
	}

	switch peerType {
	case PeerTypeMe:
		p.myPeers[peer.ID()] = peer
	case PeerTypeImportant:
		p.importantPeers[peer.ID()] = peer
	case PeerTypeMember:
		p.memberPeers[peer.ID()] = peer
	case PeerTypeRandom:
		p.randomPeers[peer.ID()] = peer
	}

	if peer.UserID != nil {
		p.userPeerMap[*peer.UserID] = peer.GetID()
	}

	return nil
}

/*
UnsetPeerType unsets the peer from the ptt peer-map.
*/
func (p *BasePtt) UnsetPeerType(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	peerID := peer.ID()
	peerType := peer.PeerType

	switch peerType {
	case PeerTypeMe:
		_, ok := p.myPeers[peerID]
		if !ok {
			return ErrNotRegistered
		}
		delete(p.myPeers, peerID)
	case PeerTypeImportant:
		_, ok := p.importantPeers[peerID]
		if !ok {
			return ErrNotRegistered
		}
		delete(p.importantPeers, peerID)
	case PeerTypeMember:
		_, ok := p.memberPeers[peerID]
		if !ok {
			return ErrNotRegistered
		}
		delete(p.memberPeers, peerID)
	case PeerTypeRandom:
		_, ok := p.randomPeers[peerID]
		if !ok {
			return ErrNotRegistered
		}
		delete(p.randomPeers, peerID)
	}

	return nil
}

/*
RegisterPeerToEntities registers peer to all the existing entities (register-peer-to-ptt is already done in CheckPeerType / SetPeerType)
	register to all the existing entities.
*/
func (p *BasePtt) RegisterPeerToEntities(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	log.Debug("RegisterPeerToEntities: start", "peer", peer)

	// register to all the existing entities.
	p.entityLock.RLock()
	defer p.entityLock.RUnlock()

	var pm ProtocolManager
	var err error
	var fitPeerType PeerType
	for _, entity := range p.entities {
		pm = entity.PM()
		fitPeerType = pm.GetPeerType(peer)

		if fitPeerType < PeerTypeMember {
			continue
		}

		err = pm.RegisterPeer(peer)
		if err != nil {
			log.Warn("RegisterPeerToEntities: unable to register peer to entity", "peer", peer, "entity", entity.Name(), "e", err)
		}
	}

	log.Debug("RegisterPeerToEntities: done", "peer", peer)

	return nil
}

/*
UnregisterPeerFromEntities unregisters the peer from all the existing entities.
*/
func (p *BasePtt) UnregisterPeerFromEntities(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	log.Debug("UnregisterPeerFromEntities: start", "peer", peer)

	p.entityLock.RLock()
	defer p.entityLock.RUnlock()

	var pm ProtocolManager
	var err error
	for _, entity := range p.entities {
		pm = entity.PM()

		err = pm.UnregisterPeer(peer)
		if err != nil {
			log.Warn("UnregisterPeerFromoEntities: unable to unregister peer from entity", "peer", peer, "entity", entity.Name(), "e", err)
		}
		// peer.RegisterEntity(goEntity, fitPeerType)
	}

	log.Debug("UnregisterPeerFromEntities: done", "peer", peer)

	return nil
}

/*
GetPeer gets specific peer
*/
func (p *BasePtt) GetPeer(id *discover.NodeID, isLocked bool) *PttPeer {
	if !isLocked {
		p.peerLock.RLock()
		defer p.peerLock.RUnlock()
	}

	peer := p.myPeers[*id]
	if peer != nil {
		return peer
	}

	peer = p.importantPeers[*id]
	if peer != nil {
		return peer
	}

	peer = p.memberPeers[*id]
	if peer != nil {
		return peer
	}

	peer = p.randomPeers[*id]
	if peer != nil {
		return peer
	}

	return nil
}

/*
DropAnyPeer drops any peers at most with the peerType.
*/
func (p *BasePtt) dropAnyPeer(peerType PeerType, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	log.Debug("dropAnyPeer: start", "peerType", peerType)
	if len(p.randomPeers) != 0 {
		return p.dropAnyPeerCore(p.randomPeers)
	}
	if peerType == PeerTypeRandom {
		return p2p.DiscTooManyPeers
	}

	if len(p.memberPeers) != 0 {
		return p.dropAnyPeerCore(p.memberPeers)
	}
	if peerType == PeerTypeMember {
		return p2p.DiscTooManyPeers
	}

	if len(p.importantPeers) != 0 {
		return p.dropAnyPeerCore(p.importantPeers)
	}

	if peerType == PeerTypeImportant {
		return p2p.DiscTooManyPeers
	}

	return nil
}

func (p *BasePtt) dropAnyPeerCore(peers map[discover.NodeID]*PttPeer) error {
	randIdx := mrand.Intn(len(peers))

	i := 0
	for eachPeerID, _ := range peers {
		if i == randIdx {
			node := &discover.Node{ID: eachPeerID}
			p.server.RemovePeer(node)
			break
		}

		i++
	}

	return nil
}

/**********
 * Dail
 **********/

func (p *BasePtt) AddDial(nodeID *discover.NodeID, opKey *common.Address) error {
	peer := p.GetPeer(nodeID, false)

	if peer != nil && peer.UserID != nil {
		return nil
	}

	err := p.dialHist.Add(nodeID, opKey)
	if err != nil {
		return err
	}

	if peer != nil {
		return p.CheckDialEntityAndIdentifyPeer(peer)
	}

	p.Server().AddPeer(&discover.Node{ID: *nodeID})

	return nil
}

func (p *BasePtt) CheckDialEntityAndIdentifyPeer(peer *PttPeer) error {
	// 1. check dial-entity
	entity, err := p.checkDialEntity(peer)
	if err != nil {
		return err
	}

	// 2. identify peer
	if entity != nil {
		entity.PM().IdentifyPeer(peer)
		return nil
	}

	return nil
}

func (p *BasePtt) checkDialEntity(peer *PttPeer) (Entity, error) {
	dialInfo := p.dialHist.Get(peer.GetID())
	if dialInfo == nil {
		return nil, nil
	}

	p.lockOps.RLock()
	defer p.lockOps.RUnlock()

	entityID := p.ops[*dialInfo.OpKey]
	if entityID == nil {
		return nil, nil
	}

	p.entityLock.RLock()
	p.entityLock.RUnlock()

	entity := p.entities[*entityID]

	return entity, nil
}

/**********
 * Misc
 **********/

func (p *BasePtt) GetPeers() (map[discover.NodeID]*PttPeer, map[discover.NodeID]*PttPeer, map[discover.NodeID]*PttPeer, map[discover.NodeID]*PttPeer, *sync.RWMutex) {
	return p.myPeers, p.importantPeers, p.memberPeers, p.randomPeers, &p.peerLock
}

func randomPttPeers(peers []*PttPeer) []*PttPeer {
	newPeers := make([]*PttPeer, len(peers))
	perm := mrand.Perm(len(peers))
	for i, v := range perm {
		newPeers[v] = peers[i]
	}
	return newPeers
}
