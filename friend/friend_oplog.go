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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type FriendOplog struct {
	*pkgservice.BaseOplog `json:"O"`
}

func (o *FriendOplog) GetBaseOplog() *pkgservice.BaseOplog {
	return o.BaseOplog
}

func NewFriendOplog(objID *types.PttID, ts types.Timestamp, doerID *types.PttID, op pkgservice.OpType, opData pkgservice.OpData, userID *types.PttID, dbLock *types.LockMap) (*FriendOplog, error) {

	oplog, err := pkgservice.NewOplog(objID, ts, doerID, op, opData, dbFriend, userID, DBFriendOplogPrefix, DBFriendIdxOplogPrefix, DBFriendMerkleOplogPrefix, dbLock)
	if err != nil {
		return nil, err
	}

	return &FriendOplog{
		BaseOplog: oplog,
	}, nil
}

func (pm *ProtocolManager) NewFriendOplog(objID *types.PttID, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	log.Debug("NewFriendOplog: to NewFriendOplogWithTS", "objID", objID)

	return pm.NewFriendOplogWithTS(objID, ts, op, opData)
}

func (pm *ProtocolManager) NewFriendOplogWithTS(objID *types.PttID, ts types.Timestamp, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	log.Debug("NewFriendOplogWithTS: start", "objID", objID)

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	oplog, err := NewFriendOplog(objID, ts, myID, op, opData, entityID, pm.dbFriendLock)
	if err != nil {
		return nil, err
	}
	pm.SetFriendDB(oplog.BaseOplog)
	return oplog, nil
}

func (spm *ServiceProtocolManager) NewFriendOplogWithTS(entityID *types.PttID, ts types.Timestamp, op pkgservice.OpType, opData pkgservice.OpData) (pkgservice.Oplog, error) {

	myID := spm.Ptt().GetMyEntity().GetID()
	log.Debug("spm.NewFriendOplogWithTS: start", "ts", ts)

	return NewFriendOplog(entityID, ts, myID, op, opData, entityID, spm.GetDBLogLock())
}

func (pm *ProtocolManager) SetFriendDB(oplog *pkgservice.BaseOplog) {
	userID := pm.Entity().GetID()
	oplog.SetDB(dbFriend, userID, DBFriendOplogPrefix, DBFriendIdxOplogPrefix, DBFriendMerkleOplogPrefix, pm.dbFriendLock)
}

func OplogsToFriendOplogs(logs []*pkgservice.BaseOplog) []*FriendOplog {
	typedLogs := make([]*FriendOplog, len(logs))
	for i, log := range logs {
		typedLogs[i] = &FriendOplog{BaseOplog: log}
	}
	return typedLogs
}

func FriendOplogsToOplogs(typedLogs []*FriendOplog) []*pkgservice.BaseOplog {
	logs := make([]*pkgservice.BaseOplog, len(typedLogs))
	for i, log := range typedLogs {
		logs[i] = log.BaseOplog
	}
	return logs
}

func OplogToFriendOplog(oplog *pkgservice.BaseOplog) *FriendOplog {
	if oplog == nil {
		return nil
	}
	return &FriendOplog{BaseOplog: oplog}
}
