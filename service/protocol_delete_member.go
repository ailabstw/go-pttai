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

func (pm *BaseProtocolManager) DeleteMember(id *types.PttID) (bool, error) {
	return false, types.ErrNotImplemented

	/*
		person := NewEmptyMember()
		pm.SetMemberObjDB(person)

		opData := &MemberOpDeleteMember{}

		err := pm.DeletePerson(
			id, MemberOpTypeDeleteMember,
			person, opData,
			pm.NewMemberOplog, pm.broadcastMemberOplogCore, pm.postdeleteMember,
		)
		if err != nil {
			return false, err
		}

		return true, nil
	*/
}

/**********
 * utils
 **********/

/*
postdeleteMember deals with ops after deletingMember. Assuming obj already locked (in DeleteObject and DeleteObjectLogs).
*/
func (pm *BaseProtocolManager) postdeleteMember(id *types.PttID, oplog *BaseOplog, origObj Object, opData OpData) error {

	return nil
}
