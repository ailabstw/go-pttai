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

import "github.com/ailabstw/go-pttai/log"

func (pm *BaseProtocolManager) GetEntityLog() (*BaseOplog, error) {
	// get entity log
	entityLogID := pm.Entity().GetLogID()

	entityLog := &BaseOplog{}
	pm.setLog0DB(entityLog)

	entityLog.ID = entityLogID
	err := entityLog.Get(entityLogID, false)
	if err != nil {
		return nil, err
	}

	return entityLog, nil
}

func (pm *BaseProtocolManager) Delete() error {
	return pm.theDelete()
}

func (pm *BaseProtocolManager) defaultDelete() error {
	return nil
}

/*
PostdeleteEntity deals with postdeleting entity.

Especially used in UnregisterMember and posttransferMember (with nil opData) and RevokeNode
*/
func (pm *BaseProtocolManager) PostdeleteEntity(opData OpData, isForce bool) error {
	return pm.postdelete(opData, isForce)
}

func (pm *BaseProtocolManager) DefaultPostdeleteEntity(opData OpData, isForce bool) error {

	// join-key
	pm.CleanJoinKey()

	// op-key
	pm.CleanOpKey()
	pm.CleanOpKeyOplog()

	// master
	pm.CleanMaster()
	pm.CleanMasterOplog()

	// member
	pm.CleanMember(true)
	log.Debug("DefaultPostdeleteEntity: to CleanMemberOplog", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	pm.CleanMemberOplog(true)

	// clean log0
	log.Debug("DefaultPostdeleteEntity: to CleanLog0", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	pm.CleanLog0(true)

	// peer
	pm.CleanPeers()

	return nil
}

func (pm *BaseProtocolManager) FullCleanLog() {
	log.Debug("FullCleanLog: start", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	pm.CleanMember(false)
	pm.CleanMemberOplog(false)
	pm.CleanLog0(false)

	log.Debug("FullCleanLog: end", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
}
