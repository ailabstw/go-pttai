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
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) handleRevokeMeLog(oplog *pkgservice.BaseOplog, info *ProcessMeInfo) ([]*pkgservice.BaseOplog, error) {

	opData := &MeOpRevokeMe{}

	toBroadcastLogs, err := pm.HandleDeleteEntityLog(oplog, info, opData, types.StatusDeleted, pm.SetMeDB, pm.postdeleteRevokeMe)
	if err != nil {
		return nil, err
	}

	return toBroadcastLogs, nil
}

func (pm *ProtocolManager) handlePendingRevokeMeLog(oplog *pkgservice.BaseOplog, info *ProcessMeInfo) ([]*pkgservice.BaseOplog, error) {

	opData := &MeOpRevokeMe{}
	return pm.HandlePendingDeleteEntityLog(oplog, info, types.StatusPendingDeleted, MeOpTypeRevokeMe, opData, pm.SetMeDB)
}

func (pm *ProtocolManager) setNewestRevokeMeLog(oplog *pkgservice.BaseOplog) (types.Bool, error) {

	return pm.SetNewestDeleteEntityLog(oplog)
}

func (pm *ProtocolManager) handleFailedRevokeMeLog(oplog *pkgservice.BaseOplog) error {

	return pm.HandleFailedDeleteEntityLog(oplog)

}
