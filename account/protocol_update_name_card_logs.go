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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleUpdateNameCardLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) ([]*pkgservice.BaseOplog, error) {
	obj := NewEmptyNameCard()
	pm.SetNameCardDB(obj)

	opData := &UserOpUpdateNameCard{}

	return pm.HandleUpdateObjectLog(
		oplog,
		opData,

		obj,

		info,

		pm.userOplogMerkle,

		pm.syncNameCardInfoFromOplog,

		pm.SetUserDB,
		nil,

		nil,

		pm.updateUpdateNameCardInfo,
	)
}

func (pm *ProtocolManager) handlePendingUpdateNameCardLogs(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) (types.Bool, []*pkgservice.BaseOplog, error) {
	obj := NewEmptyNameCard()
	pm.SetNameCardDB(obj)

	opData := &UserOpUpdateNameCard{}

	return pm.HandlePendingUpdateObjectLog(
		oplog,
		opData,

		obj,

		info,

		pm.userOplogMerkle,

		pm.syncNameCardInfoFromOplog,

		pm.SetUserDB,
		nil,

		nil,

		pm.updateUpdateNameCardInfo,
	)
}

func (pm *ProtocolManager) setNewestUpdateNameCardLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {
	obj := NewEmptyNameCard()
	pm.SetNameCardDB(obj)

	return pm.SetNewestUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedUpdateNameCardLog(oplog *pkgservice.BaseOplog) error {

	obj := NewEmptyNameCard()
	pm.SetNameCardDB(obj)

	return pm.HandleFailedUpdateObjectLog(oplog, obj)
}

func (pm *ProtocolManager) handleFailedValidUpdateNameCardLog(oplog *pkgservice.BaseOplog, info *ProcessUserInfo) error {

	obj := NewEmptyNameCard()
	pm.SetNameCardDB(obj)

	return pm.HandleFailedValidUpdateObjectLog(oplog, obj, info, pm.updateUpdateNameCardInfo)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) syncNameCardInfoFromOplog(oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	syncInfo := NewEmptySyncNameCardInfo()
	syncInfo.InitWithOplog(types.StatusInternalSync, oplog)

	return syncInfo, nil
}

func (pm *ProtocolManager) updateUpdateNameCardInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, opData pkgservice.OpData, origSyncInfo pkgservice.SyncInfo, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessUserInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.NameCardInfo[*oplog.ObjID] = oplog

	return nil
}
