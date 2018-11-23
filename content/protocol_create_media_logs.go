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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleCreateMediaLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {

	return pm.HandleCreateMediaLogs(oplog, info, pm.existsInInfoCreateMedia, pm.updateCreateMediaInfo)
}

func (pm *ProtocolManager) handlePendingCreateMediaLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	return pm.HandlePendingCreateMediaLogs(oplog, info, pm.existsInInfoCreateMedia, pm.updateCreateMediaInfo)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) existsInInfoCreateMedia(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) (bool, error) {
	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return false, pkgservice.ErrInvalidData
	}

	objID := oplog.ObjID
	_, ok = info.CreateMediaInfo[*objID]
	if ok {
		return true, nil
	}

	return false, nil
}

func (pm *ProtocolManager) updateCreateMediaInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData, theInfo pkgservice.ProcessInfo) error {
	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	blockInfo := obj.GetBlockInfo()
	if blockInfo == nil {
		return pkgservice.ErrInvalidData
	}

	info.CreateMediaInfo[*oplog.ObjID] = oplog
	info.MediaBlockInfo[*blockInfo.ID] = oplog

	return nil
}
