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
 * AddOpKeyOplog
 **********/

func (pm *BaseProtocolManager) HandleAddOpKeyOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddOplog(dataBytes, pm.HandleOpKeyOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddOpKeyOplogs(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddOplogs(dataBytes, pm.HandleOpKeyOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingOpKeyOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddPendingOplog(dataBytes, pm.HandlePendingOpKeyOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingOpKeyOplogs(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddPendingOplogs(dataBytes, pm.HandlePendingOpKeyOplogs, peer)
}

/**********
 * SyncPendingOpKeyOplog
 **********/

func (pm *BaseProtocolManager) HandleSyncPendingOpKeyOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncPendingOplog(
		dataBytes,
		peer,

		pm.HandlePendingOpKeyOplogs,
		pm.SetOpKeyDB,
		pm.HandleFailedOpKeyOplog,
		SyncPendingOpKeyOplogAckMsg,
	)
}

func (pm *BaseProtocolManager) HandleSyncPendingOpKeyOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncPendingOplogAck(
		dataBytes,
		peer,

		pm.HandlePendingOpKeyOplogs,
	)
}

/**********
 * HandleOplogs
 **********/

func (pm *BaseProtocolManager) HandleOpKeyOplogs(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error {

	info := NewProcessOpKeyInfo()

	return HandleOplogs(
		oplogs,
		peer,
		isUpdateSyncTime,

		pm,
		info,
		nil,

		pm.SetOpKeyDB,
		pm.processOpKeyLog,
		pm.postprocessOpKeyOplogs,
	)
}

func (pm *BaseProtocolManager) HandlePendingOpKeyOplogs(oplogs []*BaseOplog, peer *PttPeer) error {

	info := NewProcessOpKeyInfo()

	return HandlePendingOplogs(
		oplogs,
		peer,

		pm,
		info,
		nil,

		pm.SetOpKeyDB,
		pm.processPendingOpKeyLog,
		pm.processOpKeyLog,
		pm.postprocessOpKeyOplogs,
	)
}
