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
	MyNodeID() *discover.NodeID
}

type BasePtt struct {
	config   *Config
	myNodeID *discover.NodeID // ptt knows only my-node-id

	eventMux *event.TypeMux

	NotifyNodeRestart *types.Chan `json:"-"`
	NotifyNodeStop    *types.Chan `json:"-"`

	// peers
	peerLock sync.RWMutex

	myPeers        map[discover.NodeID]*PttPeer
	importantPeers map[discover.NodeID]*PttPeer
	memberPeers    map[discover.NodeID]*PttPeer
	randomPeers    map[discover.NodeID]*PttPeer

	newPeerCh   chan *PttPeer
	noMorePeers chan struct{}

	peerWG sync.WaitGroup

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

	networkID uint32
}

func NewPtt(ctx *ServiceContext, cfg *Config, myNodeID *discover.NodeID) (*BasePtt, error) {
	// init-service
	InitService(cfg.DataDir)

	p := &BasePtt{
		config: cfg,

		myNodeID: myNodeID,

		eventMux: new(event.TypeMux),

		NotifyNodeRestart: types.NewChan(1),
		NotifyNodeStop:    types.NewChan(1),

		// peer
		newPeerCh:   make(chan *PttPeer),
		noMorePeers: make(chan struct{}),

		myPeers:        make(map[discover.NodeID]*PttPeer),
		importantPeers: make(map[discover.NodeID]*PttPeer),
		memberPeers:    make(map[discover.NodeID]*PttPeer),
		randomPeers:    make(map[discover.NodeID]*PttPeer),

		// sync
		quitSync: make(chan struct{}),

		// services
		services: make(map[string]Service),
	}

	p.apis = p.PttAPIs()

	p.protocols = p.GenerateProtocols()

	return p, nil
}

func (p *BasePtt) Protocols() []p2p.Protocol {
	return p.protocols
}

func (p *BasePtt) APIs() []rpc.API {
	return p.apis
}

func (p *BasePtt) Start(server *p2p.Server) error {
	p.server = server

	go p.SyncWrapper()

	return nil
}

func (p *BasePtt) Stop() error {
	close(p.quitSync)
	close(p.noMorePeers)

	p.syncWG.Wait()

	// close peers
	p.ClosePeers()

	p.peerWG.Wait()

	p.eventMux.Stop()

	log.Debug("Stop: done")

	return nil
}

/**********
 * Sync
 **********/

func (p *BasePtt) SyncWrapper() {
	log.Debug("ptt.SyncWrapper: start")
loop:
	for {
		select {
		case _, ok := <-p.newPeerCh:
			if !ok {
				break loop
			}
		case <-p.quitSync:
			break loop
		}
	}

	log.Debug("ptt.SyncWrapper: done")
}

/**********
 * RW
 **********/

func (p *BasePtt) RWInit(peer *PttPeer, version uint) {
	if rw, ok := peer.RW().(MeteredMsgReadWriter); ok {
		rw.Init(version)
	}
}

/*
RegisterService
*/
func (p *BasePtt) RegisterService(service Service) error {
	log.Info("RegisterService", "name", service.Name())
	p.apis = append(p.apis, service.APIs()...)

	name := service.Name()

	p.services[name] = service

	log.Info("RegisterService: done", "name", service.Name())

	return nil
}

func (p *BasePtt) MyNodeID() *discover.NodeID {
	return p.myNodeID
}
