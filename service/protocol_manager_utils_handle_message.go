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

import "github.com/ailabstw/go-pttai/common"

func PMHandleMessageWrapper(pm ProtocolManager, hash *common.Address, encData []byte, peer *PttPeer) error {
	opKeyInfo, err := pm.GetOpKeyInfoFromHash(hash, false)

	if err != nil {
		return err
	}

	op, dataBytes, err := pm.Ptt().DecryptData(encData, opKeyInfo)
	//log.Debug("PMHandleMessageWrapper: after DecryptData", "e", err, "op", op)
	if err != nil {
		return err
	}

	switch op {
	case IdentifyPeerMsg:
		return pm.HandleIdentifyPeer(dataBytes, peer)
	case IdentifyPeerAckMsg:
		return pm.HandleIdentifyPeerAck(dataBytes, peer)

	// op-key
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

	case SyncCreateOpKeyMsg:
		return pm.HandleSyncCreateOpKey(dataBytes, peer)

	case SyncCreateOpKeyAckMsg:
		return pm.HandleSyncCreateOpKeyAck(dataBytes, peer)

	}

	fitPeerType := pm.GetPeerType(peer)

	if fitPeerType < PeerTypeMember {
		return ErrInvalidEntity
	}

	return pm.HandleMessage(op, dataBytes, peer)
}
