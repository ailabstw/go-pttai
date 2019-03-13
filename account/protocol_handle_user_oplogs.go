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

import pkgservice "github.com/ailabstw/go-pttai/service"

/**********
 * AddUserOplog
 **********/

func (pm *ProtocolManager) HandleAddUserOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddOplog(dataBytes, pm.HandleUserOplogs, peer)
}

func (pm *ProtocolManager) HandleAddUserOplogs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddOplogs(dataBytes, pm.HandleUserOplogs, peer)
}

func (pm *ProtocolManager) HandleAddPendingUserOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddPendingOplog(dataBytes, pm.HandlePendingUserOplogs, peer)
}

func (pm *ProtocolManager) HandleAddPendingUserOplogs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddPendingOplogs(dataBytes, pm.HandlePendingUserOplogs, peer)
}

/**********
 * SyncUserOplog
 **********/

func (pm *ProtocolManager) HandleSyncUserOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplog(
		dataBytes,
		peer,

		pm.userOplogMerkle,

		ForceSyncUserOplogByMerkleMsg,
		ForceSyncUserOplogByMerkleAckMsg,
		InvalidSyncUserOplogMsg,
		SyncUserOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleForceSyncUserOplogByMerkle(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByMerkle(
		dataBytes,
		peer,

		ForceSyncUserOplogByMerkleAckMsg,
		ForceSyncUserOplogByOplogAckMsg,

		pm.SetUserDB,
		pm.SetNewestUserOplog,

		pm.userOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncUserOplogByMerkleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByMerkleAck(
		dataBytes,
		peer,

		ForceSyncUserOplogByMerkleMsg,

		pm.userOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncUserOplogByOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByOplogAck(
		dataBytes,
		peer,

		pm.HandleUserOplogs,

		pm.userOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncUserOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplog(
		dataBytes,
		peer,

		pm.userOplogMerkle,
		ForceSyncUserOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleForceSyncUserOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	info := NewProcessUserInfo()

	return pm.HandleForceSyncOplogAck(
		dataBytes,
		peer,

		pm.userOplogMerkle,
		info,

		pm.SetUserDB,
		pm.HandleFailedValidUserOplog,
		pm.SetNewestUserOplog,
		pm.postprocessFailedValidUserOplogs,

		SyncUserOplogNewOplogsMsg,
	)
}

func (pm *ProtocolManager) HandleSyncUserOplogInvalid(dataBytes []byte, peer *pkgservice.PttPeer) error {

	return pm.HandleSyncOplogInvalid(
		dataBytes,
		peer,

		pm.userOplogMerkle,
		ForceSyncUserOplogByMerkleMsg,
	)
}

func (pm *ProtocolManager) HandleSyncUserOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogAck(
		dataBytes,
		peer,

		pm.userOplogMerkle,

		pm.SetUserDB,
		pm.SetNewestUserOplog,
		pm.postsyncUserOplogs,

		SyncUserOplogNewOplogsMsg,
	)
}

func (pm *ProtocolManager) HandleSyncNewUserOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogNewOplogs(
		dataBytes,
		peer,

		pm.SetUserDB,
		pm.HandleUserOplogs,
		pm.SetNewestUserOplog,

		SyncUserOplogNewOplogsAckMsg,
	)
}

func (pm *ProtocolManager) HandleSyncNewUserOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogNewOplogsAck(
		dataBytes,
		peer,

		pm.SetUserDB,
		pm.HandleUserOplogs,
		pm.postsyncUserOplogs,
	)
}

/**********
 * SyncPendingUserOplog
 **********/

func (pm *ProtocolManager) HandleSyncPendingUserOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncPendingOplog(
		dataBytes,
		peer,

		pm.HandlePendingUserOplogs,
		pm.SetUserDB,
		pm.HandleFailedUserOplog,

		SyncPendingUserOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleSyncPendingUserOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncPendingOplogAck(
		dataBytes,
		peer,

		pm.HandlePendingUserOplogs,
	)
}

/**********
 * HandleOplogs
 **********/

func (pm *ProtocolManager) HandleUserOplogs(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isUpdateSyncTime bool) error {

	info := NewProcessUserInfo()
	merkle := pm.userOplogMerkle

	return pkgservice.HandleOplogs(
		oplogs,
		peer,

		isUpdateSyncTime,
		pm,
		info,
		merkle,

		pm.SetUserDB,
		pm.processUserLog,
		pm.postprocessUserOplogs,
	)
}

func (pm *ProtocolManager) HandlePendingUserOplogs(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer) error {

	info := NewProcessUserInfo()

	return pkgservice.HandlePendingOplogs(
		oplogs,
		peer,

		pm,
		info,

		pm.userOplogMerkle,

		pm.SetUserDB,
		pm.processPendingUserLog,
		pm.processUserLog,
		pm.postprocessUserOplogs,
	)
}
