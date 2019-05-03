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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *ProtocolManager) CleanUserNode() error {
	// pm.lockUserNodeInfo.Lock()
	// defer pm.lockUserNodeInfo.Unlock()

	/*
		if pm.userNodeInfo != nil {
			pm.userNodeInfo.Delete()
			pm.userNodeInfo = nil
		}
	*/

	userNode := NewEmptyUserNode()
	pm.SetUserNodeDB(userNode)

	iter, err := userNode.BaseObject.GetObjIdxIterWithObj(nil, pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	var key []byte
	for iter.Next() {
		key = iter.Key()
		userNode.DB().DeleteAll(key)
	}

	return nil
}

func (pm *ProtocolManager) InitUserNode(entityID *types.PttID) {
	/*
		userNodeInfo := &UserNodeInfo{}
		err := userNodeInfo.Get(entityID)
		if err != nil {
			userNodeInfo = &UserNodeInfo{ID: entityID}
		}
		pm.userNodeInfo = userNodeInfo
	*/
}
