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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
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

func (k *KeyInfo) RemoveMeta() {
	k.Hash = nil
	k.Key = nil
	k.KeyBytes = nil
	k.PubKeyBytes = nil
	k.Extra = nil
	k.Count = 0
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
 * create
 **********/

func (k *KeyInfo) UpdateCreateInfo(oplog *BaseOplog, theOpData OpData, theInfo ProcessInfo) error {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return ErrInvalidData
	}

	info.CreateOpKeyInfo[*oplog.ObjID] = oplog

	return nil
}

func (k *KeyInfo) UpdateCreateObject(theObj Object) error {
	obj, ok := theObj.(*KeyInfo)
	if !ok {
		return ErrInvalidObject
	}

	key, err := crypto.ToECDSA(obj.KeyBytes)
	if err != nil {
		return err
	}

	origDB, origDBLock := k.db, k.dbLock
	k.BaseObject = obj.BaseObject
	k.db = origDB
	k.dbLock = origDBLock

	k.Hash = obj.Hash
	k.Key = key
	k.KeyBytes = obj.KeyBytes
	k.PubKeyBytes = crypto.FromECDSAPub(&key.PublicKey)
	k.Extra = obj.Extra

	return nil
}

func (k *KeyInfo) NewObjWithOplog(oplog *BaseOplog, theOpData OpData) error {
	NewObjectWithOplog(k, oplog)
	return nil
}

/**********
 * delete
 **********/

func (k *KeyInfo) UpdateDeleteInfo(oplog *BaseOplog, theInfo ProcessInfo) error {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return ErrInvalidData
	}

	info.CreateOpKeyInfo[*oplog.ObjID] = oplog
	info.DeleteOpKeyInfo[*oplog.ObjID] = oplog

	return nil
}

func (k *KeyInfo) SetPendingDeleteSyncInfo(oplog *BaseOplog) error {

	syncInfo := EmptySyncKeyInfo()
	syncInfo.InitWithOplog(oplog)
	syncInfo.Status = types.StatusToDeleteStatus(syncInfo.Status)

	k.SyncInfo = syncInfo

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
