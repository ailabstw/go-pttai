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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ApproveJoinMe struct {
	MyInfo *MyInfo `json:"M"`

	OpKeyInfo *pkgservice.KeyInfo `json:"O"`
}

/*
ApproveJoinMe deals with procedure of approving-join-me:
	1. set peer-type as me
	2. propose raft.
*/
func (pm *ProtocolManager) ApproveJoinMe(joinEntity *pkgservice.JoinEntity, keyInfo *pkgservice.KeyInfo, peer *pkgservice.PttPeer) (*pkgservice.KeyInfo, interface{}, error) {
	log.Debug("ApproveJoinMe: start")

	myInfo := pm.Entity().(*MyInfo)
	myID := myInfo.ID

	pm.lockMyNodes.Lock()
	defer pm.lockMyNodes.Unlock()

	// 1. add my node
	peerID := peer.GetID()
	raftID, err := peerID.ToRaftID()
	if err != nil {
		return nil, nil, err
	}

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	myNode, err := NewMyNode(ts, myID, peerID, 0)
	if err != nil {
		return nil, nil, err
	}

	pm.MyNodes[raftID] = myNode

	// 2. register pending peer
	peer.UserID = myID
	pm.RegisterPendingPeer(peer)

	// 3. propose raft.
	pm.ProposeRaftAddNode(peerID, 1)

	// 4. get-opkey
	myOpKeyInfo, err := pm.GetNewestOpKey(false)
	if err != nil {
		log.Error("ApproveJoinMe: unable to get newest op key", "e", err)
		return nil, nil, err
	}

	// data
	data := &ApproveJoinMe{
		OpKeyInfo: myOpKeyInfo,

		MyInfo: myInfo,
	}

	return myOpKeyInfo, data, nil
}

/*
HandleApproveJoinMe deals with procedure of handling-approve-join-me:
*/
func (pm *ProtocolManager) HandleApproveJoinMe(dataBytes []byte, joinRequest *pkgservice.JoinRequest, peer *pkgservice.PttPeer) error {

	log.Debug("HandleApproveJoinMe: start", "peer", peer)

	approveJoin := &pkgservice.ApproveJoin{Data: &ApproveJoinMe{}}
	err := json.Unmarshal(dataBytes, approveJoin)
	if err != nil {
		log.Error("HandleApproveJoinMe: unable to unmarshal", "e", err)
		return err
	}
	approveJoinMe, ok := approveJoin.Data.(*ApproveJoinMe)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	ptt := pm.myPtt
	service := pm.Entity().Service().(*Backend)
	spm := service.SPM().(*ServiceProtocolManager)
	myNodeID := pm.myPtt.MyNodeID()

	// 1. new me
	newMyInfo := approveJoinMe.MyInfo

	// 2. my-node
	newMyNode, err := NewMyNode(ts, joinRequest.ID, myNodeID, 0)
	if err != nil {
		return err
	}
	newMyNode.Status = types.StatusInit
	_, err = newMyNode.Save()
	log.Debug("HandleApproveJoinMe: after myNode2.Save", "myNode2", newMyNode, "ID", newMyNode.ID, "e", err)
	if err != nil {
		return err
	}

	newMyNode2, err := NewMyNode(ts, joinRequest.ID, peer.GetID(), 0)
	if err != nil {
		return err
	}
	newMyNode2.Status = types.StatusAlive
	_, err = newMyNode2.Save()
	log.Debug("HandleApproveJoinMe: after myNode.Save", "myNode", newMyNode2, "ID", newMyNode2.ID, "e", err)
	if err != nil {
		return err
	}

	// my-info init
	newMyInfo.Status = types.StatusInit
	newMyInfo.UpdateTS = ts
	err = newMyInfo.Init(ptt, service, service.SPM())

	err = newMyInfo.Save(false)
	if err != nil {
		return err
	}

	// new op-key
	newPM := newMyInfo.PM()

	newMyOpKeyInfo := approveJoinMe.OpKeyInfo
	newMyOpKeyInfo.Init(newPM.DBOpKey(), newPM.DBObjLock(), newMyInfo.ID, newPM.DBOpKeyPrefix(), newPM.DBOpKeyIdxPrefix())
	err = newMyOpKeyInfo.Save(false)
	if err != nil {
		return err
	}

	err = newPM.RegisterOpKey(newMyOpKeyInfo, false)
	if err != nil {
		return err
	}

	// register peer
	log.Debug("HandleApproveJoinMe: to RegisterPeer")

	peer.UserID = newMyInfo.ID
	newPM.RegisterPendingPeer(peer)

	log.Debug("HandleApproveJoinMe: after RegisterPeer")

	// add to entities
	log.Debug("HandleApproveJoinMe: to RegisterEntity")
	err = spm.RegisterEntity(newMyInfo.ID, newMyInfo)
	if err != nil {
		return err
	}
	log.Debug("HandleApproveJoinMe: after RegisterEntity")

	// me-start
	newMyInfo.PrestartAndStart()

	// remove join-me-request
	pm.lockJoinMeRequest.Lock()
	defer pm.lockJoinMeRequest.Unlock()
	delete(pm.joinMeRequests, *joinRequest.Hash)

	log.Debug("HandleApproveJoinMe: done")

	return nil
}
