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

package me

import (
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) DeleteMe() error {

	opData := &MeOpDeleteMe{}

	return pm.DeleteEntity(
		MeOpTypeDeleteMe, opData,
		types.StatusInternalDeleted, types.StatusPendingDeleted, types.StatusDeleted,
		pm.NewMeOplog, pm.setPendingDeleteMeSyncInfo, pm.broadcastMeOplogCore, pm.postdeleteDeleteMe)
}

func (pm *ProtocolManager) postdeleteDeleteMe(theOpData pkgservice.OpData, isForce bool) error {

	myInfo := pm.Entity().(*MyInfo)
	myID := myInfo.ID
	myService := pm.Entity().Service()

	raftLead := pm.GetRaftLead(true)
	myRaftID := pm.myPtt.MyRaftID()

	log.Debug("postdeleteMe: start", "myProfileID", myInfo.ProfileID, "myProfile", myInfo.Profile, "isForce", isForce, "raftLead", raftLead, "myRaftID", myRaftID)

	if raftLead != myRaftID {
		return nil
	}

	// delete profile
	myProfile := myInfo.Profile
	if myProfile != nil {
		myProfile.PM().(*account.ProtocolManager).Delete()
	}

	// delete board
	myBoard := myInfo.Board
	if myBoard != nil {
		myBoard.PM().(*content.ProtocolManager).Delete()
	}

	entities := pm.myPtt.GetEntities()
	for _, entity := range entities {
		if entity == myInfo {
			continue
		}
		if entity.Service() == myService {
			continue
		}
		if entity == myProfile {
			continue
		}
		if entity.GetStatus() > types.StatusAlive {
			continue
		}
		if !entity.IsOwner(myID) {
			continue
		}
		entity.PM().Leave()
	}

	return nil
}

func (pm *ProtocolManager) postdeleteMeCore(logID *types.PttID) (err error) {
	return
}
