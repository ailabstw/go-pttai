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
)

func (pm *BaseProtocolManager) HandleFailedPersonLog(
	oplog *BaseOplog,
	person Object,
) error {

	objID := oplog.ObjID
	person.SetID(objID)

	// 1. lock
	err := person.Lock()
	if err != nil {
		return err
	}
	defer person.Unlock()

	// 2. get obj
	err = person.GetByID(true)
	if err != nil {
		return err
	}

	// 3. check validity
	origSyncInfo := person.GetSyncInfo()
	if origSyncInfo == nil || !reflect.DeepEqual(origSyncInfo.GetLogID(), oplog.ID) {
		return nil
	}

	if oplog.UpdateTS.IsLess(origSyncInfo.GetUpdateTS()) {
		return nil
	}

	// 4. remove block/oplog
	person.SetSyncInfo(nil)

	// 5. obj-save
	err = person.Save(true)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) HandleFailedValidPersonLog(
	oplog *BaseOplog,
	person Object,
) error {
	return types.ErrNotImplemented
}

/**********
 * Handle Failed UpdatePersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedUpdatePersonLog(
	oplog *BaseOplog,
	origPerson Object,
) error {
	return pm.HandleFailedPersonLog(oplog, origPerson)
}

/**********
 * Handle Failed DeletePersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedDeletePersonLog(
	oplog *BaseOplog,
	person Object,
) error {
	return pm.HandleFailedPersonLog(oplog, person)
}

/**********
 * Handle Failed TransferPersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedTransferPersonLog(
	oplog *BaseOplog,
	person Object,
) error {
	return pm.HandleFailedPersonLog(oplog, person)
}

/**********
 * Handle Failed Valid CreatePersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedValidCreatePersonLog(
	oplog *BaseOplog,
	person Object,

	prefailed func(obj Object, oplog *BaseOplog) error,
) error {
	return types.ErrNotImplemented
}

/**********
 * Handle Failed Valid UpdatePersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedValidUpdatePersonLog(
	oplog *BaseOplog,
	origPerson Object,
) error {
	return pm.HandleFailedValidPersonLog(oplog, origPerson)
}

/**********
 * Handle Failed Valid TransferPersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedValidTransferPersonLog(
	oplog *BaseOplog,
	origPerson Object,
) error {
	return pm.HandleFailedValidPersonLog(oplog, origPerson)
}

/**********
 * Handle Failed Valid TransferPersonLog
 **********/

func (pm *BaseProtocolManager) HandleFailedValidDeletePersonLog(
	oplog *BaseOplog,
	origPerson Object,
) error {
	return pm.HandleFailedValidPersonLog(oplog, origPerson)
}
