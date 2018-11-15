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

package friend

import (
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func NewEmptyApproveJoinFriend() *pkgservice.ApproveJoinEntity {
	return &pkgservice.ApproveJoinEntity{Entity: NewEmptyFriend()}
}

type InitFriendInfoAck struct {
	FriendData  *pkgservice.ApproveJoinEntity `json:"F"`
	ProfileData *pkgservice.ApproveJoinEntity `json:"P"`
}

func (pm *ProtocolManager) InitFriendInfoAck(peer *pkgservice.PttPeer) error {
	f := pm.Entity().(*Friend)

	// register-peer
	log.Debug("InitFriendInfoAck: start", "peer", peer)
	pm.RegisterPeer(peer, pkgservice.PeerTypeMember)

	// friend-data
	joinEntity := &pkgservice.JoinEntity{ID: f.FriendID}
	_, theFriendData, err := pm.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return err
	}
	friendData, ok := theFriendData.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// profile-data
	profilePM := pm.Ptt().GetMyEntity().GetProfile().(*account.Profile).PM()
	_, theProfileData, err := profilePM.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return err
	}
	profileData, ok := theProfileData.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// add-master
	pm.AddMaster(f.FriendID, true)

	// send-data-to-peer
	initFriendInfoAck := &InitFriendInfoAck{
		FriendData:  friendData,
		ProfileData: profileData,
	}

	err = pm.SendDataToPeer(InitFriendInfoAckMsg, initFriendInfoAck, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) HandleInitFriendInfoAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	log.Debug("HandleInitFriendInfoAck: start")

	f := pm.Entity().(*Friend)
	spm := f.Service().SPM()

	// validate
	if !reflect.DeepEqual(f.FriendID, peer.UserID) {
		return types.ErrInvalidID
	}

	// lock
	err := f.Lock()
	if err != nil {
		return err
	}
	defer f.Unlock()

	theProfileData := account.NewEmptyApproveJoinProfile()
	theFriendData := NewEmptyApproveJoinFriend()
	initFriendInfoAck := &InitFriendInfoAck{ProfileData: theProfileData, FriendData: theFriendData}
	err = json.Unmarshal(dataBytes, initFriendInfoAck)
	if err != nil {
		return err
	}
	profileData := theProfileData
	friendData := theFriendData

	// profile
	profileSPM := pm.Entity().Service().(*Backend).accountBackend.SPM().(*account.ServiceProtocolManager)
	theProfile, err := profileSPM.CreateJoinEntity(profileData, peer, nil, true, true)
	if err != nil {
		return err
	}
	profile := theProfile.(*account.Profile)

	// friend
	f.ProfileID = profile.ID
	f.Profile = profile
	f.UpdateTS = friendData.Entity.GetUpdateTS()
	f.Status = types.StatusAlive

	friendData.Entity = f
	_, err = spm.CreateJoinEntity(friendData, peer, nil, false, false)
	if err != nil {
		return err
	}

	log.Debug("HandleInitFriendInfoAck: done")

	return nil
}
