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
	"github.com/ailabstw/go-pttai/event"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/rpc"
)

/*
Ptt is the public-access version of Ptt.
*/
type Ptt interface {
	// event-mux

	ErrChan() *types.Chan

	// peers
	IdentifyPeer(entityID *types.PttID, quitSync chan struct{}, peer *PttPeer) (*IdentifyPeer, error)
	IdentifyPeerAck(challenge *types.Salt, peer *PttPeer) (*IdentifyPeerAck, error)
	HandleIdentifyPeerAck(entityID *types.PttID, data *IdentifyPeerAck, peer *PttPeer) error

	FinishIdentifyPeer(peer *PttPeer, isLocked bool) error

	// join
	LockJoins()
	UnlockJoins()

	AddJoinKey(hash *common.Address, entityID *types.PttID, isLocked bool) error
	RemoveJoinKey(hash *common.Address, entityID *types.PttID, isLocked bool) error

	// op
	LockOps()
	UnlockOps()

	AddOpKey(hash *common.Address, entityID *types.PttID, isLocked bool) error
	RemoveOpKey(hash *common.Address, entityID *types.PttID, isLocked bool) error

	// me
	MyEntity() MyEntity
	MyNodeID() *discover.NodeID

	SignKey() *KeyInfo

	// master
	CreateMasterOplog(raftIdx uint64, ts types.Timestamp, op OpType, data interface{}) (*MasterOplog, error)

	// data
	EncryptData(op OpType, data []byte, keyInfo *KeyInfo) ([]byte, error)
	DecryptData(ciphertext []byte, keyInfo *KeyInfo) (OpType, []byte, error)

	MarshalData(code CodeType, hash *common.Address, encData []byte) (*PttData, error)
	UnmarshalData(pttData *PttData) (CodeType, *common.Address, []byte, error)
}

type BasePtt struct {
	config *Config

	// event-mux
	eventMux *event.TypeMux

	notifyNodeRestart *types.Chan
	notifyNodeStop    *types.Chan
	errChan           *types.Chan

	// peers
	peerLock sync.RWMutex

	myPeers        map[discover.NodeID]*PttPeer
	importantPeers map[discover.NodeID]*PttPeer
	memberPeers    map[discover.NodeID]*PttPeer
	randomPeers    map[discover.NodeID]*PttPeer

	userPeerMap map[types.PttID]*discover.NodeID

	noMorePeers chan struct{}

	peerWG sync.WaitGroup

	dialHist *DialHistory

	// entities
	entityLock sync.RWMutex

	entities map[types.PttID]Entity

	// joins
	lockJoins sync.RWMutex
	joins     map[common.Address]*types.PttID

	lockConfirmJoin sync.RWMutex
	confirmJoins    map[string]*ConfirmJoin

	// ops
	lockOps sync.RWMutex
	ops     map[common.Address]*types.PttID

	// sync
	quitSync chan struct{}
	syncWG   sync.WaitGroup

	// services
	services map[string]Service

	// p2p server
	server *p2p.Server

	// protocols
	protocols []p2p.Protocol

	// apis
	apis []rpc.API

	// network-id
	networkID uint32

	// me
	myEntity PttMyEntity
	myNodeID *discover.NodeID // ptt knows only my-node-id

	meSub *event.TypeMuxSubscription

	// MeOplog
	meOplogMerkle *Merkle

	meOplogSub  *event.TypeMuxSubscription
	meOplogsSub *event.TypeMuxSubscription

	// MasterOplog
	masterOplogMerkle *Merkle
}

