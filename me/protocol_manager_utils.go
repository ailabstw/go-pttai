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

package me

import (
	"reflect"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) LoadMyNodes() error {
	myInfo := pm.Entity().(*MyInfo)
	myNodes := make(map[uint64]*MyNode)
	myNodeByNodeSignIDs := make(map[types.PttID]*MyNode)
	myID := myInfo.ID
	ptt := pm.myPtt

	log.Info("LoadMyNodes: start", "myID", myInfo.ID)
	isMyNodeID := false

	myNode := &MyNode{ID: myInfo.ID}
	key, err := myNode.DBPrefix()
	if err != nil {
		return err
	}

	iter, err := dbMyNodes.NewIteratorWithPrefix(nil, key, pttdb.ListOrderNext)
	if err != nil {
		return err
	}
	defer iter.Release()

	toRemoveIDs := make([][]byte, 0)
	myNodeID := ptt.MyNodeID()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		eachMyNode := &MyNode{}
		err := eachMyNode.Unmarshal(v)
		if err != nil {
			log.Error("my nodes is unable to unmarshal, removing", "k", k, "v", v)
			toRemoveIDs = append(toRemoveIDs, common.CloneBytes(k))
			continue
		}

		if !reflect.DeepEqual(eachMyNode.ID, myID) {
			log.Error("my nodes is not the same ID as me, removing", "myNode", eachMyNode.ID, "me", myInfo.ID)
			toRemoveIDs = append(toRemoveIDs, common.CloneBytes(k))
			continue
		}

		if reflect.DeepEqual(eachMyNode.NodeID, myNodeID) {
			isMyNodeID = true
		}

		nodeSignID, err := setNodeSignID(eachMyNode.NodeID, myID)
		if err != nil {
			continue
		}

		log.Debug("LoadMyNodes: (in-for-loop)", "eachMyNode", eachMyNode)
		myNodes[eachMyNode.RaftID] = eachMyNode
		myNodeByNodeSignIDs[*nodeSignID] = eachMyNode
		pm.totalWeight += eachMyNode.Weight

	}

	log.Info("LoadMyNodes: after loop", "isMyNodeID", isMyNodeID)
	if !isMyNodeID {
		return ErrInvalidMe
	}

	pm.MyNodes = myNodes
	pm.MyNodeByNodeSignIDs = myNodeByNodeSignIDs

	myNode = &MyNode{}
	for _, eachID := range toRemoveIDs {
		err := myNode.DeleteRawKey(eachID)
		if err != nil {
			continue
		}
	}

	return nil
}

func (pm *ProtocolManager) LoadPeers() error {
	ptt := pm.myPtt
	opKey, err := pm.GetOldestOpKey(false)
	if err != nil {
		return err
	}

	myNodeID := ptt.MyNodeID()
	for _, myNode := range pm.MyNodes {
		if reflect.DeepEqual(myNode.NodeID, myNodeID) {
			continue
		}
		ptt.AddDial(myNode.NodeID, opKey.Hash)
	}
	return nil
}

func (pm *ProtocolManager) GetJoinKeyFromHash(hash *common.Address) (*pkgservice.KeyInfo, error) {
	keyInfo, err := pm.BaseProtocolManager.GetJoinKeyFromHash(hash)
	if err == nil {
		return keyInfo, nil
	}

	pm.lockJoinFriendRequest.RLock()
	defer pm.lockJoinFriendRequest.RUnlock()

	for _, eachKeyInfo := range pm.joinFriendKeyInfos {
		if reflect.DeepEqual(hash, eachKeyInfo.Hash) {
			keyInfo = eachKeyInfo
			break
		}
	}

	if keyInfo == nil {
		return nil, pkgservice.ErrInvalidKeyInfo
	}

	return keyInfo, nil
}
