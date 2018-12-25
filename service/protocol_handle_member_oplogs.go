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

/**********
 * AddMemberOplog
 **********/

func (pm *BaseProtocolManager) HandleAddMemberOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddOplog(dataBytes, pm.HandleMemberOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddMemberOplogs(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddOplogs(dataBytes, pm.HandleMemberOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingMemberOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddPendingOplog(dataBytes, pm.HandlePendingMemberOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingMemberOplogs(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddPendingOplogs(dataBytes, pm.HandlePendingMemberOplogs, peer)
}

/**********
 * SyncMemberOplog
 **********/

func (pm *BaseProtocolManager) HandleSyncMemberOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplog(dataBytes, peer, pm.MemberMerkle(), SyncMemberOplogAckMsg)
}

func (pm *BaseProtocolManager) HandleSyncMemberOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplogAck(dataBytes, peer, pm.MemberMerkle(), pm.SetMemberDB, pm.SetNewestMemberOplog, pm.postsyncMemberOplogs, SyncMemberOplogNewOplogsMsg)
}

func (pm *BaseProtocolManager) HandleSyncNewMemberOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplogNewOplogs(dataBytes, peer, pm.SetMemberDB, pm.HandleMemberOplogs, pm.SetNewestMemberOplog, SyncMemberOplogNewOplogsAckMsg)
}

func (pm *BaseProtocolManager) HandleSyncNewMemberOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplogNewOplogsAck(dataBytes, peer, pm.SetMemberDB, pm.HandleMemberOplogs, pm.postsyncMemberOplogs)
}

/**********
 * SyncPendingMemberOplog
 **********/

func (pm *BaseProtocolManager) HandleSyncPendingMemberOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncPendingOplog(dataBytes, peer, pm.HandlePendingMemberOplogs, pm.SetMemberDB, pm.HandleFailedMemberOplog, SyncPendingMemberOplogAckMsg)
}

func (pm *BaseProtocolManager) HandleSyncPendingMemberOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncPendingOplogAck(dataBytes, peer, pm.HandlePendingMemberOplogs)
}

/**********
 * HandleOplogs
 **********/

func (pm *BaseProtocolManager) HandleMemberOplogs(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error {

	info := NewProcessPersonInfo()

	return HandleOplogs(oplogs, peer, isUpdateSyncTime, pm, info, pm.memberMerkle, pm.SetMemberDB, pm.processMemberLog, pm.postprocessMemberOplogs)
}

func (pm *BaseProtocolManager) HandlePendingMemberOplogs(oplogs []*BaseOplog, peer *PttPeer) error {

	info := NewProcessPersonInfo()

	return HandlePendingOplogs(oplogs, peer, pm, info, pm.SetMemberDB, pm.processPendingMemberLog, pm.processMemberLog, pm.postprocessMemberOplogs)

}
