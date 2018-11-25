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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type InternalSyncFriendAck struct {
	LogID *types.PttID `json:"l"`

	InitFriendInfoAck *friend.InitFriendInfoAck `json:"i"`
}

func (pm *ProtocolManager) InternalSyncFriend(
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	syncID := &pkgservice.SyncID{ID: oplog.ObjID, LogID: oplog.ID}

	return pm.SendDataToPeer(InternalSyncFriendMsg, syncID, peer)
}

func (pm *ProtocolManager) HandleInternalSyncFriend(
	dataBytes []byte,
	peer *pkgservice.PttPeer,
) error {

	syncID := &pkgservice.SyncID{}
	err := json.Unmarshal(dataBytes, syncID)
	if err != nil {
		return err
	}

	friendSPM := pm.Entity().Service().(*Backend).friendBackend.SPM()
	f := friendSPM.Entity(syncID.ID)
	if f == nil {
		return types.ErrInvalidID
	}
	friendPM := f.PM().(*friend.ProtocolManager)

	initFriendInfoAck, err := friendPM.InitFriendInfoAckCore(peer)

	ackData := &InternalSyncFriendAck{LogID: syncID.LogID, InitFriendInfoAck: initFriendInfoAck}

	pm.SendDataToPeer(InternalSyncFriendAckMsg, ackData, peer)

	return nil
}

func (pm *ProtocolManager) HandleInternalSyncFriendAck(
	dataBytes []byte,
	peer *pkgservice.PttPeer,

) error {

	// unmarshal data
	theProfileData := account.NewEmptyApproveJoinProfile()
	theBoardData := content.NewEmptyApproveJoinBoard()
	theFriendData := friend.NewEmptyApproveJoinFriend()
	initFriendInfoAck := &friend.InitFriendInfoAck{ProfileData: theProfileData, FriendData: theFriendData, BoardData: theBoardData}

	data := &InternalSyncFriendAck{InitFriendInfoAck: initFriendInfoAck}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	// oplog
	oplog := &pkgservice.BaseOplog{ID: data.LogID}

	err = oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	pm.SetMeDB(oplog)
	err = oplog.Get(data.LogID, true)
	if oplog.IsSync {
		return nil
	}

	return types.ErrNotImplemented
}

func (pm *ProtocolManager) handleInternalSyncFriendAckNew(
	data *InternalSyncFriendAck,
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	ptt := pm.Ptt()
	friendService := pm.Entity().Service().(*Backend).friendBackend
	friendSPM := friendService.SPM()

	theFriendData := data.InitFriendInfoAck.FriendData

	f := theFriendData.Entity.(*friend.Friend)

	f.Status = types.StatusInit

	err := f.Init(ptt, friendService, friendSPM)
	err = f.Save(false)
	if err != nil {
		return err
	}
	friendSPM.RegisterEntity(f.ID, f)
	f.PrestartAndStart()

	friendPM := f.PM().(*friend.ProtocolManager)

	return friendPM.HandleInitFriendInfoAckCore(data.InitFriendInfoAck, oplog, peer, true)
}
