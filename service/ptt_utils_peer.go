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

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

/**********
 * Peer
 **********/

type ErrWrapper struct {
	err error
}

func (p *Ptt) HandlePeer(peer *PttPeer) error {
	log.Debug("HandlePeer: start", "peer", peer)

	errWrapper := &ErrWrapper{}
	errWrapper.err = peer.Handshake(p.networkID)
	log.Debug("HandlePeer: after Handshake", "e", errWrapper.err)
	if errWrapper.err != nil {
		return errWrapper.err
	}
	defer log.Debug("HandlePeer: done", "err", errWrapper, "peer", peer)

	errWrapper.err = p.AddPeer(peer)
	if errWrapper.err != nil {
		return errWrapper.err
	}
	defer p.RemovePeer(peer)

	p.RWInit(peer, peer.Version())

	log.Debug("HandlePeer: to for-loop", "peer", peer)
	for {
		errWrapper.err = p.HandleMessageWrapper(peer)
		if errWrapper.err != nil {
			log.Error("HandlePeer: message handling failed", "e", errWrapper.err)
			break
		}
	}

	return errWrapper.err
}

/*
AddPeer adds peer.
*/
func (p *Ptt) AddPeer(peer *PttPeer) error {
	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	p.SetPeerType(peer, PeerTypeRandom, false, true)

	return p.ValidatePeer(peer, true)
}

