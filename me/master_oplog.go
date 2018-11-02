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

	log, err := pkgservice.NewOplog(id, ts, doerID, op, data, dbOplog, id, DBMasterOplogPrefix, DBMasterIdxOplogPrefix, DBMasterMerkleOplogPrefix, dbLock)
	if err != nil {
		return nil, err
	}
	return &MasterOplog{
		BaseOplog: log,
	}, nil
}

func (pm *ProtocolManager) SetMasterDB(log *pkgservice.BaseOplog) {
	myID := pm.Entity().GetID()
	log.SetDB(dbOplog, myID, DBMasterOplogPrefix, DBMasterIdxOplogPrefix, DBMasterMerkleOplogPrefix, pm.dbMasterLock)
}

func OplogsToMasterOplogs(logs []*pkgservice.BaseOplog) []*MasterOplog {
	typedLogs := make([]*MasterOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &MasterOplog{BaseOplog: log}
	}
	return typedLogs
}

func MasterOplogsToOplogs(typedLogs []*MasterOplog) []*pkgservice.BaseOplog {
	logs := make([]*pkgservice.BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		logs[i] = log.BaseOplog
	}
	return logs
}
