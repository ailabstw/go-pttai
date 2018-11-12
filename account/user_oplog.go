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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type UserOplog struct {
	*pkgservice.BaseOplog `json:"O"`
}

func (o *UserOplog) GetBaseOplog() *pkgservice.BaseOplog {
	return o.BaseOplog
}

func NewUserOplog(objID *types.PttID, ts types.Timestamp, doerID *types.PttID, op pkgservice.OpType, opData pkgservice.OpData, userID *types.PttID, dbLock *types.LockMap) (*UserOplog, error) {

	oplog, err := pkgservice.NewOplog(objID, ts, doerID, op, opData, dbAccount, userID, DBUserOplogPrefix, DBUserIdxOplogPrefix, DBUserMerkleOplogPrefix, dbLock)
	if err != nil {
		return nil, err
	}

	return &UserOplog{
		BaseOplog: oplog,
	}, nil
}

func (pm *ProtocolManager) NewUserOplog(objID *types.PttID, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	log.Debug("NewUserOplog: to NewUserOplogWithTS", "objID", objID)

	return pm.NewUserOplogWithTS(objID, ts, op, opData)
}

func (pm *ProtocolManager) NewUserOplogWithTS(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	log.Debug("NewUserOplogWithTS: start", "objID", objID)

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	oplog, err := NewUserOplog(objID, ts, myID, op, opData, entityID, pm.dbUserLock)
	if err != nil {
		return nil, err
	}
	pm.SetUserDB(oplog.BaseOplog)
	return oplog, nil
}

func (spm *ServiceProtocolManager) NewUserOplogWithTS(entityID *types.PttID, ts types.Timestamp, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	myID := spm.Ptt().GetMyEntity().GetID()

	return NewUserOplog(entityID, ts, myID, op, opData, entityID, spm.GetDBLogLock())
}

func (pm *ProtocolManager) SetUserDB(oplog *pkgservice.BaseOplog) {
	userID := pm.Entity().GetID()
	oplog.SetDB(dbAccount, userID, DBUserOplogPrefix, DBUserIdxOplogPrefix, DBUserMerkleOplogPrefix, pm.dbUserLock)
}

func OplogsToUserOplogs(logs []*pkgservice.BaseOplog) []*UserOplog {
	typedLogs := make([]*UserOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &UserOplog{BaseOplog: log}
	}
	return typedLogs
}

func UserOplogsToOplogs(typedLogs []*UserOplog) []*pkgservice.BaseOplog {
	logs := make([]*pkgservice.BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		logs[i] = log.BaseOplog
	}
	return logs
}

func OplogToUserOplog(oplog *pkgservice.BaseOplog) *UserOplog {
	if oplog == nil {
		return nil
	}
	return &UserOplog{BaseOplog: oplog}
}
