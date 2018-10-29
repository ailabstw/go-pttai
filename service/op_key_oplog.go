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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type OpKeyOplog struct {
	*Oplog `json:"O"`
}

func NewOpKeyOplog(keyID *types.PttID, ts types.Timestamp, doerID *types.PttID, op OpType, data interface{}, db *pttdb.LDBBatch, entityID *types.PttID, dbLock *types.LockMap) (*OpKeyOplog, error) {

	log, err := NewOplog(keyID, ts, doerID, op, data, db, entityID, DBOpKeyOplogPrefix, DBOpKeyIdxOplogPrefix, DBOpKeyMerkleOplogPrefix, dbLock)
	if err != nil {
		return nil, err
	}

	return &OpKeyOplog{
		Oplog: log,
	}, nil
}

func (pm *BaseProtocolManager) setOpKeyDB(log *Oplog) {
	entityID := pm.Entity().GetID()
	log.SetDB(pm.DBOpKeyInfo(), entityID, DBOpKeyOplogPrefix, DBOpKeyIdxOplogPrefix, DBOpKeyMerkleOplogPrefix, pm.DBOpKeyLock())
}

func OplogsToOpKeyOplogs(logs []*Oplog) []*OpKeyOplog {
	opKeyLogs := make([]*OpKeyOplog, len(logs))
	for i, log := range logs {
		opKeyLogs[i] = &OpKeyOplog{Oplog: log}
	}
	return opKeyLogs
}

func OpKeyOplogsToOplogs(opKeyLogs []*OpKeyOplog) []*Oplog {
	logs := make([]*Oplog, len(opKeyLogs))
	for i, log := range opKeyLogs {
		logs[i] = log.Oplog
	}
	return logs
}
