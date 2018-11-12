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

func (pm *BaseProtocolManager) handleRevokeOpKeyLog(oplog *BaseOplog, info *ProcessOpKeyInfo) ([]*BaseOplog, error) {
	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	toBroadcastLogs, err := pm.HandleDeleteObjectLog(
		oplog, info,
		opKey, nil,
		pm.SetOpKeyDB, nil, pm.postdeleteOpKey, pm.updateDeleteOpKeyInfo)
	if err != nil {
		return nil, err
	}

	return toBroadcastLogs, nil
}

func (pm *BaseProtocolManager) handlePendingRevokeOpKeyLog(oplog *BaseOplog, info *ProcessOpKeyInfo) ([]*BaseOplog, error) {

	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	return pm.HandlePendingDeleteObjectLog(
		oplog, info, opKey, nil, pm.SetOpKeyDB, nil, pm.setPendingDeleteOpKeySyncInfo, pm.updateDeleteOpKeyInfo)
}

func (pm *BaseProtocolManager) setNewestRevokeOpKeyLog(oplog *BaseOplog) (types.Bool, error) {
	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	return pm.SetNewestDeleteObjectLog(oplog, opKey)
}

func (pm *BaseProtocolManager) handleFailedRevokeOpKeyLog(oplog *BaseOplog) error {
	opKey := NewEmptyOpKey()
	pm.SetOpKeyObjDB(opKey)

	return pm.HandleFailedDeleteObjectLog(oplog, opKey)
}

func (pm *BaseProtocolManager) updateDeleteOpKeyInfo(theOpKey Object, oplog *BaseOplog, theInfo ProcessInfo) error {

	info, ok := theInfo.(*ProcessOpKeyInfo)
	if !ok {
		return ErrInvalidData
	}

	info.DeleteOpKeyInfo[*oplog.ObjID] = oplog

	return nil
}
