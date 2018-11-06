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
	"github.com/ailabstw/go-pttai/log"
)

// Save
// Delete
// GetByID
// GetNewObjByID

// SetUpdateTS
// GetUpdateTS

/**********
 * data
 **********/

/*
GetBlockInfo implements Object method
*/
func (k *KeyInfo) GetBlockInfo() BlockInfo {
	return nil
}

/*
RemoveBlock implements Object method
*/
func (k *KeyInfo) RemoveBlock(blockInfo BlockInfo, info ProcessInfo, isRemoveDB bool) error {
	return nil
}

/**********
 * Sync Info
 **********/

func (k *KeyInfo) GetSyncInfo() SyncInfo {
	return k.SyncInfo
}

func (k *KeyInfo) SetSyncInfo(theSyncInfo SyncInfo) error {
	syncInfo, ok := theSyncInfo.(*SyncKeyInfo)
	if !ok {
		return ErrInvalidSyncInfo
	}

	k.SyncInfo = syncInfo

	return nil
}

func (k *KeyInfo) RemoveSyncInfo(oplog *BaseOplog, theOpData OpData, syncInfo SyncInfo, info ProcessInfo) error {
	return nil
}

/**********
 * SetObjDB
 **********/

func (pm *BaseProtocolManager) SetOpKeyObjDB(opKey *KeyInfo) {
	opKey.SetEntityID(pm.Entity().GetID())
	log.Debug("SetOpKeyObjDB: to SetDB", "dblock", pm.DBObjLock())
	opKey.SetDB(pm.DBOpKeyInfo(), pm.DBObjLock(), DBOpKeyPrefix)
}
