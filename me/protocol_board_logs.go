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

package me

import (
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleBoardLog(
	oplog *pkgservice.BaseOplog,

	info *ProcessMeInfo,
) ([]*pkgservice.BaseOplog, error) {

	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM()

	opData := &MeOpEntity{}

	log.Debug("handleBoardLog: to HandleEntityLog", "op", oplog.Op, "MeOpTypeCreateBoard", MeOpTypeCreateBoard)

	return pm.HandleEntityLog(oplog, contentSPM, opData, info, pm.updateBoardInfo)
}

func (pm *ProtocolManager) updateBoardInfo(oplog *pkgservice.BaseOplog, info *ProcessMeInfo) {
	info.BoardInfo[*oplog.ObjID] = oplog
}

func (pm *ProtocolManager) setNewestBoardLog(
	oplog *pkgservice.BaseOplog,
) (types.Bool, error) {

	opData := &MeOpEntity{}

	err := oplog.GetData(opData)
	if err != nil {
		return true, err
	}

	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM()

	entity := contentSPM.Entity(oplog.ObjID)
	if entity == nil {
		return true, err
	}

	return !types.Bool(reflect.DeepEqual(opData.LogID, entity.GetLogID())), nil

}
