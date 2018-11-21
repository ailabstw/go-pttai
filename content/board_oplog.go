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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type BoardOplog struct {
	*pkgservice.BaseOplog `json:"O"`
}

func (o *BoardOplog) GetBaseOplog() *pkgservice.BaseOplog {
	return o.BaseOplog
}

func NewBoardOplog(objID *types.PttID, ts types.Timestamp, doerID *types.PttID, op pkgservice.OpType, opData pkgservice.OpData, userID *types.PttID, dbLock *types.LockMap) (*BoardOplog, error) {

	oplog, err := pkgservice.NewOplog(objID, ts, doerID, op, opData, dbBoard, userID, DBBoardOplogPrefix, DBBoardIdxOplogPrefix, DBBoardMerkleOplogPrefix, dbLock)
	if err != nil {
		return nil, err
	}

	return &BoardOplog{
		BaseOplog: oplog,
	}, nil
}

func (pm *ProtocolManager) NewBoardOplog(objID *types.PttID, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	log.Debug("NewBoardOplog: to NewBoardOplogWithTS", "objID", objID)

	return pm.NewBoardOplogWithTS(objID, ts, op, opData)
}

func (pm *ProtocolManager) NewBoardOplogWithTS(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	log.Debug("NewBoardOplogWithTS: start", "objID", objID)

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	oplog, err := NewBoardOplog(objID, ts, myID, op, opData, entityID, pm.dbBoardLock)
	if err != nil {
		return nil, err
	}
	pm.SetBoardDB(oplog.BaseOplog)
	return oplog, nil
}

func (spm *ServiceProtocolManager) NewBoardOplogWithTS(entityID *types.PttID, ts types.Timestamp, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	myID := spm.Ptt().GetMyEntity().GetID()
	log.Debug("spm.NewBoardOplogWithTS: start", "ts", ts)

	return NewBoardOplog(entityID, ts, myID, op, opData, entityID, spm.GetDBLogLock())
}

func (pm *ProtocolManager) SetBoardDB(oplog *pkgservice.BaseOplog) {
	userID := pm.Entity().GetID()
	oplog.SetDB(dbBoard, userID, DBBoardOplogPrefix, DBBoardIdxOplogPrefix, DBBoardMerkleOplogPrefix, pm.dbBoardLock)
}

func OplogsToBoardOplogs(logs []*pkgservice.BaseOplog) []*BoardOplog {
	typedLogs := make([]*BoardOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &BoardOplog{BaseOplog: log}
	}
	return typedLogs
}

func BoardOplogsToOplogs(typedLogs []*BoardOplog) []*pkgservice.BaseOplog {
	logs := make([]*pkgservice.BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		logs[i] = log.BaseOplog
	}
	return logs
}

func OplogToBoardOplog(oplog *pkgservice.BaseOplog) *BoardOplog {
	if oplog == nil {
		return nil
	}
	return &BoardOplog{BaseOplog: oplog}
}
