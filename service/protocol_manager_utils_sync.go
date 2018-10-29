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
)

func (pm *BaseProtocolManager) ForceSyncCycle() time.Duration {
	randNum := rand.Intn(pm.maxSyncRandomSeconds-pm.minSyncRandomSeconds) + pm.minSyncRandomSeconds

	return time.Duration(randNum) * time.Second
}

func (pm *BaseProtocolManager) QuitSync() chan struct{} {
	return pm.quitSync
}

func (pm *BaseProtocolManager) SetQuitSync(quitSync chan struct{}) {
	pm.quitSync = quitSync
}

func (pm *BaseProtocolManager) SyncWG() *sync.WaitGroup {
	return pm.syncWG
}

func (pm *BaseProtocolManager) Sync(peer *PttPeer) error {
	return nil
}
