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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/log"
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
	log.Debug("InternalSyncFriend: to SendDataToPeer", "syncID", syncID, "peer", peer)

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
	log.Debug("HandleInternalSyncFriend: after InitFriendInfoAckCore", "e", err)

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
	pm.SetMeDB(oplog)

	opData := &MeOpEntity{}
	err = oplog.GetData(opData)
	if err != nil {
		return err
	}

	// lock
	err = oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	// get
	err = oplog.Get(data.LogID, true)
	log.Debug("HandleInternalSyncFriendAck: after oplog.Get", "e", err, "isSync", oplog.IsSync)
	if oplog.IsSync {
		return nil
	}

	// lock entity
	friendBackend := pm.Entity().Service().(*Backend).friendBackend
	friendSPM := friendBackend.SPM().(*friend.ServiceProtocolManager)

	err = friendSPM.Lock(oplog.ObjID)
	if err != nil {
		return err
	}
	defer friendSPM.Unlock(oplog.ObjID)

	theFriend := friendSPM.Entity(oplog.ObjID)
	log.Debug("HandleInternalSyncFriendAck: after get theFriend", "theFriend", theFriend)
	if theFriend == nil {
		err = pm.handleInternalSyncFriendAckNew(friendBackend, friendSPM, data, oplog, peer)
		if err != nil {
			return err
		}

		oplog.IsSync = true
		oplog.Save(true, pm.meOplogMerkle)

		return nil
	}
	f, ok := theFriend.(*friend.Friend)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// exists
	friendStatus := f.Status

	switch {
	case friendStatus == types.StatusAlive && reflect.DeepEqual(f.LogID, opData.LogID):
		err = pm.handleInternalSyncEntityAckSameLog(f, oplog, peer)
	case friendStatus >= types.StatusTerminal:
	case friendStatus == types.StatusAlive:
		err = pm.handleInternalSyncEntityAckDiffAliveLog(f, oplog, peer)
	default:
		err = pm.handleInternalSyncFriendAckDiffLog(f, data, oplog, peer)
	}

	oplog.IsSync = true
	oplog.Save(true, pm.meOplogMerkle)
	return nil
}

func (pm *ProtocolManager) handleInternalSyncFriendAckNew(
	svc *friend.Backend,
	spm *friend.ServiceProtocolManager,
	data *InternalSyncFriendAck,
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	// from me.HandleApproveJoinFriend

	theFriendData := data.InitFriendInfoAck.FriendData

	// 1. new friend
	f := theFriendData.Entity.(*friend.Friend)

	f.Status = types.StatusInit

	ptt := pm.Ptt()
	err := f.Init(ptt, svc, spm)
	err = f.Save(true)
	if err != nil {
		return err
	}

	// save message-create-ts
	f.SaveMessageCreateTS(f.GetCreateTS())

	friendPM := f.PM().(*friend.ProtocolManager)

	// master logs
	log.Debug("handleInternalSyncFriendAckNew: to HandleMasterOplogs", "masterLogs", theFriendData.MasterLogs)
	friendPM.HandleMasterOplogs(theFriendData.MasterLogs, peer, false)

	// member logs
	log.Debug("HandleInternalSyncFriendAckNew: to HandleMemberOplogs", "memberLogs", theFriendData.MemberLogs)
	friendPM.HandleMemberOplogs(theFriendData.MemberLogs, peer, false)

	// oplog0
	oplog0 := theFriendData.Oplog0
	friendPM.SetLog0DB(oplog0)
	err = oplog0.Save(false, friendPM.Log0Merkle())
	if err != nil {
		return err
	}

	// new op-key
	newOpKey := theFriendData.OpKey
	newOpKey.Init(friendPM)
	err = newOpKey.Save(false)
	if err != nil {
		return err
	}

	opKeyLog := theFriendData.OpKeyLog
	friendPM.SetOpKeyDB(opKeyLog)
	err = opKeyLog.Save(false, nil)
	if err != nil {
		return err
	}

	err = friendPM.RegisterOpKey(newOpKey, false)
	if err != nil {
		return err
	}

	// register-peer
	friendPM.RegisterPendingPeer(peer, false)

	// add to entity
	spm.RegisterEntity(f.ID, f)

	// start
	f.PrestartAndStart()

	log.Debug("HandleInternalSyncFriendAckNew: to HandleInitFriendInfoAckCore")

	err = friendPM.HandleInitFriendInfoAckCore(data.InitFriendInfoAck, oplog, peer, true, true)
	log.Debug("HandleInternalSyncFriendAckNew: after HandleInitFriendInfoAckCore", "e", err)
	if err != nil {
		return err
	}
	return nil
}

func (pm *ProtocolManager) handleInternalSyncFriendAckDiffLog(
	f *friend.Friend,
	data *InternalSyncFriendAck,
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	friendPM := f.PM().(*friend.ProtocolManager)

	err := friendPM.HandleInitFriendInfoAckCore(data.InitFriendInfoAck, oplog, peer, false, true)
	log.Debug("HandleInternalSyncFriendAckDiffLog: after HandleInitFriendInfoAckCore", "e", err)
	if err != nil {
		return err
	}

	return nil
}
