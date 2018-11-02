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

package me

import pkgservice "github.com/ailabstw/go-pttai/service"

// It's possible that we have multiple ids due to multi-device setup.
// Requiring per-entity-level oplog, not unique MeOplog / MasterOplog / PttOplog in ptt-layer

func (pm *ProtocolManager) SetPttDB(log *pkgservice.BaseOplog) {
	myID := pm.Entity().GetID()
	myPtt := pm.myPtt
	log.SetDB(myPtt.DBOplog(), myID, pkgservice.DBPttOplogPrefix, pkgservice.DBPttIdxOplogPrefix, pkgservice.DBPttMerkleOplogPrefix, pkgservice.DBPttLockMap)
}
