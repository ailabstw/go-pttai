// Copyright 2019 The go-pttai Authors
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
	"math/rand"
	"sync"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
)

func PMSync(pm ProtocolManager) error {
	var err error
	forceSyncTicker := time.NewTicker(pm.ForceSyncCycle())

	var peer *PttPeer
looping:
	for {
		select {
		case peer, ok := <-pm.NewPeerCh():
			if !ok {
				break looping
			}

			log.Debug("PMSync: NewPeerCh: start", "entity", pm.Entity().IDString(), "peer", peer)

			pm.SyncOpKeyOplog(peer, SyncOpKeyOplogMsg)
			err = pm.Sync(peer)
			log.Debug("PMSync: NewPeerCh: after pm.Sync", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
			if err != nil {
				log.Error("PMSync: unable to Sync after newPeer", "e", err, "peer", peer, "entity", pm.Entity().IDString())

				if err == p2p.ErrPeerShutdown {
					pm.UnregisterPeer(peer, false, true, false)
				}
			}
			log.Debug("PMSync: NewPeerCh: done", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
		case <-pm.ForceSync():
			log.Debug("PMSync: ForceSync: start", "entity", pm.Entity().IDString())
			peer, err = pmSyncPeer(pm)
			log.Debug("PMSync: ForceSync: after pmSyncPeer", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
			if err != nil {
				break looping
			}

			pm.SyncOpKeyOplog(peer, SyncOpKeyOplogMsg)
			log.Debug("PMSync: ForceSync: to Sync", "entity", pm.Entity().IDString(), "peer", peer)
			err = pm.Sync(peer)
			log.Debug("PMSync: ForceSync: after Sync", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
			if err != nil {
				log.Error("PMSync: unable to Sync after forceSync", "e", err, "peer", peer, "entity", pm.Entity().IDString())
				if err == p2p.ErrPeerShutdown {
					pm.UnregisterPeer(peer, false, true, false)
				}
			}
			log.Debug("PMSync: ForceSync: done", "entity", pm.Entity().IDString(), "peer", peer, "e", err)

		case <-forceSyncTicker.C:
			forceSyncTicker.Stop()
			forceSyncTicker = time.NewTicker(pm.ForceSyncCycle())

			log.Debug("PMSync: ticker: start", "entity", pm.Entity().IDString())

			peer, err = pmSyncPeer(pm)
			log.Debug("PMSync: ticker: after pmSyncPeer", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
			if err != nil {
				break looping
			}

			pm.SyncOpKeyOplog(peer, SyncOpKeyOplogMsg)
			log.Debug("PMSync: ticker: to Sync", "entity", pm.Entity().IDString(), "peer", peer)
			err = pm.Sync(peer)
			log.Debug("PMSync: ticker: after pm.Sync", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
			if err != nil {
				log.Error("PMSync: unable to Sync after ticker", "e", err, "peer", peer, "entity", pm.Entity().IDString())
				if err == p2p.ErrPeerShutdown {
					pm.UnregisterPeer(peer, false, true, false)
				}
			}
			log.Debug("PMSync: ticker: done", "entity", pm.Entity().IDString(), "peer", peer, "e", err)
		case <-pm.QuitSync():
			log.Debug("PMSync: QuitSync", "entity", pm.Entity().IDString())
			err = p2p.DiscQuitting
			break looping
		}
	}
	forceSyncTicker.Stop()

	return err
}

func pmSyncPeer(pm ProtocolManager) (*PttPeer, error) {
	peerList := pm.Peers().PeerList(false)
	if len(peerList) == 0 {
		pm.LoadPeers()
		return nil, nil
	}
	peer := RandomPeer(peerList)

	return peer, nil
}

func (pm *BaseProtocolManager) ForceSyncCycle() time.Duration {
	if pm.Ptt().GetMyEntity().GetStatus() < types.StatusAlive {
		return time.Duration(5) * time.Second
	}

	if pm.Entity().GetStatus() < types.StatusAlive {
		return time.Duration(5) * time.Second
	}
	randNum := rand.Intn(pm.maxSyncRandomSeconds-pm.minSyncRandomSeconds) + pm.minSyncRandomSeconds

	return time.Duration(randNum) * time.Second
}

func (pm *BaseProtocolManager) QuitSync() chan struct{} {
	return pm.quitSync
}

func (pm *BaseProtocolManager) ForceSync() chan struct{} {
	return pm.forceSync
}

func (pm *BaseProtocolManager) SyncWG() *sync.WaitGroup {
	return &pm.syncWG
}

func (pm *BaseProtocolManager) Sync(peer *PttPeer) error {
	return nil
}
