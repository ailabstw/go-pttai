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

func (pm *BaseProtocolManager) handleDeleteMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	return nil, types.ErrNotImplemented

	/*
		obj := NewEmptyMember()
		pm.SetMemberObjDB(obj)

		toBroadcastLogs, err := pm.HandleDeleteObjectLog(oplog, info, obj, nil, pm.SetMemberDB, pm.postdeleteMember)
		if err != nil {
			return nil, err
		}

		return toBroadcastLogs, nil
	*/
}

func (pm *BaseProtocolManager) handlePendingDeleteMemberLog(oplog *BaseOplog, info *ProcessPersonInfo) ([]*BaseOplog, error) {

	return nil, types.ErrNotImplemented

	/*
		obj := NewEmptyMember()
		pm.SetMemberObjDB(obj)

		return pm.HandlePendingDeleteObjectLog(oplog, info, obj, nil, pm.SetMemberDB)
	*/
}

func (pm *BaseProtocolManager) setNewestDeleteMemberLog(oplog *BaseOplog) (types.Bool, error) {
	return false, types.ErrNotImplemented

	/*
		obj := NewEmptyMember()
		pm.SetMemberObjDB(obj)

		return pm.SetNewestDeleteObjectLog(oplog, obj)
	*/
}

func (pm *BaseProtocolManager) handleFailedDeleteMemberLog(oplog *BaseOplog) error {
	return types.ErrNotImplemented

	/*
		obj := NewEmptyMember()
		pm.SetMemberObjDB(obj)

		return pm.HandleFailedDeleteObjectLog(oplog, obj)
	*/

}
