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

package account

import (
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) SyncAddUserNode(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer) error {

	userID := pm.Entity().GetID()

	opData := &UserOpAddUserNode{}

	var err error
	var obj *UserNode
	objs := make([]*UserNode, 0, len(oplogs))
	for _, oplog := range oplogs {
		err = oplog.GetData(opData)
		if err != nil {
			continue
		}

		obj = NewEmptyUserNode()
		pkgservice.NewObjectWithOplog(obj, oplog)
		obj.UserID = userID
		obj.NodeID = opData.NodeID
		objs = append(objs, obj)
	}

	return pm.HandleSyncAddUserNodeAck(objs, peer)
}
