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

package service

import (
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/log"
)

func PMHandleMessageWrapper(pm ProtocolManager, hash *common.Address, encData []byte, peer *PttPeer) error {
	opKeyInfo, err := pm.GetOpKeyFromHash(hash, false)
	log.Debug("PMHandleMessageWrapper: after GetOpKeyInfoFromHash", "e", err)
	if err != nil {
		return err
	}

	op, dataBytes, err := pm.Ptt().DecryptData(encData, opKeyInfo)
	log.Debug("PMHandleMessageWrapper: after DecryptData", "e", err, "op", op, "NMsg", NMsg)
	if err != nil {
		return err
	}

	switch op {
	case IdentifyPeerMsg:
		return pm.HandleIdentifyPeer(dataBytes, peer)
	case IdentifyPeerAckMsg:
		return pm.HandleIdentifyPeerAck(dataBytes, peer)

	// me oplog
	case SyncMasterOplogMsg:
		err = pm.HandleSyncMasterOplog(dataBytes, peer)
	case SyncMasterOplogAckMsg:
		err = pm.HandleSyncMasterOplogAck(dataBytes, peer)
	case SyncMasterOplogNewOplogsMsg:
		err = pm.HandleSyncNewMasterOplog(dataBytes, peer)
	case SyncMasterOplogNewOplogsAckMsg:
		err = pm.HandleSyncNewMasterOplogAck(dataBytes, peer)
	case SyncPendingMasterOplogMsg:
		err = pm.HandleSyncPendingMasterOplog(dataBytes, peer)
	case SyncPendingMasterOplogAckMsg:
		err = pm.HandleSyncPendingMasterOplogAck(dataBytes, peer)

	case AddMasterOplogMsg:
		err = pm.HandleAddMasterOplog(dataBytes, peer)
	case AddMasterOplogsMsg:
		err = pm.HandleAddMasterOplogs(dataBytes, peer)
	case AddPendingMasterOplogMsg:
		err = pm.HandleAddPendingMasterOplog(dataBytes, peer)
	case AddPendingMasterOplogsMsg:
		err = pm.HandleAddPendingMasterOplogs(dataBytes, peer)

	// me oplog
	case SyncMemberOplogMsg:
		err = pm.HandleSyncMemberOplog(dataBytes, peer)
	case SyncMemberOplogAckMsg:
		err = pm.HandleSyncMemberOplogAck(dataBytes, peer)
	case SyncMemberOplogNewOplogsMsg:
		err = pm.HandleSyncNewMemberOplog(dataBytes, peer)
	case SyncMemberOplogNewOplogsAckMsg:
		err = pm.HandleSyncNewMemberOplogAck(dataBytes, peer)
	case SyncPendingMemberOplogMsg:
		err = pm.HandleSyncPendingMemberOplog(dataBytes, peer)
	case SyncPendingMemberOplogAckMsg:
		err = pm.HandleSyncPendingMemberOplogAck(dataBytes, peer)

	case AddMemberOplogMsg:
		err = pm.HandleAddMemberOplog(dataBytes, peer)
	case AddMemberOplogsMsg:
		err = pm.HandleAddMemberOplogs(dataBytes, peer)
	case AddPendingMemberOplogMsg:
		err = pm.HandleAddPendingMemberOplog(dataBytes, peer)
	case AddPendingMemberOplogsMsg:
		err = pm.HandleAddPendingMemberOplogs(dataBytes, peer)

	// op-key-oplog
	case SyncOpKeyOplogMsg:
		return pm.HandleSyncOpKeyOplog(dataBytes, peer, SyncOpKeyOplogMsg)
	case SyncOpKeyOplogAckMsg:
		return pm.HandleSyncOpKeyOplog(dataBytes, peer, SyncOpKeyOplogAckMsg)
	case SyncPendingOpKeyOplogMsg:
		return pm.HandleSyncPendingOpKeyOplog(dataBytes, peer)
	case SyncPendingOpKeyOplogAckMsg:
		return pm.HandleSyncPendingOpKeyOplogAck(dataBytes, peer)

	case AddOpKeyOplogMsg:
		return pm.HandleAddOpKeyOplog(dataBytes, peer)
	case AddOpKeyOplogsMsg:
		return pm.HandleAddOpKeyOplogs(dataBytes, peer)
	case AddPendingOpKeyOplogMsg:
		return pm.HandleAddPendingOpKeyOplog(dataBytes, peer)
	case AddPendingOpKeyOplogsMsg:
		return pm.HandleAddPendingOpKeyOplogs(dataBytes, peer)

	// sync create-op-key
	case SyncCreateOpKeyMsg:
		return pm.HandleSyncCreateOpKey(dataBytes, peer)
	case SyncCreateOpKeyAckMsg:
		return pm.HandleSyncCreateOpKeyAck(dataBytes, peer)

	}

	log.Debug("PMHandleMessageWrapper: to GetPeerType", "peer", peer, "entity", pm.Entity().GetID())

	fitPeerType := pm.GetPeerType(peer)

	log.Debug("PMHandleMessageWrapper: after GetPeerType", "peer", peer, "entity", pm.Entity().GetID(), "fitPeerType", fitPeerType)

	if fitPeerType < PeerTypePending {
		return ErrInvalidEntity
	}

	return pm.HandleMessage(op, dataBytes, peer)
}
