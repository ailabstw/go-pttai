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

import (
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *BaseProtocolManager) Fix138MasterLog() error {
	// oplog
	oplogs, err := pm.GetMasterOplogList(nil, 0, pttdb.ListOrderNext, types.StatusAlive)
	if err != nil {
		return err
	}

	lenOplogs := len(oplogs)

	if lenOplogs <= 1 {
		return nil
	}

	minLog := oplogs[0]
	var eachLog *MasterOplog
	for plogs := oplogs; len(plogs) > 0; plogs = plogs[1:] {
		eachLog = plogs[0]
		if eachLog.UpdateTS.IsLess(minLog.UpdateTS) {
			minLog = eachLog
		}
	}

	eachMaster := NewEmptyMaster()
	pm.SetMasterObjDB(eachMaster)

	for plogs := oplogs; len(plogs) > 0; plogs = plogs[1:] {
		eachLog = plogs[0]
		if eachLog == minLog {
			continue
		}

		if eachLog.Op != MasterOpTypeAddMaster {
			continue
		}

		pm.SetMasterDB(eachLog.BaseOplog)
		eachLog.Delete(false)

		eachMaster.SetID(eachLog.ObjID)
		err = eachMaster.GetByID(false)
		if err != nil {
			log.Warn("Fix138MasterLog: unable to get master", "master", eachMaster.ID, "err", err, "log", eachLog.ID, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
			continue
		}

		if !reflect.DeepEqual(eachMaster.LogID, eachLog.ID) {
			log.Warn("Fix138MasterLog: unmatched LogID", "master", eachMaster.ID, "masterLog", eachMaster.LogID, "log", eachLog.ID, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
			continue
		}

		eachMaster.Delete(false)
	}

	return nil
}
