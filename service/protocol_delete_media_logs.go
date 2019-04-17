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

package service

import (
	"github.com/ailabstw/go-pttai/common/types"
)

func (pm *BaseProtocolManager) BaseHandleDeleteMediaLogs(
	oplog *BaseOplog,
	info ProcessInfo,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,

) ([]*BaseOplog, error) {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	opData := &OpDeleteMedia{}

	return pm.HandleDeleteObjectLog(
		oplog,
		info,

		obj,
		opData,

		merkle,

		setLogDB,
		nil,
		nil,
		updateDeleteInfo,
	)
}

func (pm *BaseProtocolManager) BaseHandlePendingDeleteMediaLogs(
	oplog *BaseOplog,
	info ProcessInfo,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,

) (types.Bool, []*BaseOplog, error) {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	opData := &OpDeleteMedia{}

	return pm.HandlePendingDeleteObjectLog(
		oplog,
		info,

		obj,
		opData,

		merkle,

		setLogDB,
		nil,
		pm.setPendingDeleteMediaSyncInfo,
		updateDeleteInfo,
	)
}

func (pm *BaseProtocolManager) SetNewestDeleteMediaLog(oplog *BaseOplog) (types.Bool, error) {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.SetNewestDeleteObjectLog(oplog, obj)
}

func (pm *BaseProtocolManager) HandleFailedDeleteMediaLog(oplog *BaseOplog) error {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.HandleFailedDeleteObjectLog(oplog, obj)
}

func (pm *BaseProtocolManager) BaseHandleFailedValidDeleteMediaLog(
	oplog *BaseOplog,
	info ProcessInfo,

	updateDeleteInfo func(obj Object, oplog *BaseOplog, info ProcessInfo) error,
) error {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.HandleFailedValidDeleteObjectLog(oplog, obj, info, updateDeleteInfo)
}
