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

type InitFriendInfo struct {
	ProfileData *account.ApproveJoinEntity    `json:"P"`
	BoardData   *pkgservice.ApproveJoinEntity `json:"b"`
}

func (pm *ProtocolManager) InitFriendInfo(peer *pkgservice.PttPeer) error {
	f := pm.Entity().(*Friend)

	log.Debug("InitFriendInfo: start", "peer", peer, "userID", peer)
	friendID := f.FriendID

	// profile
	profilePM := pm.Ptt().GetMyEntity().GetProfile().(*account.Profile).PM()
	joinEntity := &pkgservice.JoinEntity{ID: friendID}
	_, theProfileData, err := profilePM.ApproveJoin(joinEntity, nil, peer)
	log.Debug("InitFriendInfo: after profile ApproveJoin", "e", err, "profile", profilePM.Entity().GetID())
	if err != nil {
		return err
	}
	profileData, ok := theProfileData.(*account.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// board
	boardPM := pm.Ptt().GetMyEntity().GetBoard().(*content.Board).PM()
	_, theBoardData, err := boardPM.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return err
	}
	boardData, ok := theBoardData.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// to send data
	initFriendInfo := &InitFriendInfo{
		ProfileData: profileData,
		BoardData:   boardData,
	}

	err = pm.SendDataToPeer(InitFriendInfoMsg, initFriendInfo, peer)
	log.Debug("InitFriendInfo: after send data", "e", err, "entity", pm.Entity().GetID())
	if err != nil {
		return err
	}

	log.Debug("InitFriendInfo: done")

	return nil
}

func (pm *ProtocolManager) HandleInitFriendInfo(dataBytes []byte, peer *pkgservice.PttPeer) error {

	theProfileData := account.NewEmptyApproveJoinProfile()
	theBoardData := content.NewEmptyApproveJoinBoard()
	data := &InitFriendInfo{ProfileData: theProfileData, BoardData: theBoardData}
	err := json.Unmarshal(dataBytes, data)
	log.Debug("HandleInitFriendInfo: after unmarshal", "e", err)
	if err != nil {
		return err
	}
	profileData := data.ProfileData
	boardData := data.BoardData

	log.Debug("HandleInitFriendInfo: start", "peer", peer.RemoteAddr(), "userID", peer.UserID)

	// validate
	f := pm.Entity().(*Friend)

	if peer.PeerType != pkgservice.PeerTypeMe {
		if !reflect.DeepEqual(profileData.MyID, peer.UserID) {
			return pkgservice.ErrInvalidData
		}

		if !reflect.DeepEqual(f.FriendID, peer.UserID) {
			return types.ErrInvalidID
		}
	}

	// profile
	profileSPM := pm.Entity().Service().(*Backend).accountBackend.SPM().(*account.ServiceProtocolManager)

	theProfile, err := profileSPM.CreateJoinEntity(profileData, peer, nil, true, true, true, false, true)
	log.Debug("HandleInitFriendInfo: after profile create join entity", "e", err)
	if err != nil {
		return err
	}
	profile := theProfile.(*account.Profile)
	profile.PM().RegisterPeer(peer, pkgservice.PeerTypeImportant)

	// content
	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM().(*content.ServiceProtocolManager)
	theBoard, err := contentSPM.CreateJoinEntity(boardData, peer, nil, true, true, true, false, true)
	log.Debug("HandleInitFriendInfo: after board create join entity", "e", err)
	if err != nil {
		return err
	}
	board := theBoard.(*content.Board)
	board.PM().RegisterPeer(peer, pkgservice.PeerTypeImportant)

	// f
	f.ProfileID = profile.ID
	f.Profile = profile
	f.BoardID = board.ID
	f.Board = board
	f.Status = types.StatusAlive

	err = f.Save(false)
	if err != nil {
		log.Error("HandleInitFriendInfo: unable to save", "e", err)
		return err
	}

	pm.postcreateFriend(f)

	// ack
	err = pm.InitFriendInfoAck(peer)

	log.Debug("HandleInitFriendInfo: end")

	return nil
}

func (pm *ProtocolManager) postcreateFriend(entity pkgservice.Entity) error {

	// me-oplog
	err := pm.Ptt().GetMyEntity().CreateEntityOplog(entity)

	if err != nil {
		return err
	}

	// ptt-oplog
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	myID := pm.Ptt().GetMyEntity().GetID()

	f := entity.(*Friend)

	oplog, err := pkgservice.NewPttOplog(entity.GetID(), ts, f.FriendID, pkgservice.PttOpTypeCreateFriend, pkgservice.PttOpTypeCreateFriend, myID)
	if err != nil {
		return err
	}
	err = oplog.Save(false)
	if err != nil {
		return err
	}

	return nil
}
