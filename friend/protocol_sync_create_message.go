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

func (pm *ProtocolManager) SyncMessage(op pkgservice.OpType, syncIDs []*pkgservice.SyncID, peer *pkgservice.PttPeer) error {
	return pm.SyncObject(op, syncIDs, peer)
}

func (pm *ProtocolManager) HandleSyncCreateMessage(dataBytes []byte, peer *pkgservice.PttPeer, syncAckMsg pkgservice.OpType) error {

	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	return pm.HandleSyncCreateObject(dataBytes, peer, obj, syncAckMsg)
}

/**********
 * Sync Message Block
 **********/

func (pm *ProtocolManager) SyncMessageBlock(op pkgservice.OpType, syncBlockIDs []*pkgservice.SyncBlockID, peer *pkgservice.PttPeer) error {
	return pm.SyncBlock(SyncCreateMessageBlockMsg, syncBlockIDs, peer)
}

func (pm *ProtocolManager) HandleSyncCreateMessageBlock(dataBytes []byte, peer *pkgservice.PttPeer) error {

	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	log.Debug("HandleSyncCreateMessageBlock: to HandleSyncBlock")

	return pm.HandleSyncCreateBlock(dataBytes, peer, obj, SyncCreateMessageBlockAckMsg)
}

func (pm *ProtocolManager) HandleSyncCreateMessageBlockAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	return pm.HandleSyncCreateBlockAck(dataBytes, peer, obj, pm.SetFriendDB, pm.postcreateMessage, pm.broadcastFriendOplogCore)
}
