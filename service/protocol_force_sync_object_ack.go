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

func (pm *BaseProtocolManager) HandleForceSyncObjectAck(
	obj Object,
	peer *PttPeer,

	origObj Object,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),

) error {

	// oplog
	objID := obj.GetID()
	logID := obj.GetLogID()

	oplog := &BaseOplog{ID: logID}
	setLogDB(oplog)

	err := oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	// the temporal-oplog may be already deleted.
	err = oplog.Get(logID, true)
	if err != nil {
		return nil
	}

	// obj lock.
	err = obj.Lock()
	if err != nil {
		return err
	}
	defer obj.Unlock()

	// get orig-obj
	origObj.SetID(objID)
	err = origObj.GetByID(true)
	if err == nil && origObj.GetStatus() != types.StatusFailed {
		return ErrInvalidStatus
	}

	obj.SetIsGood(true)
	isAllGood := obj.CheckIsAllGood()
	if !isAllGood {
		return obj.Save(true)
	}

	err = pm.saveNewObjectWithOplog(obj, oplog, true, true, nil)
	if err != nil {
		return err
	}

	// save oplog
	if oplog.IsSync {
		return nil
	}

	oplog.IsSync = true

	return oplog.Save(true, merkle)
}
