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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleFriendLog(
	oplog *pkgservice.BaseOplog,

	info *ProcessMeInfo,
) ([]*pkgservice.BaseOplog, error) {

	friendSPM := pm.Entity().Service().(*Backend).friendBackend.SPM()

	opData := &MeOpEntity{}

	return pm.HandleEntityLog(oplog, friendSPM, opData, info, pm.updateFriendInfo)

}

func (pm *ProtocolManager) updateFriendInfo(oplog *pkgservice.BaseOplog, info *ProcessMeInfo) {
	info.FriendInfo[*oplog.ObjID] = oplog
}

func (pm *ProtocolManager) setNewestFriendLog(
	oplog *pkgservice.BaseOplog,
) (types.Bool, error) {

	opData := &MeOpEntity{}

	err := oplog.GetData(opData)
	if err != nil {
		return true, err
	}

	friendSPM := pm.Entity().Service().(*Backend).friendBackend.SPM()

	entity := friendSPM.Entity(oplog.ObjID)
	if entity == nil {
		return true, err
	}

	return !types.Bool(reflect.DeepEqual(opData.LogID, entity.GetLogID())), nil
}
