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
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func NewEmptyApproveJoinFriend() *pkgservice.ApproveJoinEntity {
	return &pkgservice.ApproveJoinEntity{Entity: NewEmptyFriend()}
}

type InitFriendInfoAck struct {
	FriendData  *pkgservice.ApproveJoinEntity `json:"F"`
	ProfileData *account.ApproveJoinEntity    `json:"P"`
	BoardData   *pkgservice.ApproveJoinEntity `json:"b"`
}

func (pm *ProtocolManager) InitFriendInfoAck(peer *pkgservice.PttPeer) error {

	// init friend info-ack
	initFriendInfoAck, err := pm.InitFriendInfoAckCore(peer)
	if err != nil {
		return err
	}

	// register-peer
	pm.RegisterPeer(peer, pkgservice.PeerTypeMember)

	// add-master
	f := pm.Entity().(*Friend)

	pm.AddMaster(f.FriendID, true)

	// send-data-to-peer

	log.Debug("InitFriendInfoAck: to SendDataToPeer", "peer", peer)

	err = pm.SendDataToPeer(InitFriendInfoAckMsg, initFriendInfoAck, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) InitFriendInfoAckCore(peer *pkgservice.PttPeer) (*InitFriendInfoAck, error) {

	f := pm.Entity().(*Friend)

	// friend-data
	joinEntity := &pkgservice.JoinEntity{ID: f.FriendID}
	_, theFriendData, err := pm.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return nil, err
	}
	friendData, ok := theFriendData.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// profile-data
	profilePM := pm.Ptt().GetMyEntity().GetProfile().(*account.Profile).PM()
	_, theProfileData, err := profilePM.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return nil, err
	}
	profileData, ok := theProfileData.(*account.ApproveJoinEntity)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// board-data
	boardPM := pm.Ptt().GetMyEntity().GetBoard().(*content.Board).PM()
	_, theBoardData, err := boardPM.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return nil, err
	}
	boardData, ok := theBoardData.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	initFriendInfoAck := &InitFriendInfoAck{
		FriendData:  friendData,
		ProfileData: profileData,
		BoardData:   boardData,
	}

	return initFriendInfoAck, nil
}

func (pm *ProtocolManager) HandleInitFriendInfoAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	log.Debug("HandleInitFriendInfoAck: start")

	theProfileData := account.NewEmptyApproveJoinProfile()
	theBoardData := content.NewEmptyApproveJoinBoard()
	theFriendData := NewEmptyApproveJoinFriend()
	initFriendInfoAck := &InitFriendInfoAck{ProfileData: theProfileData, FriendData: theFriendData, BoardData: theBoardData}
	err := json.Unmarshal(dataBytes, initFriendInfoAck)
	if err != nil {
		return err
	}

	return pm.HandleInitFriendInfoAckCore(initFriendInfoAck, nil, peer, true, false)
}

func (pm *ProtocolManager) HandleInitFriendInfoAckCore(
	initFriendInfoAck *InitFriendInfoAck,
	oplog *pkgservice.BaseOplog,

	peer *pkgservice.PttPeer,

	isNew bool,
	isLocked bool,
) error {

	profileData, friendData, boardData := initFriendInfoAck.ProfileData, initFriendInfoAck.FriendData, initFriendInfoAck.BoardData

	f := pm.Entity().(*Friend)
	spm := f.Service().SPM().(*ServiceProtocolManager)

	// validate
	if peer.PeerType != pkgservice.PeerTypeMe && !reflect.DeepEqual(f.FriendID, peer.UserID) {
		return types.ErrInvalidID
	}

	// lock
	if !isLocked {
		err := f.Lock()
		if err != nil {
			return err
		}
		defer f.Unlock()
	}

	// profile
	profileSPM := pm.Entity().Service().(*Backend).accountBackend.SPM().(*account.ServiceProtocolManager)
	log.Debug("HandleInitFriendInfoAckCore: to CreateJoinProfile", "isNew", isNew)
	theProfile, err := profileSPM.CreateJoinEntity(profileData, peer, nil, isNew, isNew, true, false, true)
	if err != nil {
		return err
	}
	profile := theProfile.(*account.Profile)
	profile.PM().RegisterPeer(peer, pkgservice.PeerTypeImportant)

	// board
	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM().(*content.ServiceProtocolManager)
	theBoard, err := contentSPM.CreateJoinEntity(boardData, peer, nil, isNew, isNew, true, false, true)
	if err != nil {
		return err
	}
	board := theBoard.(*content.Board)
	board.PM().RegisterPeer(peer, pkgservice.PeerTypeImportant)

	// friend
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	f.ProfileID = profile.ID
	f.Profile = profile
	f.BoardID = board.ID
	f.Board = board
	f.UpdateTS = friendData.Entity.GetUpdateTS()

	friendData.Entity = f
	log.Debug("HandleInitFriendInfoAck: to CreateJoinFriend", "f", f.ID)
	_, err = spm.CreateJoinEntity(friendData, peer, nil, false, false, false, true, true)
	if err != nil {
		return err
	}

	// ptt-oplog

	myID := pm.Ptt().GetMyEntity().GetID()

	pttOplog, err := pkgservice.NewPttOplog(f.GetID(), ts, f.FriendID, pkgservice.PttOpTypeCreateFriend, pkgservice.PttOpTypeCreateFriend, myID)
	if err != nil {
		return err
	}
	err = pttOplog.Save(false)
	if err != nil {
		return err
	}

	log.Debug("HandleInitFriendInfoAck: done")

	return nil
}
