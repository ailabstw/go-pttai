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

import "github.com/ailabstw/go-pttai/common/types"

func (pm *BaseProtocolManager) removeBlockAndInfoBySyncInfo(
	syncInfo SyncInfo,
	info ProcessInfo,
	oplog *BaseOplog,
	isRetainValid bool,

	removeInfoByBlockInfo func(blockInfo BlockInfo, info ProcessInfo, oplog *BaseOplog),
	setLogDB func(oplog *BaseOplog),
) error {

	// remove oplog
	syncLogID := syncInfo.GetLogID()
	_, err := pm.removeNonSyncOplog(setLogDB, syncLogID, isRetainValid, false)
	if err != nil {
		return err
	}

	// remove block
	blockInfo := syncInfo.GetBlock()
	return pm.removeBlockAndInfoByBlock(blockInfo, info, oplog, true, removeInfoByBlockInfo)
}

func (pm *BaseProtocolManager) removeBlockAndInfoByBlock(
	blockInfo BlockInfo,
	info ProcessInfo,
	oplog *BaseOplog,
	isRemoveDB bool,

	removeInfoByBlockInfo func(blockInfo BlockInfo, info ProcessInfo, oplog *BaseOplog),
) error {

	if blockInfo == nil {
		return nil
	}

	if removeInfoByBlockInfo != nil {
		removeInfoByBlockInfo(blockInfo, info, oplog)
	}

	if isRemoveDB {
		return blockInfo.Remove()
	}

	return nil
}

func (pm *BaseProtocolManager) removeNonSyncOplog(setDB func(oplog *BaseOplog), logID *types.PttID, isRetainValid bool, isLocked bool) (*BaseOplog, error) {

	oplog := &BaseOplog{}
	setDB(oplog)
	oplog.ID = logID

	if !isLocked {
		err := oplog.Lock()
		if err != nil {
			return nil, err
		}
		defer oplog.Unlock()
	}

	err := oplog.Get(logID, true)
	if err != nil {
		return nil, err
	}

	status := oplog.ToStatus()
	if oplog.IsSync && status == types.StatusAlive {
		return nil, nil
	}

	if isRetainValid && status == types.StatusAlive {
		oplog.IsSync = true
		err = oplog.SaveWithIsSync(true)
		return oplog, nil
	}

	err = oplog.Delete(true)

	return nil, err
}
