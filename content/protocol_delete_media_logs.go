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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleDeleteMediaLogs(
	oplog *pkgservice.BaseOplog,
	info *ProcessBoardInfo,
) ([]*pkgservice.BaseOplog, error) {

	return pm.BaseHandleDeleteMediaLogs(
		oplog,
		info,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.updateMediaDeleteInfo,
	)
}

func (pm *ProtocolManager) handlePendingDeleteMediaLogs(oplog *pkgservice.BaseOplog, info *ProcessBoardInfo) (types.Bool, []*pkgservice.BaseOplog, error) {

	return pm.BaseHandlePendingDeleteMediaLogs(
		oplog,
		info,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.updateMediaDeleteInfo,
	)
}

func (pm *ProtocolManager) handleFailedValidDeleteMediaLog(
	oplog *pkgservice.BaseOplog,
	info *ProcessBoardInfo,

) error {

	return pm.BaseHandleFailedValidDeleteMediaLog(
		oplog,
		info,

		pm.updateMediaDeleteInfo,
	)
}

/**********
 * Customize
 **********/

func (pm *ProtocolManager) updateMediaDeleteInfo(obj pkgservice.Object, oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) error {

	info, ok := theInfo.(*ProcessBoardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	info.MediaInfo[*oplog.ObjID] = oplog

	return nil
}
