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
	pm.CleanMember()
	pm.CleanMemberOplog()

	// peer
	pm.CleanPeers()

	return nil
}
