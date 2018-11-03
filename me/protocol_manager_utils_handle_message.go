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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleMessage(op pkgservice.OpType, dataBytes []byte, peer *pkgservice.PttPeer) (err error) {

	myInfo := pm.Entity().(*MyInfo)
	log.Debug("HandleMessage: Received msg", "myID", myInfo.ID, "op", op, "SyncMeOplogMsg", SyncMeOplogMsg)

	switch op {
	// init me info
	case InitMeInfoMsg:
		return pm.HandleInitMeInfo(dataBytes, peer)
	case InitMeInfoAckMsg:
		return pm.HandleInitMeInfoAck(dataBytes, peer)
	case InitMeInfoSyncMsg:
		return pm.HandleInitMeInfoSync(dataBytes, peer)

	// raft
	case SendRaftMsgsMsg:
		return pm.HandleSendRaftMsgs(dataBytes, peer)
	}

	fitPeerType := pm.GetPeerType(peer)

	if fitPeerType < pkgservice.PeerTypeMember {
		return pkgservice.ErrInvalidEntity
	}

	switch op {
	// me oplog
	case SyncMeOplogMsg:
		err = pm.HandleSyncMeOplog(dataBytes, peer)
	case SyncMeOplogAckMsg:
		err = pm.HandleSyncMeOplogAck(dataBytes, peer)
	case SyncMeOplogNewOplogsMsg:
		err = pm.HandleSyncNewMeOplog(dataBytes, peer)
	case SyncMeOplogNewOplogsAckMsg:
		err = pm.HandleSyncNewMeOplogAck(dataBytes, peer)
	case SyncPendingMeOplogMsg:
		err = pm.HandleSyncPendingMeOplog(dataBytes, peer)
	case SyncPendingMeOplogAckMsg:
		err = pm.HandleSyncPendingMeOplogAck(dataBytes, peer)

	case AddMeOplogMsg:
		err = pm.HandleAddMeOplog(dataBytes, peer)
	case AddMeOplogsMsg:
		err = pm.HandleAddMeOplogs(dataBytes, peer)
	case AddPendingMeOplogMsg:
		err = pm.HandleAddPendingMeOplog(dataBytes, peer)
	case AddPendingMeOplogsMsg:
		err = pm.HandleAddPendingMeOplogs(dataBytes, peer)

	default:
		err = pkgservice.ErrInvalidMsgCode
	}
	if err != nil {
		log.Error("HandleMessage: unable to handle message", "op", op, "peer", peer, "e", err)
	}

	log.Debug("HandleMessage: done")

	return
}
