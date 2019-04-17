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

package me

import (
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ApproveJoinFriend struct {
	FriendData *friend.ApproveJoin `json:"F"`
}

func (pm *ProtocolManager) ApproveJoinFriend(joinEntity *pkgservice.JoinEntity, keyInfo *pkgservice.KeyInfo, peer *pkgservice.PttPeer) (*pkgservice.KeyInfo, interface{}, error) {

	friendSPM := pm.Entity().Service().(*Backend).friendBackend.SPM().(*friend.ServiceProtocolManager)

	// Reset friend
	origFriend, err := friendSPM.GetFriendEntityByFriendID(joinEntity.ID)
	log.Debug("ApproveJoinFriend: after GetFriendEntityByFriendID", "e", err, "origFriend", origFriend, "friendID", joinEntity.ID)
	if err == nil && origFriend.Status >= types.StatusDeleted {
		origFriend.PM().FullCleanLog()
		err = friendSPM.UnregisterEntity(origFriend.ID)
		log.Debug("ApproveJoinFriend: after UnregisterEntity", "e", err)
	}

	// create friend

	theFriend, err := friendSPM.CreateFriend(joinEntity.ID)
	log.Debug("ApproveJoinFriend: after CreateFriend", "e", err)
	if err != nil {
		return nil, nil, err
	}

	// save message-create-ts
	theFriend.SaveMessageCreateTS(theFriend.GetCreateTS())

	// get friend-key and friend oplog
	friendPM := theFriend.PM().(*friend.ProtocolManager)
	friendOpKeyInfo, friendData, err := friendPM.ApproveJoinFriend(joinEntity, keyInfo, peer)
	log.Debug("ApproveJoinFriend: after friend.ApproveJoinFriend", "e", err)
	if err != nil {
		return nil, nil, err
	}

	// register pending peer
	if peer.UserID == nil {
		peer.UserID = joinEntity.ID
		friendPM.Ptt().FinishIdentifyPeer(peer, false, false)
	}
	friendPM.RegisterPendingPeer(peer, false)

	data := &ApproveJoinFriend{
		FriendData: friendData.(*friend.ApproveJoin),
	}

	return friendOpKeyInfo, data, nil
}

func (pm *ProtocolManager) HandleApproveJoinFriend(dataBytes []byte, joinRequest *pkgservice.JoinRequest, peer *pkgservice.PttPeer) error {

	log.Debug("HandleApproveJoinFriend: start")

	approveJoin := &pkgservice.ApproveJoin{Data: &ApproveJoinFriend{}}
	err := json.Unmarshal(dataBytes, approveJoin)
	if err != nil {
		log.Error("HandleApproveJoinFriend: unable to unmarshal", "e", err)
		return err
	}

	approveJoinFriend, ok := approveJoin.Data.(*ApproveJoinFriend)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	ptt := pm.Ptt()
	myID := ptt.GetMyEntity().GetID()

	friendService := pm.Entity().Service().(*Backend).friendBackend
	friendSPM := friendService.SPM().(*friend.ServiceProtocolManager)

	// friendID
	var friendID *types.PttID

	friendData := approveJoinFriend.FriendData
	f := friendData.Friend
	if reflect.DeepEqual(myID, f.Friend0ID) {
		friendID = f.Friend1ID
	} else {
		friendID = f.Friend0ID
	}

	// Reset friend
	origFriend, err := friendSPM.GetFriendEntityByFriendID(friendID)
	log.Debug("HandleApproveJoinFriend: after GetFriendEntityByFriendID", "e", err, "origFriend", origFriend, "friendID", friendID)
	if err == nil && origFriend.Status >= types.StatusDeleted {
		origFriend.PM().FullCleanLog()
		err = friendSPM.UnregisterEntity(origFriend.ID)
		log.Debug("HandleApproveJoinFriend: after UnregisterEntity", "e", err)
	}

	// 1. new friend
	f.FriendID = friendID

	f.Status = types.StatusInit
	f.UpdateTS = types.ZeroTimestamp
	f.OwnerIDs = []*types.PttID{myID}

	err = f.Init(ptt, friendService, friendSPM)
	err = f.Save(false)
	if err != nil {
		return err
	}

	// save message-create-ts
	f.SaveMessageCreateTS(f.GetCreateTS())

	newPM := f.PM().(*friend.ProtocolManager)

	// master logs
	log.Debug("HandleApproveJoinFriend: to HandleMasterOplogs", "masterLogs", friendData.MasterLogs)
	newPM.HandleMasterOplogs(friendData.MasterLogs, peer, false)

	// member logs
	log.Debug("HandleApproveJoinFriend: to HandleMemberOplogs", "memberLogs", friendData.MemberLogs)
	newPM.HandleMemberOplogs(friendData.MemberLogs, peer, false)

	// oplog0
	oplog0 := friendData.Oplog0
	newPM.SetLog0DB(oplog0)
	err = oplog0.Save(false, newPM.Log0Merkle())
	if err != nil {
		return err
	}

	// new op-key

	newOpKey := friendData.OpKey
	newOpKey.Init(newPM)
	err = newOpKey.Save(false)
	if err != nil {
		return err
	}

	opKeyLog := friendData.OpKeyLog
	newPM.SetOpKeyDB(opKeyLog)
	err = opKeyLog.Save(false, nil)
	if err != nil {
		return err
	}

	err = newPM.RegisterOpKey(newOpKey, false)
	if err != nil {
		return err
	}

	// register-peer
	if peer.UserID == nil {
		peer.UserID = f.FriendID
		newPM.Ptt().FinishIdentifyPeer(peer, false, false)
	}
	newPM.RegisterPendingPeer(peer, false)

	// add to entity
	friendSPM.RegisterEntity(f.ID, f)

	// start
	f.PrestartAndStart()

	// remove joinFriendRequest
	pm.lockJoinFriendRequest.Lock()
	defer pm.lockJoinFriendRequest.Unlock()

	delete(pm.joinFriendRequests, *joinRequest.Hash)

	// init-friend-info
	log.Debug("HandleApproveJoinFriend: to InitFriendInfo", "f", f.ID)
	err = newPM.InitFriendInfo(peer)
	if err != nil {
		return nil
	}

	return nil
}
