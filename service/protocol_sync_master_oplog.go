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

type SyncMasterOplog struct {
	Oplogs []*BaseOplog `json:"O"`
}

func (pm *BaseProtocolManager) SyncPendingMasterOplog(peer *PttPeer) error {
	return pm.SyncPendingOplog(peer, pm.SetMasterDB, pm.HandleFailedMasterOplog, SyncPendingMasterOplogMsg)
}

func (pm *BaseProtocolManager) ForceSyncMasterMerkle() (bool, error) {
	err := pm.masterMerkle.TryForceSync(pm)
	if err != nil {
		return false, err
	}

	return true, nil
}
