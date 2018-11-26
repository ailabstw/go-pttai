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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) removeBlockAndMediaInfoBySyncInfo(
	syncInfo SyncInfo,

	info ProcessInfo,
	oplog *BaseOplog,
	isRetainValid bool,

	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),
	setLogDB func(oplog *BaseOplog),
) error {

	// remove oplog
	syncLogID := syncInfo.GetLogID()
	_, err := pm.removeNonSyncOplog(setLogDB, syncLogID, isRetainValid, oplog.UpdateTS, false)
	log.Debug("removeBlockAndMediaInfoBySyncInfo: after removeNonSyncOplog", "e", err)
	if err != nil {
		return err
	}

	// remove block
	blockInfo := syncInfo.GetBlockInfo()
	return pm.removeBlockAndMediaInfoByBlockInfo(blockInfo, info, oplog, true, removeMediaInfoByBlockInfo)
}

func (pm *BaseProtocolManager) removeBlockAndMediaInfoByBlockInfo(
	blockInfo *BlockInfo,

	info ProcessInfo,
	oplog *BaseOplog,

	isRemoveDB bool,

	removeMediaInfoByBlockInfo func(blockInfo *BlockInfo, info ProcessInfo, oplog *BaseOplog),
) error {

	log.Debug("removeBlockAndMediaInfoByBlockInfo: start", "blockInfo", blockInfo)

	if blockInfo == nil {
		return nil
	}

	if info != nil && removeMediaInfoByBlockInfo != nil {
		removeMediaInfoByBlockInfo(blockInfo, info, oplog)
	}

	var err error
	if isRemoveDB {
		pm.SetBlockInfoDB(blockInfo, oplog.ObjID)
		err = blockInfo.Remove(false)
		log.Debug("removeBlockAndMediaInfoBySyncInfo: after remove DB", "e", err)
	}

	return nil
}

func (pm *BaseProtocolManager) RemoveMediaInfosByOplog(mediaInfos map[types.PttID]*BaseOplog, mediaIDs []*types.PttID, oplog *BaseOplog, deleteMediaOp OpType) {

	var origLog *BaseOplog
	var ok bool
	for _, mediaID := range mediaIDs {
		origLog, ok = mediaInfos[*mediaID]
		if ok && origLog.Op == deleteMediaOp {
			continue
		}

		mediaInfos[*mediaID] = oplog
	}
}

func (pm *BaseProtocolManager) removeNonSyncOplog(
	setDB func(oplog *BaseOplog),
	logID *types.PttID,

	isRetainValid bool,
	newUpdateTS types.Timestamp,

	isLocked bool,
) (*BaseOplog, error) {

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

	if isRetainValid && status == types.StatusAlive && oplog.UpdateTS.IsLess(newUpdateTS) {
		oplog.IsSync = true
		err = oplog.Save(true)
		return oplog, nil
	}

	err = oplog.Delete(true)

	return nil, err
}
