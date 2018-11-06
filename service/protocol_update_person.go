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
	"bytes"

	"github.com/ailabstw/go-pttai/common/types"
)

func (pm *BaseProtocolManager) UpdatePerson(
	id *types.PttID,
	createOp OpType,
	isForce bool,

	origPerson Object,
	opData OpData,

	setLogDB func(oplog *BaseOplog),
	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
	postupdate func(obj Object, oplog *BaseOplog) error,
) (Object, *BaseOplog, error) {

	var err error

	// 3. check validity
	origStatus := origPerson.GetStatus()
	origStatusClass := types.StatusToStatusClass(origStatus)
	if origStatusClass <= types.StatusClassAlive {
		return nil, nil, types.ErrAlreadyExists
	}
	if origStatus == types.StatusTransferred {
		return nil, nil, types.ErrAlreadyDeleted
	}

	// 4. oplog
	theOplog, err := newOplog(id, createOp, opData)
	if err != nil {
		return nil, nil, err
	}
	oplog := theOplog.GetBaseOplog()
	oplog.PreLogID = origPerson.GetLogID()

	err = pm.SignOplog(oplog)
	if err != nil {
		return nil, nil, err
	}

	// 5. update person
	oplogStatus := oplog.ToStatus()

	if oplogStatus == types.StatusAlive {
		err = pm.handleUpdatePersonLogCore(oplog, origPerson, opData, setLogDB, postupdate)
	} else {
		err = pm.handlePendingUpdatePersonLogCore(oplog, origPerson, opData, setLogDB)
	}
	if err != nil {
		return nil, nil, err
	}

	// oplog save
	err = oplog.Save(false)
	if err != nil {
		return nil, nil, err
	}

	broadcastLog(oplog)

	return origPerson, oplog, nil
}

func isReplaceOrigSyncPersonInfo(syncInfo SyncInfo, status types.Status, ts types.Timestamp, newLogID *types.PttID) bool {

	if syncInfo == nil {
		return true
	}

	statusClass := types.StatusToStatusClass(status)
	syncStatusClass := types.StatusToStatusClass(syncInfo.GetStatus())

	switch syncStatusClass {
	case types.StatusClassInternalDelete:
		syncStatusClass = types.StatusClassInternalPendingAlive
	case types.StatusClassPendingDelete:
		syncStatusClass = types.StatusClassPendingAlive
	case types.StatusClassDeleted:
		syncStatusClass = types.StatusClassAlive
	}

	if statusClass < syncStatusClass {
		return false
	}
	if statusClass > syncStatusClass {
		return true
	}

	syncTS := syncInfo.GetUpdateTS()
	if syncTS.IsLess(ts) {
		return false
	}
	if ts.IsLess(syncTS) {
		return true
	}

	origLogID := syncInfo.GetLogID()
	return bytes.Compare(origLogID[:], newLogID[:]) > 0
}
