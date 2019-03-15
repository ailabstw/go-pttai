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

			log.Debug("PMSync: NewPeerCh: start", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

			pm.SyncOpKeyOplog(peer, SyncOpKeyOplogMsg)
			err = pm.Sync(peer)
			log.Debug("PMSync: NewPeerCh: after pm.Sync", "entity", pm.Entity(), "peer", peer, "e", err)
			if err != nil {
				log.Error("unable to Sync after newPeer", "e", err, "peer", peer, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())

				if err == p2p.ErrPeerShutdown {
					pm.UnregisterPeer(peer, false, true, false)
				}
			}
		case <-pm.ForceSync():
			log.Debug("PMSync: ForceSync: start", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
			peer, err = pmSyncPeer(pm)
			log.Debug("PMSync: ForceSync: after pmSyncPeer", "entity", pm.Entity(), "peer", peer, "e", err)
			if err != nil {
				break looping
			}

			pm.SyncOpKeyOplog(peer, SyncOpKeyOplogMsg)
			log.Debug("PMSync: ForceSync: to Sync", "entity", pm.Entity(), "peer", peer)
			err = pm.Sync(peer)
			log.Debug("PMSync: ForceSync: after Sync", "entity", pm.Entity(), "peer", peer)
			if err != nil {
				log.Error("unable to Sync after forceSync", "e", err, "peer", peer, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
				if err == p2p.ErrPeerShutdown {
					pm.UnregisterPeer(peer, false, true, false)
				}
			}

		case <-forceSyncTicker.C:
			forceSyncTicker.Stop()
			forceSyncTicker = time.NewTicker(pm.ForceSyncCycle())

			peer, err = pmSyncPeer(pm)
			if err != nil {
				break looping
			}

			pm.SyncOpKeyOplog(peer, SyncOpKeyOplogMsg)
			err = pm.Sync(peer)
			if err != nil {
				log.Error("unable to Sync after forceSyncTicker", "e", err, "peer", peer, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
				if err == p2p.ErrPeerShutdown {
					pm.UnregisterPeer(peer, false, true, false)
				}
			}
		case <-pm.QuitSync():
			log.Info("PMSync: QuitSync", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
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

	validPeerList := make([]*PttPeer, 0, len(peerList))
	for _, peer := range peerList {
		if peer.IsReady {
			validPeerList = append(validPeerList, peer)
		}
	}
	peer := RandomPeer(validPeerList)

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
