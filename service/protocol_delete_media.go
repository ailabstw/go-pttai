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

func (pm *BaseProtocolManager) BaseDeleteMedia(
	id *types.PttID,
	op OpType,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
) error {

	media := NewEmptyMedia()
	pm.SetMediaDB(media)

	opData := &OpDeleteMedia{}

	return pm.DeleteObject(
		id,
		op,

		media,
		opData,

		merkle,

		setLogDB,
		newOplog,
		nil,
		pm.setPendingDeleteMediaSyncInfo,
		broadcastLog,
		nil,
	)
}

func (pm *BaseProtocolManager) setPendingDeleteMediaSyncInfo(
	obj Object,
	status types.Status,
	oplog *BaseOplog,
) error {

	syncInfo := &BaseSyncInfo{}
	syncInfo.InitWithOplog(status, oplog)

	obj.SetSyncInfo(syncInfo)

	return nil
}
