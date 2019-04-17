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

package me

import (
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type InternalSyncBoardAck struct {
	LogID *types.PttID `json:"l"`

	BoardData *pkgservice.ApproveJoinEntity `json:"B"`
}

func (pm *ProtocolManager) InternalSyncBoard(
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	syncID := &pkgservice.SyncID{ID: oplog.ObjID, LogID: oplog.ID}
	log.Debug("InternalSyncBoard: to SendDataToPeer", "syncID", syncID, "peer", peer)

	return pm.SendDataToPeer(InternalSyncBoardMsg, syncID, peer)
}

func (pm *ProtocolManager) HandleInternalSyncBoard(
	dataBytes []byte,
	peer *pkgservice.PttPeer,
) error {

	syncID := &pkgservice.SyncID{}
	err := json.Unmarshal(dataBytes, syncID)
	if err != nil {
		return err
	}

	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM()
	board := contentSPM.Entity(syncID.ID)
	if board == nil {
		return types.ErrInvalidID
	}
	boardPM := board.PM()

	myID := pm.Ptt().GetMyEntity().GetID()
	joinEntity := &pkgservice.JoinEntity{ID: myID}
	_, theApproveJoinEntity, err := boardPM.ApproveJoin(joinEntity, nil, peer)
	log.Debug("HandleInternalSyncBoard: after ApproveJoin", "e", err)
	if err != nil {
		return err
	}

	approveJoinEntity, ok := theApproveJoinEntity.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	ackData := &InternalSyncBoardAck{LogID: syncID.LogID, BoardData: approveJoinEntity}
	log.Debug("HandleInternalSyncBoard: to SendData", "ackData", ackData)

	pm.SendDataToPeer(InternalSyncBoardAckMsg, ackData, peer)

	return nil
}

func (pm *ProtocolManager) HandleInternalSyncBoardAck(
	dataBytes []byte,
	peer *pkgservice.PttPeer,

) error {

	// unmarshal data
	log.Debug("HandleInternalSyncBoardAck: start")
	theBoardData := content.NewEmptyApproveJoinBoard()

	data := &InternalSyncBoardAck{BoardData: theBoardData}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	// oplog
	oplog := &pkgservice.BaseOplog{ID: data.LogID}
	pm.SetMeDB(oplog)

	// lock
	err = oplog.Lock()
	if err != nil {
		return err
	}
	defer oplog.Unlock()

	// get
	err = oplog.Get(data.LogID, true)
	log.Debug("HandleInternalSyncBoardAck: after oplog.Get", "e", err, "isSync", oplog.IsSync)
	if oplog.IsSync {
		return nil
	}

	// lock entity
	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM().(*content.ServiceProtocolManager)

	err = contentSPM.Lock(oplog.ObjID)
	if err != nil {
		return err
	}
	defer contentSPM.Unlock(oplog.ObjID)

	theBoard := contentSPM.Entity(oplog.ObjID)
	if theBoard == nil {
		err = pm.handleInternalSyncBoardAckNew(contentSPM, theBoardData, oplog, peer)
		if err != nil {
			return err
		}

		oplog.IsSync = true
		oplog.Save(true, pm.meOplogMerkle)

		return nil
	}
	board, ok := theBoard.(*content.Board)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// exists
	boardStatus := board.Status

	switch {
	case boardStatus == types.StatusAlive && reflect.DeepEqual(board.LogID, oplog.ID):
		err = pm.handleInternalSyncEntityAckSameLog(board, oplog, peer)
	case boardStatus >= types.StatusTerminal:
	case boardStatus == types.StatusAlive:
		err = pm.handleInternalSyncEntityAckDiffAliveLog(board, oplog, peer)
	default:
		err = pm.handleInternalSyncBoardAckDiffLog(contentSPM, theBoardData, oplog, peer)
	}
	if err != nil {
		return err
	}

	oplog.IsSync = true
	oplog.Save(true, pm.meOplogMerkle)
	return nil
}

func (pm *ProtocolManager) handleInternalSyncBoardAckNew(
	spm *content.ServiceProtocolManager,
	data *pkgservice.ApproveJoinEntity,
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	_, err := spm.CreateJoinEntity(data, peer, oplog, true, true, true, true, false)
	log.Debug("HandleInternalSyncBoardAckNew: after CreateJoinEntity", "e", err)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) handleInternalSyncEntityAckSameLog(entity pkgservice.Entity, oplog *pkgservice.BaseOplog, peer *pkgservice.PttPeer) error {
	log.Debug("HandleInternalSyncEntityAckSameLog: to check updateTS", "oplog.UpdateTS", oplog.UpdateTS, "entity.UpdateTS", entity.GetUpdateTS())
	if oplog.UpdateTS.IsLess(entity.GetUpdateTS()) {
		pkgservice.SetEntityWithOplog(entity, types.StatusAlive, oplog)
		entity.Save(true)
	}

	return nil
}

func (pm *ProtocolManager) handleInternalSyncEntityAckDiffAliveLog(entity pkgservice.Entity, oplog *pkgservice.BaseOplog, peer *pkgservice.PttPeer) error {

	entityUpdateTS := entity.GetUpdateTS()
	log.Debug("HandleInternalSyncEntityAckDiffAliveLog: to check updateTS", "oplog.UpdateTS", oplog.UpdateTS, "entity.UpdateTS", entityUpdateTS)
	if entityUpdateTS.IsLess(oplog.UpdateTS) {
		pkgservice.SetEntityWithOplog(entity, types.StatusAlive, oplog)
		entity.Save(true)
	}

	return nil
}

func (pm *ProtocolManager) handleInternalSyncBoardAckDiffLog(
	spm *content.ServiceProtocolManager,
	data *pkgservice.ApproveJoinEntity,
	oplog *pkgservice.BaseOplog,
	peer *pkgservice.PttPeer,
) error {

	_, err := spm.CreateJoinEntity(data, peer, oplog, false, false, true, true, false)
	log.Debug("HandleInternalSyncBoardAckDiffLog: after CreateJoinEntity", "e", err)
	if err != nil {
		return err
	}

	return nil
}
