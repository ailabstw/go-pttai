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

type Ptt struct {
	config   *Config
	MyNodeID *discover.NodeID

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

	// p2p server
	server *p2p.Server

	// protocols
	protocols []p2p.Protocol

	// apis
	apis []rpc.API

	networkID uint32

	// misc
	ErrChan *types.Chan
}

func NewPtt(ctx *ServiceContext, cfg *Config, myNodeID *discover.NodeID) (*Ptt, error) {
	log.Debug("NewPtt: start")

	p := &Ptt{
		config: cfg,

		eventMux: new(event.TypeMux),

		NotifyNodeRestart: types.NewChan(1),
		NotifyNodeStop:    types.NewChan(1),

		// peer
		newPeerCh:   make(chan *PttPeer),
		noMorePeers: make(chan struct{}),

		// sync
		quitSync: make(chan struct{}),

		myPeers:        make(map[discover.NodeID]*PttPeer),
		importantPeers: make(map[discover.NodeID]*PttPeer),
		memberPeers:    make(map[discover.NodeID]*PttPeer),
		randomPeers:    make(map[discover.NodeID]*PttPeer),
	}

	log.Debug("NewPtt: done", "quitSync", p.quitSync)

	p.apis = p.PttAPIs()

	p.protocols = p.GenerateProtocols()

	return p, nil
}

func (p *Ptt) Protocols() []p2p.Protocol {
	return p.protocols
}

func (p *Ptt) APIs() []rpc.API {
	return p.apis
}

func (p *Ptt) Start(server *p2p.Server) error {
	p.server = server

	go p.SyncWrapper()

	return nil
}

func (p *Ptt) Stop() error {
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

func (p *Ptt) SyncWrapper() {
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

func (p *Ptt) RWInit(peer *PttPeer, version uint) {
	if rw, ok := peer.RW().(MeteredMsgReadWriter); ok {
		rw.Init(version)
	}
}
