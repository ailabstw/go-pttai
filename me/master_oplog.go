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

// It's possible that we have multiple ids due to multi-device setup.
// Requiring per-entity-level oplog, not unique MeOplog / MasterOplog / PttOplog in ptt-layer

type MasterOplog struct {
	*pkgservice.BaseOplog `json:"O"`
}

func NewMasterOplog(id *types.PttID, ts types.Timestamp, doerID *types.PttID, op pkgservice.OpType, data interface{}, dbLock *types.LockMap) (*MasterOplog, error) {

	oplog, err := pkgservice.NewOplog(id, ts, doerID, op, data, dbMe, id, DBMasterOplogPrefix, DBMasterIdxOplogPrefix, nil, dbLock)
	if err != nil {
		return nil, err
	}

	return &MasterOplog{
		BaseOplog: oplog,
	}, nil
}

func (pm *ProtocolManager) SetMasterDB(oplog *pkgservice.BaseOplog) {
	myID := pm.Entity().GetID()
	oplog.SetDB(dbMe, myID, DBMasterOplogPrefix, DBMasterIdxOplogPrefix, nil, pm.dbMasterLock)
}

func OplogsToMasterOplogs(logs []*pkgservice.BaseOplog) []*MasterOplog {
	typedLogs := make([]*MasterOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &MasterOplog{BaseOplog: log}
	}
	return typedLogs
}

func MasterOplogsToOplogs(typedLogs []*MasterOplog) []*pkgservice.BaseOplog {
	oplogs := make([]*pkgservice.BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		oplogs[i] = log.BaseOplog
	}
	return oplogs
}