func NewPtt(ctx *ServiceContext, cfg *Config, myNodeID *discover.NodeID) (*BasePtt, error) {
	// init-service
	InitService(cfg.DataDir)

	p := &BasePtt{
		config: cfg,

		myNodeID: myNodeID,

		eventMux: new(event.TypeMux),

		notifyNodeRestart: types.NewChan(1),
		notifyNodeStop:    types.NewChan(1),

		// peer
		noMorePeers: make(chan struct{}),

		myPeers:        make(map[discover.NodeID]*PttPeer),
		importantPeers: make(map[discover.NodeID]*PttPeer),
		memberPeers:    make(map[discover.NodeID]*PttPeer),
		randomPeers:    make(map[discover.NodeID]*PttPeer),

		userPeerMap: make(map[types.PttID]*discover.NodeID),

		dialHist: NewDialHistory(),

		// sync
		quitSync: make(chan struct{}),

		// services
		services: make(map[string]Service),

		errChan: types.NewChan(1),
	}

	p.apis = p.PttAPIs()

	p.protocols = p.GenerateProtocols()

	return p, nil
}

/**********
 * PttService
 **********/

func (p *BasePtt) Protocols() []p2p.Protocol {
	return p.protocols
}

func (p *BasePtt) APIs() []rpc.API {
	return p.apis
}

func (p *BasePtt) Start(server *p2p.Server) error {
	p.server = server

	// Start services
	var err error
	successMap := make(map[string]Service)
	errMap := make(map[string]error)
	for name, service := range p.services {
		log.Info("Start: to start service", "name", name)
		err = service.Start()
		if err != nil {
			errMap[name] = err
			break
		}
	}

	if err != nil {
		for name, successService := range successMap {
			err = successService.Stop()
			if err != nil {
				errMap[name] = err
			}
		}
	}
	if len(errMap) != 0 {
		return errMapToErr(errMap)
	}

	return nil
}

func (p *BasePtt) Stop() error {
	close(p.quitSync)
	close(p.noMorePeers)

	// close all service-loop
	errMap := make(map[string]error)
	for name, service := range p.services {
		err := service.Stop()
		if err != nil {
			errMap[name] = err
		}
	}

	p.syncWG.Wait()

	p.peerWG.Wait()

	// remove ptt-level chan

	p.meOplogSub.Unsubscribe()
	p.meOplogsSub.Unsubscribe()

	p.eventMux.Stop()

	log.Debug("Stop: done")

	if len(errMap) != 0 {
		return errMapToErr(errMap)
	}

	return nil
}

/**********
 * RW
 **********/

func (p *BasePtt) RWInit(peer *PttPeer, version uint) {
	if rw, ok := peer.RW().(MeteredMsgReadWriter); ok {
		rw.Init(version)
	}
}

/**********
 * Service
 **********/

/*
RegisterService registers service into ptt.
*/
func (p *BasePtt) RegisterService(service Service) error {
	log.Info("RegisterService", "name", service.Name())
	p.apis = append(p.apis, service.APIs()...)

	name := service.Name()

	p.services[name] = service

	log.Info("RegisterService: done", "name", service.Name())

	return nil
}

/**********
 * Me
 **********/

func (p *BasePtt) MyNodeID() *discover.NodeID {
	return p.myNodeID
}

func (p *BasePtt) SetMyEntity(myEntity PttMyEntity) error {
	var err error
	p.myEntity = myEntity

	p.meOplogMerkle, err = NewMerkle(DBMeOplogPrefix, DBMeMerkleOplogPrefix, myEntity.GetID(), dbOplog)
	if err != nil {
		return err
	}

	p.masterOplogMerkle, err = NewMerkle(DBMasterOplogPrefix, DBMasterMerkleOplogPrefix, myEntity.GetID(), dbOplog)
	if err != nil {
		return err
	}

	return nil
}

func (p *BasePtt) MyEntity() MyEntity {
	return p.myEntity
}

func (p *BasePtt) SignKey() *KeyInfo {
	return p.myEntity.SignKey()
}

/**********
 * Chan
 **********/

func (p *BasePtt) NotifyNodeRestart() *types.Chan {
	return p.notifyNodeRestart
}

func (p *BasePtt) NotifyNodeStop() *types.Chan {
	return p.notifyNodeStop
}

func (p *BasePtt) ErrChan() *types.Chan {
	return p.errChan
}

/**********
 * Server
 **********/

func (p *BasePtt) Server() *p2p.Server {
	return p.server
}

/**********
 * Peer
 **********/

func (p *BasePtt) NoMorePeers() chan struct{} {
	return p.noMorePeers
}
