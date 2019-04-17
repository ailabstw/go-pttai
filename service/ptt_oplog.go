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

package service

import "github.com/ailabstw/go-pttai/common/types"

type PttOplog struct {
	*BaseOplog `json:"O"`
}

func NewPttOplog(objID *types.PttID, ts types.Timestamp, doerID *types.PttID, op OpType, data OpData, myID *types.PttID) (*PttOplog, error) {

	oplog, err := NewOplog(objID, ts, doerID, op, data, dbOplog, myID, DBPttOplogPrefix, DBPttIdxOplogPrefix, nil, DBPttLockMap)
	if err != nil {
		return nil, err
	}
	oplog.IsSync = false
	oplog.MasterLogID = oplog.ID

	return &PttOplog{
		BaseOplog: oplog,
	}, nil
}

func SetPttDB(myID *types.PttID, oplog *BaseOplog) {
	oplog.SetDB(dbOplog, myID, DBPttOplogPrefix, DBPttIdxOplogPrefix, nil, DBPttLockMap)
}

func OplogsToPttOplogs(logs []*BaseOplog) []*PttOplog {
	typedLogs := make([]*PttOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &PttOplog{BaseOplog: log}
	}
	return typedLogs
}

func PttOplogsToOplogs(typedLogs []*PttOplog) []*BaseOplog {
	logs := make([]*BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		logs[i] = log.BaseOplog
	}
	return logs
}
