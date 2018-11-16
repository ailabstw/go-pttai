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

type SyncUserNodeAck struct {
	Objs []*UserNode `json:"o"`
}

func (pm *ProtocolManager) HandleSyncAddUserNodeAck(objs []*UserNode, peer *pkgservice.PttPeer) error {

	origObj := NewEmptyUserNode()
	pm.SetUserNodeDB(origObj)
	for _, obj := range objs {
		pm.SetUserNodeDB(obj)

		pm.HandleSyncCreateObjectAck(
			obj, peer, origObj,
			pm.SetUserDB, pm.updateCreateUserNode, pm.postcreateUserNode, pm.broadcastUserOplogCore)
	}

	return nil
}

func (pm *ProtocolManager) updateCreateUserNode(theToObj pkgservice.Object, theFromObj pkgservice.Object) error {
	toObj, ok := theToObj.(*UserNode)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*UserNode)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	toObj.UserID = fromObj.UserID
	toObj.NodeID = fromObj.NodeID

	return nil

}
