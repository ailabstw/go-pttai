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

func (pm *BaseProtocolManager) RevokeOpKey(keyID *types.PttID) (bool, error) {

	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	opData := &OpKeyOpRevokeOpKey{}

	err := pm.DeleteObject(
		keyID,
		OpKeyOpTypeRevokeOpKey,

		opKey,
		opData,

		nil,

		pm.SetOpKeyDB,

		pm.NewOpKeyOplog,
		nil,
		pm.setPendingDeleteOpKeySyncInfo,
		pm.broadcastOpKeyOplogCore,
		pm.postdeleteOpKey,
	)

	log.Debug("RevokeOpKeyInfo: after DeleteObject", "e", err)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) setPendingDeleteOpKeySyncInfo(theOpKey Object, status types.Status, oplog *BaseOplog) error {

	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	syncInfo := &BaseSyncInfo{}
	syncInfo.InitWithOplog(status, oplog)

	opKey.SyncInfo = syncInfo

	return nil
}

/*
postdeleteOpKey deals with ops after deletingOpKey. Assuming obj already locked (in DeleteObject and DeleteObjectLogs).
*/
func (pm *BaseProtocolManager) postdeleteOpKey(
	id *types.PttID,

	oplog *BaseOplog,
	opData OpData,

	origObj Object,
	blockInfo *BlockInfo,
) error {

	opKey, ok := origObj.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	hash := keyInfoIDToHash(id)

	// update create-log-id
	opKey.CreateLogID = oplog.PreLogID

	err := opKey.Save(true)
	if err != nil {
		return err
	}

	log.Debug("postdeleteOpKey: to RemoveOpKeyInfoFromHash")

	return pm.RemoveOpKeyFromHash(hash, false, false, false)
}

/**********
 * KeyInfo
 **********/

func (k *KeyInfo) RemoveMeta() {
	k.Hash = nil
	k.Key = nil
	k.KeyBytes = nil
	k.PubKeyBytes = nil
	k.Extra = nil
	k.Count = 0
}
