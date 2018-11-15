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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) RemoveUserNode(nodeID *discover.NodeID) error {
	// 1. validate
	myID := pm.Ptt().GetMyEntity().GetID()
	if !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	origObj := NewEmptyUserNode()
	pm.SetUserNodeDB(origObj)
	id, err := origObj.GetIDByNodeID(nodeID)
	log.Debug("RemoveUserNode: after GetIDByNodeID", "nodeID", nodeID, "id", id, "e", err)
	if err != nil {
		return err
	}

	opData := &UserOpRemoveUserNode{
		NodeID: nodeID,
	}

	err = pm.DeleteObject(
		id, UserOpTypeRemoveUserNode,
		origObj, opData,
		pm.SetUserDB, pm.NewUserOplog, nil, pm.setPendingDeleteUserNodeSyncInfo, pm.broadcastUserOplogCore, pm.postdeleteUserNode)
	log.Debug("RemoveUserNode: after DeleteObject", "e", err)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) setPendingDeleteUserNodeSyncInfo(theObj pkgservice.Object, status types.Status, oplog *pkgservice.BaseOplog) error {

	obj, ok := theObj.(*UserNode)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	syncInfo := &pkgservice.BaseSyncInfo{}
	syncInfo.InitWithDeleteOplog(status, oplog)

	obj.SyncInfo = syncInfo

	return nil
}

func (pm *ProtocolManager) postdeleteUserNode(
	id *types.PttID,

	oplog *pkgservice.BaseOplog,
	opData pkgservice.OpData,

	origObj pkgservice.Object,
	blockInfo *pkgservice.BlockInfo,
) error {

	var err error
	userNodeInfo := pm.userNodeInfo

	pm.lockUserNodeInfo.Lock()
	defer pm.lockUserNodeInfo.Unlock()

	obj := NewEmptyUserNode()
	pm.SetUserNodeDB(obj)

	theUserNodes, err := pkgservice.GetObjList(obj, nil, 0, pttdb.ListOrderNext, false)
	log.Debug("postdeleteUserNode: after GetObjList", "e", err, "userNodes", theUserNodes)
	if err != nil {
		return err
	}

	userNodes := pkgservice.AliveObjects(theUserNodes)

	lenUserNodes := len(userNodes)
	userNodeInfo.NUserNode = lenUserNodes

	if lenUserNodes == 0 {
		userNodeInfo.UserNodeID = nil
		return userNodeInfo.Save()
	}

	if !reflect.DeepEqual(userNodeInfo.UserNodeID, id) {
		userNodeInfo.UserNodeID = nil
		return userNodeInfo.Save()
	}

	randN := rand.Intn(lenUserNodes)

	selectedUserNode := userNodes[randN]

	userNodeInfo.UserNodeID = selectedUserNode.GetID()

	return userNodeInfo.Save()
}
