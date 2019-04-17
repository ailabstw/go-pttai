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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type PersonOpTransferPerson struct {
	ToID *types.PttID `json:"t"`
}

/**********
 * Handle CreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandleTransferPersonLog(
	oplog *BaseOplog,
	origPerson Object,
	opData *PersonOpTransferPerson,

	merkle *Merkle,

	statusTransferred types.Status,

	setLogDB func(oplog *BaseOplog),
	posttransfer func(fromID *types.PttID, toID *types.PttID, person Object, oplog *BaseOplog, opData OpData) error,

) ([]*BaseOplog, error) {

	// 1. lock person
	fromID := oplog.ObjID
	origPerson.SetID(fromID)

	err := origPerson.Lock()
	if err != nil {
		return nil, err
	}
	defer origPerson.Unlock()

	// 2. get person
	err = origPerson.GetByID(true)
	if err != nil {
		return nil, ErrNewerOplog
	}
	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		return nil, ErrInvalidPreLog
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus == statusTransferred {
		if oplog.UpdateTS.IsLess(origPerson.GetUpdateTS()) {
			err = pm.saveUpdateObjectWithOplog(origPerson, oplog, true)
			if err != nil {
				return nil, err
			}
		}
		return nil, ErrNewerOplog
	}

	// 4. core
	err = pm.handleTransferPersonLogCore(
		oplog,
		origPerson,
		opData,

		merkle,

		statusTransferred,

		setLogDB,
		posttransfer,
	)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (pm *BaseProtocolManager) handleTransferPersonLogCore(
	oplog *BaseOplog,
	origPerson Object,
	opData *PersonOpTransferPerson,

	merkle *Merkle,

	statusTransferred types.Status,

	setLogDB func(oplog *BaseOplog),
	posttransfer func(fromID *types.PttID, toID *types.PttID, person Object, oplog *BaseOplog, opData OpData) error,

) error {

	fromID := oplog.ObjID
	toID := opData.ToID

	var err error

	// 1. check sync-info
	oplogStatus := statusTransferred

	var isReplaceSyncInfo bool
	origSyncInfo := origPerson.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncPersonInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
	}

	// 1.1. replace sync-info
	if isReplaceSyncInfo {
		err = pm.removeBlockAndMediaInfoBySyncInfo(
			origSyncInfo,

			nil,
			oplog,
			true,

			merkle,

			nil,
			setLogDB,
		)
		if err != nil {
			return err
		}
		origPerson.SetSyncInfo(nil)
	}

	// 4. saveUpdateObj
	err = pm.saveDeleteObjectWithOplog(origPerson, oplog, oplogStatus, true)
	log.Debug("handleTransferPersonLogCore: after saveDeleteObjectWithOplog", "e", err)
	if err != nil {
		return err
	}

	log.Debug("handleTransferPersonLogCore: to posttransfer")

	posttransfer(fromID, toID, origPerson, oplog, opData)

	return nil
}

/**********
 * Handle PendingCreateObjectLog
 **********/

func (pm *BaseProtocolManager) HandlePendingTransferPersonLog(
	oplog *BaseOplog,
	origPerson Object,
	opData *PersonOpTransferPerson,

	merkle *Merkle,

	statusInternalTransfer types.Status,
	statusPendingTransfer types.Status,
	statusTransferred types.Status,

	setLogDB func(oplog *BaseOplog),

) (types.Bool, []*BaseOplog, error) {

	// 1. lock person
	fromID := oplog.ObjID
	origPerson.SetID(fromID)

	err := origPerson.Lock()
	if err != nil {
		return false, nil, err
	}
	defer origPerson.Unlock()

	// 2. get person
	err = origPerson.GetByID(true)
	if err != nil {
		return false, nil, err
	}
	if !reflect.DeepEqual(origPerson.GetLogID(), oplog.PreLogID) {
		return false, nil, ErrInvalidPreLog
	}

	// 3. check validity
	origStatus := origPerson.GetStatus()
	origStatusClass := types.StatusToStatusClass(origStatus)
	statusClass := types.StatusToStatusClass(statusTransferred)
	if origStatusClass == statusClass {
		return false, nil, ErrNewerOplog
	}

	// 4. core
	err = pm.handlePendingTransferPersonLogCore(
		oplog,

		origPerson,
		opData,

		merkle,

		statusInternalTransfer,
		statusPendingTransfer,
		statusTransferred,

		setLogDB,
	)
	if err != nil {
		return false, nil, err
	}

	return true, nil, nil
}

func (pm *BaseProtocolManager) handlePendingTransferPersonLogCore(
	oplog *BaseOplog,

	origObj Object,
	opData OpData,

	merkle *Merkle,

	statusInternalTransfer types.Status,
	statusPendingTransfer types.Status,
	statusTransferred types.Status,

	setLogDB func(oplog *BaseOplog),
) error {

	var err error

	// 1. sync info
	oplogStatus := types.StatusToDeleteStatus(
		oplog.ToStatus(),
		statusInternalTransfer,
		statusPendingTransfer,
		statusTransferred,
	)

	var isReplaceSyncInfo bool
	origSyncInfo := origObj.GetSyncInfo()

	if origSyncInfo != nil {
		isReplaceSyncInfo = isReplaceOrigSyncPersonInfo(origSyncInfo, oplogStatus, oplog.UpdateTS, oplog.ID)
		if !isReplaceSyncInfo {
			return ErrNewerOplog
		}

		// 1.1 replace sync-info
		syncLogID := origSyncInfo.GetLogID()
		if !reflect.DeepEqual(syncLogID, oplog.ID) {
			pm.removeBlockAndMediaInfoBySyncInfo(
				origSyncInfo,

				nil,
				oplog,
				false,

				merkle,

				nil,
				setLogDB,
			)
		}
		origObj.SetSyncInfo(nil)
	}

	// 4. save obj
	SetPendingPersonSyncInfo(origObj, oplogStatus, oplog)
	err = origObj.Save(true)
	if err != nil {
		return err
	}

	return nil
}
