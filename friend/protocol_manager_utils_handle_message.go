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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleMessage(op pkgservice.OpType, dataBytes []byte, peer *pkgservice.PttPeer) error {

	log.Debug("friend.HandleMessage: start", "op", op, "AddFriendOplogMsg", AddFriendOplogMsg)

	var err error
	switch op {
	// friend oplog
	case SyncFriendOplogMsg:
		err = pm.HandleSyncFriendOplog(dataBytes, peer)
	case SyncFriendOplogAckMsg:
		err = pm.HandleSyncFriendOplogAck(dataBytes, peer)
	case SyncFriendOplogNewOplogsMsg:
		err = pm.HandleSyncNewFriendOplog(dataBytes, peer)
	case SyncFriendOplogNewOplogsAckMsg:
		err = pm.HandleSyncNewFriendOplogAck(dataBytes, peer)
	case SyncPendingFriendOplogMsg:
		err = pm.HandleSyncPendingFriendOplog(dataBytes, peer)
	case SyncPendingFriendOplogAckMsg:
		err = pm.HandleSyncPendingFriendOplogAck(dataBytes, peer)

	case AddFriendOplogMsg:
		err = pm.HandleAddFriendOplog(dataBytes, peer)
	case AddFriendOplogsMsg:
		err = pm.HandleAddFriendOplogs(dataBytes, peer)
	case AddPendingFriendOplogMsg:
		err = pm.HandleAddPendingFriendOplog(dataBytes, peer)
	case AddPendingFriendOplogsMsg:
		err = pm.HandleAddPendingFriendOplogs(dataBytes, peer)

	// init friend info
	case InitFriendInfoMsg:
		err = pm.HandleInitFriendInfo(dataBytes, peer)
	case InitFriendInfoAckMsg:
		err = pm.HandleInitFriendInfoAck(dataBytes, peer)

	// message
	case SyncCreateMessageMsg:
		err = pm.HandleSyncCreateMessage(dataBytes, peer, SyncCreateMessageAckMsg)
	case SyncCreateMessageAckMsg:
		err = pm.HandleSyncCreateMessageAck(dataBytes, peer)
	case SyncCreateMessageBlockMsg:
		err = pm.HandleSyncCreateMessageBlock(dataBytes, peer)
	case SyncCreateMessageBlockAckMsg:
		err = pm.HandleSyncCreateMessageBlockAck(dataBytes, peer)

	default:
		log.Error("invalid op", "op", op, "InitFriendInfoMsg", InitFriendInfoMsg)
		err = pkgservice.ErrInvalidMsgCode
	}

	return err
}
