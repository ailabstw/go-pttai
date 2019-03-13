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

	// master oplog
	case SyncMasterOplogMsg:
		return pm.HandleSyncMasterOplog(dataBytes, peer)

	case ForceSyncMasterOplogByMerkleMsg:
		return pm.HandleForceSyncMasterOplogByMerkle(dataBytes, peer)
	case ForceSyncMasterOplogByMerkleAckMsg:
		return pm.HandleForceSyncMasterOplogByMerkleAck(dataBytes, peer)
	case ForceSyncMasterOplogByOplogAckMsg:
		return pm.HandleForceSyncMasterOplogByOplogAck(dataBytes, peer)
	case InvalidSyncMasterOplogMsg:
		return pm.HandleSyncMasterOplogInvalid(dataBytes, peer)

	case SyncMasterOplogAckMsg:
		return pm.HandleSyncMasterOplogAck(dataBytes, peer)
	case SyncMasterOplogNewOplogsMsg:
		return pm.HandleSyncNewMasterOplog(dataBytes, peer)
	case SyncMasterOplogNewOplogsAckMsg:
		return pm.HandleSyncNewMasterOplogAck(dataBytes, peer)
	case SyncPendingMasterOplogMsg:
		return pm.HandleSyncPendingMasterOplog(dataBytes, peer)
	case SyncPendingMasterOplogAckMsg:
		return pm.HandleSyncPendingMasterOplogAck(dataBytes, peer)

	case AddMasterOplogMsg:
		return pm.HandleAddMasterOplog(dataBytes, peer)
	case AddMasterOplogsMsg:
		return pm.HandleAddMasterOplogs(dataBytes, peer)
	case AddPendingMasterOplogMsg:
		return pm.HandleAddPendingMasterOplog(dataBytes, peer)
	case AddPendingMasterOplogsMsg:
		return pm.HandleAddPendingMasterOplogs(dataBytes, peer)

	// member oplog
	case SyncMemberOplogMsg:
		return pm.HandleSyncMemberOplog(dataBytes, peer)

	case ForceSyncMemberOplogByMerkleMsg:
		return pm.HandleForceSyncMemberOplogByMerkle(dataBytes, peer)
	case ForceSyncMemberOplogByMerkleAckMsg:
		return pm.HandleForceSyncMemberOplogByMerkleAck(dataBytes, peer)
	case ForceSyncMemberOplogByOplogAckMsg:
		return pm.HandleForceSyncMemberOplogByOplogAck(dataBytes, peer)
	case InvalidSyncMemberOplogMsg:
		return pm.HandleSyncMemberOplogInvalid(dataBytes, peer)

	case SyncMemberOplogAckMsg:
		return pm.HandleSyncMemberOplogAck(dataBytes, peer)
	case SyncMemberOplogNewOplogsMsg:
		return pm.HandleSyncNewMemberOplog(dataBytes, peer)
	case SyncMemberOplogNewOplogsAckMsg:
		return pm.HandleSyncNewMemberOplogAck(dataBytes, peer)
	case SyncPendingMemberOplogMsg:
		return pm.HandleSyncPendingMemberOplog(dataBytes, peer)
	case SyncPendingMemberOplogAckMsg:
		return pm.HandleSyncPendingMemberOplogAck(dataBytes, peer)

	case AddMemberOplogMsg:
		return pm.HandleAddMemberOplog(dataBytes, peer)
	case AddMemberOplogsMsg:
		return pm.HandleAddMemberOplogs(dataBytes, peer)
	case AddPendingMemberOplogMsg:
		return pm.HandleAddPendingMemberOplog(dataBytes, peer)
	case AddPendingMemberOplogsMsg:
		return pm.HandleAddPendingMemberOplogs(dataBytes, peer)

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

	var origPeer *PttPeer
	if fitPeerType < PeerTypePending {
		origPeer = pm.Peers().Peer(peer.GetID(), false)
		if origPeer != nil {
			pm.UnregisterPeer(peer, false, false, false)
		}
		return ErrInvalidEntity
	}

	origPeer = pm.Peers().GetPeerWithPeerType(peer.GetID(), fitPeerType, false)

	if origPeer == nil {
		pm.RegisterPeer(peer, fitPeerType, false)
	}

	return pm.HandleMessage(op, dataBytes, peer)
}
