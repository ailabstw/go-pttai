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

package service

/**********
 * Sync Media
 **********/

func (pm *BaseProtocolManager) SyncMedia(op OpType, syncIDs []*SyncID, peer *PttPeer) error {
	return pm.SyncObject(op, syncIDs, peer)
}

func (pm *BaseProtocolManager) HandleSyncCreateMedia(dataBytes []byte, peer *PttPeer, syncAckMsg OpType) error {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.HandleSyncCreateObject(dataBytes, peer, obj, syncAckMsg)
}

/**********
 * Sync Media Block
 **********/

func (pm *BaseProtocolManager) HandleSyncMediaBlock(
	dataBytes []byte,
	peer *PttPeer,
	ackMsg OpType,
) error {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.HandleSyncBlock(dataBytes, peer, obj, ackMsg)
}

func (pm *BaseProtocolManager) HandleSyncCreateMediaBlockAck(
	dataBytes []byte,
	peer *PttPeer,

	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
	broadcastLog func(oplog *BaseOplog) error,

) error {

	obj := NewEmptyMedia()
	pm.SetMediaDB(obj)

	return pm.HandleSyncCreateBlockAck(
		dataBytes,
		peer,
		obj,

		merkle,

		setLogDB,
		nil,
		broadcastLog,
	)
}
