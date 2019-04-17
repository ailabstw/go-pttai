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

package content

import pkgservice "github.com/ailabstw/go-pttai/service"

/**********
 * AddBoardOplog
 **********/

func (pm *ProtocolManager) HandleAddBoardOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddOplog(dataBytes, pm.HandleBoardOplogs, peer)
}

func (pm *ProtocolManager) HandleAddBoardOplogs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddOplogs(dataBytes, pm.HandleBoardOplogs, peer)
}

func (pm *ProtocolManager) HandleAddPendingBoardOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddPendingOplog(dataBytes, pm.HandlePendingBoardOplogs, peer)
}

func (pm *ProtocolManager) HandleAddPendingBoardOplogs(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleAddPendingOplogs(dataBytes, pm.HandlePendingBoardOplogs, peer)
}

/**********
 * SyncBoardOplog
 **********/

func (pm *ProtocolManager) HandleSyncBoardOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplog(
		dataBytes,
		peer,

		pm.boardOplogMerkle,

		ForceSyncBoardOplogByMerkleMsg,
		ForceSyncBoardOplogByMerkleAckMsg,
		InvalidSyncBoardOplogMsg,
		SyncBoardOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleForceSyncBoardOplogByMerkle(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByMerkle(
		dataBytes,
		peer,

		ForceSyncBoardOplogByMerkleAckMsg,
		ForceSyncBoardOplogByOplogAckMsg,

		pm.SetBoardDB,
		pm.SetNewestBoardOplog,

		pm.boardOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncBoardOplogByMerkleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByMerkleAck(
		dataBytes,
		peer,

		ForceSyncBoardOplogByMerkleMsg,

		pm.boardOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncBoardOplogByOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplogByOplogAck(
		dataBytes,
		peer,

		pm.HandleBoardOplogs,

		pm.boardOplogMerkle,
	)
}

func (pm *ProtocolManager) HandleForceSyncBoardOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleForceSyncOplog(
		dataBytes,
		peer,

		pm.boardOplogMerkle,
		ForceSyncBoardOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleForceSyncBoardOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	info := NewProcessBoardInfo()

	return pm.HandleForceSyncOplogAck(
		dataBytes,
		peer,

		pm.boardOplogMerkle,
		info,

		pm.SetBoardDB,
		pm.HandleFailedValidBoardOplog,
		pm.SetNewestBoardOplog,
		pm.postprocessFailedValidBoardOplogs,

		SyncBoardOplogNewOplogsMsg,
	)
}

func (pm *ProtocolManager) HandleSyncBoardOplogInvalid(dataBytes []byte, peer *pkgservice.PttPeer) error {

	return pm.HandleSyncOplogInvalid(
		dataBytes,
		peer,

		pm.boardOplogMerkle,
		ForceSyncBoardOplogMsg,
	)
}

func (pm *ProtocolManager) HandleSyncBoardOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogAck(
		dataBytes,
		peer,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.SetNewestBoardOplog,
		pm.postsyncBoardOplogs,

		SyncBoardOplogNewOplogsMsg,
	)
}

func (pm *ProtocolManager) HandleSyncNewBoardOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogNewOplogs(
		dataBytes,
		peer,

		pm.SetBoardDB,
		pm.HandleBoardOplogs,
		pm.SetNewestBoardOplog,

		SyncBoardOplogNewOplogsAckMsg,
	)
}

func (pm *ProtocolManager) HandleSyncNewBoardOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncOplogNewOplogsAck(
		dataBytes,
		peer,

		pm.SetBoardDB,
		pm.HandleBoardOplogs,
		pm.postsyncBoardOplogs,
	)
}

/**********
 * SyncPendingBoardOplog
 **********/

func (pm *ProtocolManager) HandleSyncPendingBoardOplog(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncPendingOplog(
		dataBytes,
		peer,

		pm.HandlePendingBoardOplogs,
		pm.SetBoardDB,
		pm.HandleFailedBoardOplog,

		SyncPendingBoardOplogAckMsg,
	)
}

func (pm *ProtocolManager) HandleSyncPendingBoardOplogAck(dataBytes []byte, peer *pkgservice.PttPeer) error {
	return pm.HandleSyncPendingOplogAck(
		dataBytes,
		peer,

		pm.HandlePendingBoardOplogs,
	)
}

/**********
 * HandleOplogs
 **********/

func (pm *ProtocolManager) HandleBoardOplogs(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer, isUpdateSyncTime bool) error {

	info := NewProcessBoardInfo()

	return pkgservice.HandleOplogs(
		oplogs,
		peer,

		isUpdateSyncTime,
		pm,
		info,
		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.processBoardLog,
		pm.postprocessBoardOplogs,
	)
}

func (pm *ProtocolManager) HandlePendingBoardOplogs(oplogs []*pkgservice.BaseOplog, peer *pkgservice.PttPeer) error {

	info := NewProcessBoardInfo()

	return pkgservice.HandlePendingOplogs(
		oplogs,
		peer,

		pm,
		info,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.processPendingBoardLog,
		pm.processBoardLog,
		pm.postprocessBoardOplogs,
	)
}
