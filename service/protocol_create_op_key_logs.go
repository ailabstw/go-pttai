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
)

func (pm *BaseProtocolManager) handleCreateOpKeyLog(oplog *BaseOplog, info *ProcessOpKeyInfo) ([]*BaseOplog, error) {

	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	opData := &OpKeyOpCreateOpKey{}

	return pm.HandleCreateObjectLog(
		oplog, opKey, opData, info,
		pm.existsInInfoCreateOpKey, pm.newOpKeyWithOplog, pm.postcreateOpKey, pm.updateCreateOpKeyInfo)
}

func (pm *BaseProtocolManager) handlePendingCreateOpKeyLog(oplog *BaseOplog, info *ProcessOpKeyInfo) (types.Bool, []*BaseOplog, error) {

	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	opData := &OpKeyOpCreateOpKey{}

	return pm.HandlePendingCreateObjectLog(oplog, opKey, opData, info, pm.existsInInfoCreateOpKey, pm.newOpKeyWithOplog, pm.postcreateOpKey, pm.updateCreateOpKeyInfo)
}

func (pm *BaseProtocolManager) setNewestCreateOpKeyLog(oplog *BaseOplog) (types.Bool, error) {
	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	return pm.SetNewestCreateObjectLog(oplog, opKey)
}

func (pm *BaseProtocolManager) handleFailedCreateOpKeyLog(oplog *BaseOplog) error {

	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	return pm.HandleFailedCreateObjectLog(oplog, opKey, pm.postfailedCreateOpKey)
}

/**********
 * Customize
 **********/

/***
 * handleCreateObject
 ***/

func (pm *BaseProtocolManager) existsInInfoCreateOpKey(oplog *BaseOplog, theInfo ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return false, ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.DeleteOpKeyInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *BaseProtocolManager) newOpKeyWithOplog(oplog *BaseOplog, theOpData OpData) Object {
	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)
	NewObjectWithOplog(opKey, oplog)

	return opKey
}

func (pm *BaseProtocolManager) updateCreateOpKeyInfo(obj Object, oplog *BaseOplog, theOpData OpData, theInfo ProcessInfo) error {
	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return ErrInvalidData
	}

	info.CreateOpKeyInfo[*oplog.ObjID] = oplog

	return nil
}

/***
 * handleFailedCreateObject
 ***/

func (pm *BaseProtocolManager) postfailedCreateOpKey(theOpKey Object, oplog *BaseOplog) error {
	opKey, ok := theOpKey.(*KeyInfo)
	if !ok {
		return ErrInvalidData
	}

	opKey.Delete(true)

	return pm.RemoveOpKeyFromHash(opKey.Hash, false, false, false)
}
