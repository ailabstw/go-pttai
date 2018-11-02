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

    1. create-my-node with unknown node-type
    2. master-oplog
    2. ptt set peer type
    3. private-key-bytes
    4. op-key
    5. my-info
    6. user-name
    7. board-data
    8. register-peer
    9. sign-key-info
    10. me-oplogs
    11. master-oplogs
*/
func (pm *ProtocolManager) ApproveJoinMe(joinEntity *pkgservice.JoinEntity, keyInfo *pkgservice.KeyInfo, peer *pkgservice.PttPeer) (*pkgservice.KeyInfo, interface{}, error) {
	log.Debug("ApproveJoinMe: start")

	pm.lockMyNodes.Lock()
	defer pm.lockMyNodes.Unlock()

	peerID := peer.GetID()
	raftID, err := peerID.ToRaftID()

	if err != nil {
		return nil, nil, err
	}

	ptt := pm.myPtt

	ptt.SetPeerType(peer, pkgservice.PeerTypeMe, false, false)

	myID := pm.Entity().GetID()
	pm.ProposeRaftAddNode(peerID, 1)

	// get-opkey
	myOpKeyInfo, err := pm.GetNewestOpKey(false)
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

	// my-info
	myInfo := pm.Entity().(*MyInfo)

	// set my-node to pm
	pm.MyNodes[raftID] = myNode

	// XXX Hack for user-id register-peer
	peer.UserID = myID

	pm.RegisterPeer(peer)

	// data
	data := &ApproveJoinMe{
		OpKeyInfo: myOpKeyInfo,

		MyInfo: myInfo,
	}

	return myOpKeyInfo, data, nil
}

/*
HandleApproveJoinMe deals with procedure of handling-approve-join-me:

    1. migrate personal board
    2. remaster other entities (boards / friends)
    3. invalidate original my-info. (retain the opkey, but we will not use the opkey anymore.)
    8. new private-key
    4. new opkey
    5. new me (with status pending)
    6. sign-key
    7. user-name
    7. board
    9. new me-oplogs
    10. renew-me
    11. remove-join-requests.
    12. restart.
*/
func (pm *ProtocolManager) HandleApproveJoinMe(dataBytes []byte, joinRequest *pkgservice.JoinRequest, peer *pkgservice.PttPeer) error {

	approveJoin := &pkgservice.ApproveJoin{Data: &ApproveJoinMe{}}
	err := json.Unmarshal(dataBytes, approveJoin)
	if err != nil {
		log.Error("HandleApproveJoinMe: unable to unmarshal", "e", err)
		return err
	}
	approveJoinMe := approveJoin.Data.(*ApproveJoinMe)

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	ptt := pm.myPtt
	service := pm.Entity().Service().(*Backend)
	spm := service.SPM().(*ServiceProtocolManager)
	MyID := ptt.GetMyEntity().GetID()

	// new me
	newMyInfo := approveJoinMe.MyInfo

	// my-node

	myNode, err := NewMyNode(ts, joinRequest.ID, peer.GetID(), 0)
	if err != nil {
		return err
	}
	myNode.Status = types.StatusAlive
	_, err = myNode.Save()
	log.Debug("HandleApproveJoinMe: after myNode.Save", "myNode", myNode, "ID", myNode.ID, "e", err)
	if err != nil {
		return err
	}

	myNodeID := pm.myPtt.MyNodeID()
	myNode2, err := NewMyNode(ts, joinRequest.ID, myNodeID, 0)
	if err != nil {
		return err
	}
	myNode2.Status = types.StatusInit
	_, err = myNode2.Save()
	log.Debug("HandleApproveJoinMe: after myNode2.Save", "myNode2", myNode2, "ID", myNode2.ID, "e", err)
	if err != nil {
		return err
	}

	// my-info init
	newMyInfo.Status = types.StatusInit
	newMyInfo.UpdateTS = ts
	err = newMyInfo.Init(ptt, service, MyID)

	err = newMyInfo.Save()
	if err != nil {
		return err
	}

	// new op-key
	newMyOpKeyInfo := approveJoinMe.OpKeyInfo
	newMyOpKeyInfo.Init(pm.DBOpKeyInfo(), pm.DBObjLock())
	err = newMyOpKeyInfo.Save(false)
	if err != nil {
		return err
	}

	err = newMyInfo.PM().RegisterOpKeyInfo(newMyOpKeyInfo, false)
	if err != nil {
		return err
	}

	// add to entities
	err = spm.RegisterEntity(newMyInfo.ID, newMyInfo)
	if err != nil {
		return err
	}

	// me-start
	newMyInfo.Start()

	// remove join-me-request
	pm.lockJoinMeRequest.Lock()
	defer pm.lockJoinMeRequest.Unlock()
	delete(pm.joinMeRequests, *joinRequest.Hash)

	log.Debug("HandleApproveJoineMe: done")

	return nil
}
