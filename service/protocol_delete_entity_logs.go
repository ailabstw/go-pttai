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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) HandleDeleteEntityLog(
	oplog *BaseOplog,
	info ProcessInfo,

	opData OpData,
	status types.Status,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	postdelete func(opData OpData, isForce bool) error,
	updateDeleteInfo func(oplog *BaseOplog, info ProcessInfo) error,
) ([]*BaseOplog, error) {

	log.Debug("HandleDeleteEntityLog: start", "e", pm.Entity().GetID(), "status", status)

	entity := pm.Entity()

	err := oplog.GetData(opData)
	if err != nil {
		return nil, err
	}

	// 1. lock object
	err = entity.Lock()
	if err != nil {
		return nil, err
	}
	defer entity.Unlock()

	// 3. already deleted
	statusClass := types.StatusToStatusClass(status)
	origStatus := entity.GetStatus()
	origStatusClass := types.StatusToStatusClass(origStatus)
	if origStatusClass == statusClass {
		if oplog.UpdateTS.IsLess(entity.GetUpdateTS()) {
			err = SetNewEntityWithOplog(entity, status, oplog)
			if err != nil {
				return nil, err
			}
			err = entity.Save(true)
			if err != nil {
				return nil, err
			}
		}
		return nil, ErrNewerOplog
	}
	if origStatusClass >= types.StatusClassDeleted {
		return nil, ErrNewerOplog
	}

	// 4. sync-info
	origSyncInfo := entity.GetSyncInfo()
	if origSyncInfo != nil {
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			err = pm.removeBlockAndMediaInfoBySyncInfo(
				origSyncInfo,

				info,
				oplog,
				true,

				merkle,

				nil,
				setLogDB,
			)
			if err != nil {
				return nil, err
			}
		}

		entity.SetSyncInfo(nil)
	}

	// 7. saveDeleteEntity
	err = SetNewEntityWithOplog(entity, status, oplog)
	log.Debug("HandleDeleteEntityLog: after SetNewEntityWithOplog", "e", err, "entity.Status", entity.GetStatus(), "entity", entity.GetID())
	if err != nil {
		return nil, err
	}
	err = entity.Save(true)
	if err != nil {
		return nil, err
	}

	// 7.1
	if postdelete != nil {
		postdelete(opData, false)
	}

	// 6. set oplog is-sync
	oplog.IsSync = true

	// 8. updateDeleteInfo
	updateDeleteInfo(oplog, info)

	log.Debug("HandleDeleteEntityLog: done", "e", pm.Entity().GetID())

	return nil, nil
}

/**********
 * Handle PendingDeleteEntityLog
 **********/

func (pm *BaseProtocolManager) HandlePendingDeleteEntityLog(
	oplog *BaseOplog,
	info ProcessInfo,

	internalPendingStatus types.Status,
	pendingStatus types.Status,
	op OpType,
	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	setPendingDeleteSyncInfo func(entity Entity, status types.Status, oplog *BaseOplog) error,

	updateDeleteInfo func(oplog *BaseOplog, info ProcessInfo) error,
) (types.Bool, []*BaseOplog, error) {

	entity := pm.Entity()

	log.Debug("HandlePendingDeleteEntityLog: start", "entity", pm.Entity().GetID())

	// 1. lock obj
	err := entity.Lock()
	if err != nil {
		return false, nil, err
	}
	defer entity.Unlock()

	// 3. already deleted
	origStatus := entity.GetStatus()
	if origStatus >= types.StatusDeleted {
		return false, nil, ErrNewerOplog
	}

	// 4. sync info
	oplogStatus := types.StatusToDeleteStatus(oplog.ToStatus(), internalPendingStatus, pendingStatus, types.StatusDeleted)

	origSyncInfo := entity.GetSyncInfo()

	if origSyncInfo != nil {
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			if !isReplaceOrigSyncInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID) {
				return false, nil, ErrNewerOplog
			}

			pm.removeBlockAndMediaInfoBySyncInfo(
				origSyncInfo,

				info,
				oplog,
				false,

				merkle,

				nil,
				setLogDB,
			)
		}
		entity.SetSyncInfo(nil)
	}

	// 5. save obj
	setPendingDeleteSyncInfo(entity, oplogStatus, oplog)
	err = entity.Save(true)
	if err != nil {
		return false, nil, err
	}

	// 7. update delete info
	updateDeleteInfo(oplog, info)

	return true, nil, nil
}

/**********
 * Handle Failed DeleteEntityLog
 **********/

func (pm *BaseProtocolManager) HandleFailedDeleteEntityLog(
	oplog *BaseOplog,
) error {

	entity := pm.Entity()

	// 1. lock obj
	err := entity.Lock()
	if err != nil {
		return err
	}
	defer entity.Unlock()

	// 3. check validity
	syncInfo := entity.GetSyncInfo()
	if syncInfo == nil || !reflect.DeepEqual(syncInfo.GetLogID(), oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(syncInfo.GetUpdateTS()) {
		return nil
	}

	// 4. handle fails
	entity.SetSyncInfo(nil)
	err = entity.Save(true)
	if err != nil {
		return err
	}

	return nil
}

/**********
 * Handle Failed DeleteEntityLog
 **********/

func (pm *BaseProtocolManager) HandleFailedValidDeleteEntityLog(
	oplog *BaseOplog,
) error {
	return types.ErrNotImplemented
}
