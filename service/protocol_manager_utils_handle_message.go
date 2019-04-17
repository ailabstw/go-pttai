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
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"
)

func PMHandleMessageWrapper(pm ProtocolManager, hash *common.Address, encData []byte, peer *PttPeer) error {
	opKeyInfo, err := pm.GetOpKeyFromHash(hash, false)
	if err != nil {
		log.Error("PMHandleMessageWrapper: unable to GetOpKeyFromHash", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		return err
	}

	op, dataBytes, err := pm.Ptt().DecryptData(encData, opKeyInfo)
	if err != nil {
		log.Error("PMHandleMessageWrapper: unable to DecryptData", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		return err
	}

	// handle identify-peer message

	switch op {
	case IdentifyPeerMsg:
		err = pm.HandleIdentifyPeer(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleIdentifyPeer", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
		return err
	case IdentifyPeerAckMsg:
		err = pm.HandleIdentifyPeerAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleIdentifyPeerAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
		return err
	}

	// check peer valid with the pm.
	fitPeerType := pm.GetPeerType(peer)

	var origPeer *PttPeer

	if fitPeerType < PeerTypePending {
		origPeer = pm.Peers().Peer(peer.GetID(), false)
		if origPeer != nil {
			pm.UnregisterPeer(peer, false, false, false)
		}
		return ErrInvalidEntity
	}

	// handle non-registtered message
	err = pm.HandleNonRegisteredMessage(op, dataBytes, peer)
	if err == nil {
		return nil
	}

	// check peer registered
	if !peer.IsRegistered {
		return ErrNotRegistered
	}

	origPeer = pm.Peers().GetPeerWithPeerType(peer.GetID(), fitPeerType, false)
	if origPeer == nil {
		pm.RegisterPeer(peer, fitPeerType, false)
	}

	// handle message

	switch op {
	// master oplog
	case SyncMasterOplogMsg:
		err = pm.HandleSyncMasterOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncMasterOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case ForceSyncMasterOplogByMerkleMsg:
		err = pm.HandleForceSyncMasterOplogByMerkle(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleForceSyncMasterOplogByMerkle", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case ForceSyncMasterOplogByMerkleAckMsg:
		err = pm.HandleForceSyncMasterOplogByMerkleAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleForceSyncMasterOplogByMerkleAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case ForceSyncMasterOplogByOplogAckMsg:
		err = pm.HandleForceSyncMasterOplogByOplogAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleForceSyncMasterOplogByOplogAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case InvalidSyncMasterOplogMsg:
		err = pm.HandleSyncMasterOplogInvalid(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncMasterOplogInvalidAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case ForceSyncMasterOplogMsg:
		err = pm.HandleForceSyncMasterOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleForceSyncMasterOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case ForceSyncMasterOplogAckMsg:
		err = pm.HandleForceSyncMasterOplogAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleForceSyncMasterOplogAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case SyncMasterOplogAckMsg:
		err = pm.HandleSyncMasterOplogAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncMasterOplogAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case SyncMasterOplogNewOplogsMsg:
		err = pm.HandleSyncNewMasterOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncNewMasterOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case SyncMasterOplogNewOplogsAckMsg:
		err = pm.HandleSyncNewMasterOplogAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncNewMasterOplogAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case SyncPendingMasterOplogMsg:
		err = pm.HandleSyncPendingMasterOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncPendingMasterOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case SyncPendingMasterOplogAckMsg:
		err = pm.HandleSyncPendingMasterOplogAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncPendingMasterOplogAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case AddMasterOplogMsg:
		err = pm.HandleAddMasterOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddMasterOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case AddMasterOplogsMsg:
		err = pm.HandleAddMasterOplogs(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddMasterOplogs", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case AddPendingMasterOplogMsg:
		err = pm.HandleAddPendingMasterOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddPendingMasterOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case AddPendingMasterOplogsMsg:
		err = pm.HandleAddPendingMasterOplogs(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddPendingMasterOplogs", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	// member oplog
	case SyncMemberOplogMsg:
		err = pm.HandleSyncMemberOplog(dataBytes, peer)

	case ForceSyncMemberOplogByMerkleMsg:
		err = pm.HandleForceSyncMemberOplogByMerkle(dataBytes, peer)
	case ForceSyncMemberOplogByMerkleAckMsg:
		err = pm.HandleForceSyncMemberOplogByMerkleAck(dataBytes, peer)
	case ForceSyncMemberOplogByOplogAckMsg:
		err = pm.HandleForceSyncMemberOplogByOplogAck(dataBytes, peer)
	case InvalidSyncMemberOplogMsg:
		err = pm.HandleSyncMemberOplogInvalid(dataBytes, peer)

	case ForceSyncMemberOplogMsg:
		err = pm.HandleForceSyncMemberOplog(dataBytes, peer)
	case ForceSyncMemberOplogAckMsg:
		err = pm.HandleForceSyncMemberOplogAck(dataBytes, peer)

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
		err = pm.HandleSyncOpKeyOplog(dataBytes, peer, SyncOpKeyOplogMsg)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncOpKeyOplog(Msg)", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case SyncOpKeyOplogAckMsg:
		err = pm.HandleSyncOpKeyOplog(dataBytes, peer, SyncOpKeyOplogAckMsg)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncOpKeyOplog(Ack)", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case SyncPendingOpKeyOplogMsg:
		err = pm.HandleSyncPendingOpKeyOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandlePendingSyncOpKeyOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case SyncPendingOpKeyOplogAckMsg:
		err = pm.HandleSyncPendingOpKeyOplogAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandlePendingSyncOpKeyOplogAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	case AddOpKeyOplogMsg:
		err = pm.HandleAddOpKeyOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddOpKeyOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case AddOpKeyOplogsMsg:
		err = pm.HandleAddOpKeyOplogs(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddOpKeyOplogs", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case AddPendingOpKeyOplogMsg:
		err = pm.HandleAddPendingOpKeyOplog(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddPendingOpKeyOplog", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case AddPendingOpKeyOplogsMsg:
		err = pm.HandleAddPendingOpKeyOplogs(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleAddPendingOpKeyOplogs", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	// sync create-op-key
	case SyncCreateOpKeyMsg:
		err = pm.HandleSyncCreateOpKey(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncCreateOpKey", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	case SyncCreateOpKeyAckMsg:
		err = pm.HandleSyncCreateOpKeyAck(dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleSyncCreateOpKeyAck", "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}

	// default
	default:
		err = pm.HandleMessage(op, dataBytes, peer)
		if err != nil {
			log.Error("PMHandleMessageWrapper: unable to HandleMessage", "op", op, "NMsg", NMsg, "e", err, "entity", pm.Entity().IDString(), "peer", peer)
		}
	}

	return err
}
