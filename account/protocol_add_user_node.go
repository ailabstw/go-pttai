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
	"math/rand"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) AddUserNode(nodeID *discover.NodeID) error {
	// 1. validate
	myID := pm.Ptt().GetMyEntity().GetID()
	if !pm.IsMaster(myID, false) {
		log.Error("AddUserNode: not Master", "myID", myID, "entity", pm.Entity().GetID())
		return types.ErrInvalidID
	}

	// 2. add object
	data := &UserOpAddUserNode{NodeID: nodeID}

	log.Debug("AddUserNode: to CreateObject", "entity", pm.Entity().GetID())
	_, err := pm.CreateObject(
		data,
		UserOpTypeAddUserNode,

		pm.NewUserNode,
		pm.NewUserOplogWithTS,
		nil,

		pm.SetUserDB,
		pm.broadcastUserOplogsCore,
		pm.broadcastUserOplogCore,

		pm.postcreateUserNode,
	)
	log.Debug("AddUserNode: after CreateObject", "entity", pm.Entity().GetID(), "nodeID", nodeID, "e", err)
	if err != nil {
		return err
	}
	return nil
}

func (pm *ProtocolManager) NewUserNode(theData pkgservice.CreateData) (pkgservice.Object, pkgservice.OpData, error) {

	data := theData.(*UserOpAddUserNode)

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	entity := pm.Entity().(*Profile)

	userNode, err := NewUserNode(ts, entity.MyID, entity.ID, nil, types.StatusInit, entity.MyID, data.NodeID)
	if err != nil {
		return nil, nil, err
	}
	pm.SetUserNodeDB(userNode)

	return userNode, data, nil
}

func (pm *ProtocolManager) postcreateUserNode(obj pkgservice.Object, oplog *pkgservice.BaseOplog) error {

	log.Debug("postcreateUserNode: start", "entity", pm.Entity().GetID(), "objID", oplog.ObjID)

	pm.lockUserNodeInfo.Lock()
	defer pm.lockUserNodeInfo.Unlock()

	pm.userNodeInfo.NUserNode++
	randN := rand.Intn(pm.userNodeInfo.NUserNode)
	if randN == 0 {
		pm.userNodeInfo.UserNodeID = oplog.ObjID
	}

	err := pm.userNodeInfo.Save()
	if err != nil {
		return err
	}

	return nil
}
