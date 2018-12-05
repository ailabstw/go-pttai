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

func (pm *BaseProtocolManager) HandleCreateMediaLogs(
	oplog *BaseOplog,
	info ProcessInfo,

	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,

) ([]*BaseOplog, error) {
	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	opData := &OpCreateMedia{}

	return pm.HandleCreateObjectLog(
		oplog,
		obj,
		opData,
		info,

		existsInInfo,
		pm.newMediaWithOplog,
		nil,
		updateCreateInfo,
	)
}

func (pm *BaseProtocolManager) HandlePendingCreateMediaLogs(
	oplog *BaseOplog,
	info ProcessInfo,

	existsInInfo func(oplog *BaseOplog, info ProcessInfo) (bool, error),
	updateCreateInfo func(obj Object, oplog *BaseOplog, opData OpData, info ProcessInfo) error,

) (types.Bool, []*BaseOplog, error) {
	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	opData := &OpCreateMedia{}

	return pm.HandlePendingCreateObjectLog(
		oplog,
		obj,
		opData,
		info,

		existsInInfo,
		pm.newMediaWithOplog,
		nil,
		updateCreateInfo,
	)
}

func (pm *BaseProtocolManager) SetNewestCreateMediaLog(oplog *BaseOplog) (types.Bool, error) {
	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.SetNewestCreateObjectLog(oplog, obj)
}

func (pm *BaseProtocolManager) HandleFailedCreateMediaLog(oplog *BaseOplog) error {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.HandleFailedCreateObjectLog(oplog, obj, nil)
}

func (pm *BaseProtocolManager) newMediaWithOplog(oplog *BaseOplog, theOpData OpData) Object {

	opData, ok := theOpData.(*OpCreateMedia)
	if !ok {
		return nil
	}

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)
	NewObjectWithOplog(obj, oplog)

	blockInfo, err := NewBlockInfo(opData.BlockInfoID, opData.Hashs, nil, oplog.CreatorID)
	if err != nil {
		return nil
	}
	pm.SetBlockInfoDB(blockInfo, obj.ID)
	blockInfo.InitIsGood()
	obj.SetBlockInfo(blockInfo)

	return obj
}
