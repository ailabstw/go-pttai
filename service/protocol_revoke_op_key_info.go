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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) RevokeOpKeyInfo(keyID *types.PttID) (bool, error) {

	opKey := NewEmptyKeyInfo()
	pm.SetOpKeyObjDB(opKey)

	opData := &OpKeyOpRevokeOpKey{}

	err := pm.DeleteObject(keyID, opKey, OpKeyOpTypeRevokeOpKey, opData, pm.NewOpKeyOplog, nil, pm.broadcastOpKeyOplogCore, pm.postdeleteOpKey)
	if err != nil {
		return false, err
	}

	return true, nil
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

/*
postdeleteOpKey deals with ops after deletingOpKey. Assuming obj already locked (in DeleteObject and DeleteObjectLogs).
*/
func (pm *BaseProtocolManager) postdeleteOpKey(id *types.PttID, oplog *BaseOplog, origObj Object, opData OpData) error {
	hash := keyInfoIDToHash(id)

	opKey, ok := origObj.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	opKey.CreateLogID = oplog.PreLogID

	err := opKey.Save(true)
	if err != nil {
		return err
	}

	log.Debug("DeleteOpKeyPostprocess: to RemoveOpKeyInfoFromHash")

	return pm.RemoveOpKeyInfoFromHash(hash, false, false, false)
}

func (k *KeyInfo) RemoveMeta() {
	k.Hash = nil
	k.Key = nil
	k.KeyBytes = nil
	k.PubKeyBytes = nil
	k.Extra = nil
	k.Count = 0
}
