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
	"github.com/ailabstw/go-pttai/log"
)

func (pm *BaseProtocolManager) UpdatePerson(
	id *types.PttID,
	addOp OpType,
	isForce bool,

	origPerson Object,

	opData OpData,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	newOplog func(objID *types.PttID, op OpType, opData OpData) (Oplog, error),
	broadcastLog func(oplog *BaseOplog) error,
	postupdate func(obj Object, oplog *BaseOplog) error,
) (Object, *BaseOplog, error) {

	var err error

	// 3. check validity
	origStatus := origPerson.GetStatus()
	if origStatus <= types.StatusAlive {
		return nil, nil, types.ErrAlreadyExists
	}
	if origStatus >= types.StatusMigrated {
		return nil, nil, types.ErrAlreadyDeleted
	}

	// 3.1. check original update-info
	origSyncInfo := origPerson.GetSyncInfo()
	if origSyncInfo != nil {
		return nil, nil, ErrAlreadyPending
	}

	// 4. oplog
	theOplog, err := newOplog(id, addOp, opData)
	if err != nil {
		return nil, nil, err
	}
	oplog := theOplog.GetBaseOplog()
	// oplog.PreLogID = origPerson.GetLogID()

	err = pm.SignOplog(oplog)
	if err != nil {
		return nil, nil, err
	}

	// 5. update person
	oplogStatus := oplog.ToStatus()

	log.Debug("UpdatePerson: to handleCore", "status", oplogStatus, "id", id, "entity", pm.Entity().IDString())

	if oplogStatus == types.StatusAlive {
		err = pm.handleUpdatePersonLogCore(
			oplog,
			origPerson,
			opData,

			merkle,

			setLogDB,
			postupdate,
		)
	} else {
		err = pm.handlePendingUpdatePersonLogCore(
			oplog,

			origPerson,
			opData,

			merkle,

			setLogDB,
		)
	}
	if err != nil {
		return nil, nil, err
	}

	// oplog save
	err = oplog.Save(false, merkle)
	if err != nil {
		return nil, nil, err
	}

	broadcastLog(oplog)

	return origPerson, oplog, nil
}
