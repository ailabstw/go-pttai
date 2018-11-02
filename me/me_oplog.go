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

type MeOplog struct {
	*pkgservice.BaseOplog `json:"O"`
}

func NewMeOplog(objID *types.PttID, ts types.Timestamp, doerID *types.PttID, op pkgservice.OpType, data interface{}, myID *types.PttID, dbLock *types.LockMap) (*MeOplog, error) {

	oplog, err := pkgservice.NewOplog(objID, ts, doerID, op, data, dbOplog, myID, DBMeOplogPrefix, DBMeIdxOplogPrefix, DBMeMerkleOplogPrefix, dbLock)
	if err != nil {
		return nil, err
	}

	return &MeOplog{
		BaseOplog: oplog,
	}, nil
}

func (pm *ProtocolManager) SetMeDB(log *pkgservice.BaseOplog) {
	myID := pm.Entity().GetID()
	log.SetDB(dbOplog, myID, DBMeOplogPrefix, DBMeIdxOplogPrefix, DBMeMerkleOplogPrefix, pm.dbMeLock)
}

func OplogsToMeOplogs(logs []*pkgservice.BaseOplog) []*MeOplog {
	typedLogs := make([]*MeOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &MeOplog{BaseOplog: log}
	}
	return typedLogs
}

func MeOplogsToOplogs(typedLogs []*MeOplog) []*pkgservice.BaseOplog {
	logs := make([]*pkgservice.BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		logs[i] = log.BaseOplog
	}
	return logs
}

func OplogToMeOplog(oplog *pkgservice.BaseOplog) *MeOplog {
	if oplog == nil {
		return nil
	}
	return &MeOplog{BaseOplog: oplog}
}
