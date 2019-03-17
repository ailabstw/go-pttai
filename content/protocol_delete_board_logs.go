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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleDeleteBoardLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) ([]*pkgservice.BaseOplog, error) {

	opData := &BoardOpDeleteBoard{}

	log.Debug("handleDeleteBoardLogs: start", "entity", pm.Entity().IDString())

	return pm.HandleDeleteEntityLog(
		oplog,
		info,

		opData,
		types.StatusTerminal,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		nil,
		pm.updateBoardDeleteInfo,
	)
}

func (pm *ProtocolManager) handlePendingDeleteBoardLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	opData := &BoardOpDeleteBoard{}

	return pm.HandlePendingDeleteEntityLog(
		oplog,
		info,

		types.StatusInternalTerminal,
		types.StatusPendingTerminal,
		BoardOpTypeDeleteBoard,
		opData,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.setPendingDeleteBoardSyncInfo,
		pm.updateBoardDeleteInfo,
	)
}

func (pm *ProtocolManager) setNewestDeleteBoardLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {

	return false, nil
}

func (pm *ProtocolManager) handleFailedDeleteBoardLog(oplog *pkgservice.BaseOplog) error {

	return pm.HandleFailedDeleteEntityLog(oplog)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) updateBoardDeleteInfo(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.BoardInfo[*oplog.ObjID] = oplog

	return nil
}
