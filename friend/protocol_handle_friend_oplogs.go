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

import pkgservice "github.com/ailabstw/go-pttai/service"

/**********
 * AddFriendOplog
 **********/

func (pm *ProtocolManager) HandleAddFriendOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddOplog(dataBytes, pm.HandleFriendOplogs, peer)
}

func (pm *ProtocolManager) HandleAddFriendOplogs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddOplogs(dataBytes, pm.HandleFriendOplogs, peer)
}

func (pm *ProtocolManager) HandleAddPendingFriendOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddPendingOplog(dataBytes, pm.HandlePendingFriendOplogs, peer)
}

func (pm *ProtocolManager) HandleAddPendingFriendOplogs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddPendingOplogs(dataBytes, pm.HandlePendingFriendOplogs, peer)
}

/**********
 * SyncFriendOplog
 **********/

func (pm *ProtocolManager) HandleSyncFriendOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplog(
		dataBytes,
		peer,

		pm.friendOplogMerkle,

		ForceSyncFriendOplogByMerkleMsg,
		ForceSyncFriendOplogByMerkleAckMsg,
		InvalidSyncFriendOplogMsg,
		SyncFriendOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleForceSyncFriendOplogByMerkle(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByMerkle(
		dataBytes,
		peer,

		ForceSyncFriendOplogByMerkleAckMsg,
		ForceSyncFriendOplogByOplogAckMsg,

		pm.SetFriendDB,
		pm.SetNewestFriendOplog,

		pm.friendOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncFriendOplogByMerkleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByMerkleAck(
		dataBytes,
		peer,

		ForceSyncFriendOplogByMerkleMsg,

		pm.friendOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncFriendOplogByOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByOplogAck(
		dataBytes,
		peer,

		pm.HandleFriendOplogs,

		pm.friendOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncFriendOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplog(
		dataBytes,
		peer,

		pm.friendOplogMerkle,
		ForceSyncFriendOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleForceSyncFriendOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	info := NewProcessFriendInfo()

	return pm.HandleForceSyncOplogAck(
		dataBytes,
		peer,

		pm.friendOplogMerkle,
		info,

		pm.SetFriendDB,
		pm.HandleFailedValidFriendOplog,
		pm.SetNewestFriendOplog,
		pm.postprocessFailedValidFriendOplogs,

		SyncFriendOplogNewOplogsMsg,
	)
}

func (pm *ProtocolManager) HandleSyncFriendOplogInvalid(dataBytes []byte, peer *pkgservice.PttPeer) error {

	return pm.HandleSyncOplogInvalid(
		dataBytes,
		peer,

		pm.friendOplogMerkle,
		ForceSyncFriendOplogByMerkleMsg,
	)
}

func (pm *ProtocolManager) HandleSyncFriendOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogAck(
		dataBytes,
		peer,

		pm.friendOplogMerkle,
		pm.SetFriendDB,
		pm.SetNewestFriendOplog,
		pm.postsyncFriendOplogs,

		SyncFriendOplogNewOplogsMsg,
	)
}

func (pm *ProtocolManager) HandleSyncNewFriendOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogNewOplogs(
		dataBytes,
		peer,

		pm.SetFriendDB,
		pm.HandleFriendOplogs,
		pm.SetNewestFriendOplog,

		SyncFriendOplogNewOplogsAckMsg,
	)
}

func (pm *ProtocolManager) HandleSyncNewFriendOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogNewOplogsAck(
		dataBytes,
		peer,

		pm.SetFriendDB,
		pm.HandleFriendOplogs,
		pm.postsyncFriendOplogs,
	)
}

/**********
 * SyncPendingFriendOplog
 **********/

func (pm *ProtocolManager) HandleSyncPendingFriendOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncPendingOplog(
		dataBytes,
		peer,

		pm.HandlePendingFriendOplogs,
		pm.SetFriendDB,
		pm.HandleFailedFriendOplog,

		SyncPendingFriendOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleSyncPendingFriendOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncPendingOplogAck(
		dataBytes,
		peer,

		pm.HandlePendingFriendOplogs,
	)
}

/**********
 * HandleOplogs
 **********/

func (pm *ProtocolManager) HandleFriendOplogs(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isUpdateSyncTime bool) error {

	info := NewProcessFriendInfo()

	return pkgservice.HandleOplogs(
		oplogs,
		peer,

		isUpdateSyncTime,
		pm,
		info,
		pm.friendOplogMerkle,

		pm.SetFriendDB,
		pm.processFriendLog,
		pm.postprocessFriendOplogs,
	)
}

func (pm *ProtocolManager) HandlePendingFriendOplogs(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer) error {

	info := NewProcessFriendInfo()

	return pkgservice.HandlePendingOplogs(
		oplogs,
		peer,

		pm,
		info,

		pm.friendOplogMerkle,

		pm.SetFriendDB,
		pm.processPendingFriendLog,
		pm.processFriendLog,
		pm.postprocessFriendOplogs,
	)
}
