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

package account

import (
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleMessage(op pkgservice.OpType, dataBytes []byte, peer *pkgservice.PttPeer) (err error) {

	log.Debug("account.HandleMessage: start", "op", op, "AddUserOplogMsg", AddUserOplogMsg)

	switch op {
	// user oplog
	case SyncUserOplogMsg:
		err = pm.HandleSyncUserOplog(dataBytes, peer)
	case ForceSyncUserOplogAckMsg:
		err = pm.HandleForceSyncUserOplogAck(dataBytes, peer)
	case SyncUserOplogAckMsg:
		err = pm.HandleSyncUserOplogAck(dataBytes, peer)
	case SyncUserOplogNewOplogsMsg:
		err = pm.HandleSyncNewUserOplog(dataBytes, peer)
	case SyncUserOplogNewOplogsAckMsg:
		err = pm.HandleSyncNewUserOplogAck(dataBytes, peer)
	case SyncPendingUserOplogMsg:
		err = pm.HandleSyncPendingUserOplog(dataBytes, peer)
	case SyncPendingUserOplogAckMsg:
		err = pm.HandleSyncPendingUserOplogAck(dataBytes, peer)

	case AddUserOplogMsg:
		err = pm.HandleAddUserOplog(dataBytes, peer)
	case AddUserOplogsMsg:
		err = pm.HandleAddUserOplogs(dataBytes, peer)
	case AddPendingUserOplogMsg:
		err = pm.HandleAddPendingUserOplog(dataBytes, peer)
	case AddPendingUserOplogsMsg:
		err = pm.HandleAddPendingUserOplogs(dataBytes, peer)

	// user-name
	case SyncCreateUserNameMsg:
		err = pm.HandleSyncCreateUserName(dataBytes, peer, SyncCreateUserNameAckMsg)
	case SyncCreateUserNameAckMsg:
		err = pm.HandleSyncCreateUserNameAck(dataBytes, peer)
	case SyncUpdateUserNameMsg:
		err = pm.HandleSyncUpdateUserName(dataBytes, peer, SyncUpdateUserNameAckMsg)
	case SyncUpdateUserNameAckMsg:
		err = pm.HandleSyncUpdateUserNameAck(dataBytes, peer)

	// user-img
	case SyncCreateUserImgMsg:
		err = pm.HandleSyncCreateUserImg(dataBytes, peer, SyncCreateUserImgAckMsg)
	case SyncCreateUserImgAckMsg:
		err = pm.HandleSyncCreateUserImgAck(dataBytes, peer)
	case SyncUpdateUserImgMsg:
		err = pm.HandleSyncUpdateUserImg(dataBytes, peer, SyncUpdateUserImgAckMsg)
	case SyncUpdateUserImgAckMsg:
		err = pm.HandleSyncUpdateUserImgAck(dataBytes, peer)

	// name-card
	case SyncCreateNameCardMsg:
		err = pm.HandleSyncCreateNameCard(dataBytes, peer, SyncCreateNameCardAckMsg)
	case SyncCreateNameCardAckMsg:
		err = pm.HandleSyncCreateNameCardAck(dataBytes, peer)
	case SyncUpdateNameCardMsg:
		err = pm.HandleSyncUpdateNameCard(dataBytes, peer, SyncUpdateNameCardAckMsg)
	case SyncUpdateNameCardAckMsg:
		err = pm.HandleSyncUpdateNameCardAck(dataBytes, peer)

	}
	return nil
}