/*
SetupPeer setup peer with known user-id and register to entities.
*/
func (p *Ptt) SetupPeer(peer *PttPeer) error {
	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	err := p.addPeerKnownUserID(peer, true)
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
AddPeerKnownUserID deals with peer already with user-id.
	1. check-peer-type
	2. validate-peer
*/
func (p *Ptt) addPeerKnownUserID(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	err := p.CheckPeerType(peer, true)
	if err != nil {
		return err
	}

	err = p.ValidatePeer(peer, true)
	log.Debug("addPeerKnownUserID: after ValidatePeer", "e", err)
	if err != nil {
		return err
	}

	return nil
}

func (p *Ptt) RemovePeer(peer *PttPeer) error {
	id := peer.ID()

	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	registeredPeer := p.GetPeer(&id, true)
	if registeredPeer == nil {
		return nil
	}

	err := p.UnregisterPeer(registeredPeer, true)
	if err != nil {
		log.Error("unable to remove peer", "id", id, "e", err)
	}

	peer.Peer.Disconnect(p2p.DiscUselessPeer)

	return err
}

func (p *Ptt) NewPeer(version uint, peer *p2p.Peer, rw p2p.MsgReadWriter) (*PttPeer, error) {
	meteredMsgReadWriter, err := NewBaseMeteredMsgReadWriter(rw, version)
	if err != nil {
		return nil, err
	}
	return NewPttPeer(version, peer, meteredMsgReadWriter, p)
}

/*
CheckPeerType goes through all the registered entities and check the corresponding peer-type.
*/
func (p *Ptt) CheckPeerType(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	return p.SetPeerType(peer, PeerTypeRandom, false, true)
}

/*
SetPeerType sets the peer to the new peer-type.
*/
func (p *Ptt) SetPeerType(peer *PttPeer, peerType PeerType, isForce bool, isLocked bool) error {
	if !isLocked {
		peer.LockPeerType.Lock()
		defer peer.LockPeerType.Unlock()

		p.peerLock.Lock()
		defer p.peerLock.Unlock()

	}

	if !isForce && peer.PeerType >= peerType {
		return nil
	}

	origPeerType := peer.PeerType
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

	return nil
}

func (p *Ptt) ValidatePeer(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	peerType := peer.PeerType

	// no need to do anything with peer-type-me
	if peerType == PeerTypeMe {
		return nil
	}

	// check repeated user-id
	for _, eachPeer := range p.importantPeers {
		if peer != eachPeer && reflect.DeepEqual(peer.UserID, eachPeer.UserID) {
			log.Error("ValidatePeer: already registered (important)", "peerID", peer.GetID(), "userID", peer.UserID, "exist-peerID", eachPeer.GetID())
			return ErrAlreadyRegistered
		}
	}

	for _, eachPeer := range p.memberPeers {
		if peer != eachPeer && reflect.DeepEqual(peer.UserID, eachPeer.UserID) {
			log.Error("ValidatePeer: already registered (member)", "peerID", peer.GetID(), "userID", peer.UserID, "exist-peerID", eachPeer.GetID())
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
		err := p.DropAnyPeer(peerType, true)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
RegisterPeer registers peer to all the existing entities (register-peer-to-ptt is already done in CheckPeerType / SetPeerType)
	register to all the existing entities.
*/
func (p *Ptt) RegisterPeerToEntities(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	peerID := peer.Peer.ID()

	log.Debug("RegisterPeer: start", "peerID", peerID, "peerType", peer.PeerType)

	return nil
}

func (p *Ptt) UnregisterPeer(peer *PttPeer, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	err := p.UnsetPeerType(peer, peer.PeerType, true)
	if err != nil {
		return err
	}

	peer.Peer.Disconnect(p2p.DiscUselessPeer)

	log.Debug("Unregister Peer: done", "peer", peer)

	return nil
}

func (p *Ptt) UnsetPeerType(peer *PttPeer, peerType PeerType, isLocked bool) error {
	peerID := peer.ID()

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

func (p *Ptt) ClosePeers() {
	p.peerLock.RLock()
	defer p.peerLock.RUnlock()

	for _, peer := range p.myPeers {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
		log.Debug("ClosePeers: disconnect", "peer", peer)
	}

	for _, peer := range p.importantPeers {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
		log.Debug("ClosePeers: disconnect", "peer", peer)
	}

	for _, peer := range p.memberPeers {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
		log.Debug("ClosePeers: disconnect", "peer", peer)
	}

	for _, peer := range p.randomPeers {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
		log.Debug("ClosePeers: disconnect", "peer", peer)
	}
}

func (p *Ptt) GetPeer(id *discover.NodeID, isLocked bool) *PttPeer {
	if !isLocked {
		p.peerLock.RLock()
		defer p.peerLock.RUnlock()
	}

	idVal := *id
	peer := p.myPeers[idVal]
	if peer != nil {
		return peer
	}

	peer = p.importantPeers[idVal]
	if peer != nil {
		return peer
	}

	peer = p.memberPeers[idVal]
	if peer != nil {
		return peer
	}

	peer = p.randomPeers[idVal]
	if peer != nil {
		return peer
	}

	return nil
}

/*
DropAnyPeer drops any peers at most with the peerType.
*/
func (p *Ptt) DropAnyPeer(peerType PeerType, isLocked bool) error {
	if !isLocked {
		p.peerLock.Lock()
		defer p.peerLock.Unlock()
	}

	log.Debug("DropAnyPeer: start", "peerType", peerType)
	if len(p.randomPeers) != 0 {
		return p.DropPeer(p.randomPeers)
	}
	if peerType == PeerTypeRandom {
		return p2p.DiscTooManyPeers
	}

	if len(p.memberPeers) != 0 {
		return p.DropPeer(p.memberPeers)
	}
	if peerType == PeerTypeMember {
		return p2p.DiscTooManyPeers
	}

	if len(p.importantPeers) != 0 {
		return p.DropPeer(p.importantPeers)
	}

	if peerType == PeerTypeImportant {
		return p2p.DiscTooManyPeers
	}

	return nil
}

func (p *Ptt) DropPeer(peers map[discover.NodeID]*PttPeer) error {
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
 * Misc
 **********/

func (p *Ptt) GetPeers() (map[discover.NodeID]*PttPeer, map[discover.NodeID]*PttPeer, map[discover.NodeID]*PttPeer, map[discover.NodeID]*PttPeer, *sync.RWMutex) {
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
