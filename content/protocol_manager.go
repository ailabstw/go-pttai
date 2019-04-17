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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProtocolManager struct {
	*pkgservice.BaseProtocolManager

	// db
	dbBoardLock      *types.LockMap
	boardOplogMerkle *pkgservice.Merkle

	// title
	dbTitlePrefix    []byte
	dbTitleIdxPrefix []byte

	// article
	dbArticlePrefix    []byte
	dbArticleIdxPrefix []byte

	// comment
	dbCommentPrefix    []byte
	dbCommentIdxPrefix []byte
}

func newBaseProtocolManager(pm *ProtocolManager, ptt pkgservice.Ptt, entity pkgservice.Entity, svc pkgservice.Service) *pkgservice.BaseProtocolManager {

	b, err := pkgservice.NewBaseProtocolManager(
		ptt,

		RenewOpKeySeconds,
		ExpireOpKeySeconds,
		MaxSyncRandomSeconds,
		MinSyncRandomSeconds,

		MaxMasters,

		pm.boardOplogMerkle, // log0Merkle

		// sign
		nil,
		nil,
		nil,
		nil,

		pm.SetBoardDB,        // setLog0DB
		pm.HandleBoardOplogs, // handleLog0s

		nil, // isMaster
		nil, // isMember

		// peer-type
		nil,
		nil,
		nil,
		nil,
		nil,

		pm.SyncBoardOplog, // postsyncMemberOplog

		pm.DeleteBoard,     // theDelete
		pm.postdeleteBoard, // postdelete

		entity, // entity
		svc,

		dbBoard, //db
	)
	if err != nil {
		return nil
	}

	return b
}

func NewProtocolManager(b *Board, ptt pkgservice.Ptt, svc pkgservice.Service) (*ProtocolManager, error) {
	dbBoardLock, err := types.NewLockMap(pkgservice.SleepTimeLock)
	if err != nil {
		return nil, err
	}

	entityID := b.ID
	entityIDBytes, _ := entityID.MarshalText()
	entityIDStr := string(entityIDBytes)

	boardOplogMerkle, err := pkgservice.NewMerkle(DBBoardOplogPrefix, DBBoardMerkleOplogPrefix, b.ID, dbBoard, "("+entityIDStr+"/"+svc.Name()+":board)")
	if err != nil {
		return nil, err
	}
	pm := &ProtocolManager{
		dbBoardLock:      dbBoardLock,
		boardOplogMerkle: boardOplogMerkle,
	}
	pm.BaseProtocolManager = newBaseProtocolManager(pm, ptt, b, svc)

	// title
	pm.dbTitlePrefix = DBTitlePrefix
	pm.dbTitleIdxPrefix = DBTitleIdxPrefix

	// article
	pm.dbArticlePrefix = append(DBArticlePrefix, entityID[:]...)
	pm.dbArticleIdxPrefix = append(DBArticleIdxPrefix, entityID[:]...)

	// comment
	pm.dbCommentPrefix = append(DBCommentPrefix, entityID[:]...)
	pm.dbCommentIdxPrefix = append(DBCommentIdxPrefix, entityID[:]...)

	return pm, nil
}

func (pm *ProtocolManager) Start() error {
	err := pm.BaseProtocolManager.Start()
	if err == pkgservice.ErrAlreadyStarted {
		log.Warn("Start: already started", "entity", pm.Entity().IDString())
		return nil
	}
	if err != nil {
		log.Error("Start: unable to start BaseProtocolManager", "e", err, "entity", pm.Entity().IDString())
		return err
	}

	// XXX #237
	err = pm.Fix237PrelogInCreateArticle()
	if err != nil {
		log.Error("Start: unable to Fix237PrelogInCreateArticle", "e", err, "entity", pm.Entity().IDString())
		return err
	}

	// sync-wg
	syncWG := pm.SyncWG()

	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pm.CreateJoinKeyLoop()
	}()

	log.Debug("Start: to oplog-merkle-tree-loop", "entity", pm.Entity().IDString())

	// oplog-merkle-tree
	syncWG.Add(1)
	go func() {
		defer syncWG.Done()
		pkgservice.PMOplogMerkleTreeLoop(pm, pm.boardOplogMerkle)
	}()

	return nil
}

func (pm *ProtocolManager) Stop() error {
	return nil
}

func (pm *ProtocolManager) Sync(peer *pkgservice.PttPeer) error {
	log.Debug("Sync: start", "entity", pm.Entity().IDString(), "peer", peer, "status", pm.Entity().GetStatus())
	if peer == nil {
		pm.SyncPendingMasterOplog(peer)
		pm.SyncPendingMemberOplog(peer)
		pm.SyncPendingBoardOplog(peer)
		return nil
	}

	err := pm.SyncOplog(peer, pm.MasterMerkle(), pkgservice.SyncMasterOplogMsg)

	log.Debug("Sync: after SyncOplog", "entity", pm.Entity().IDString(), "peer", peer, "e", err)

	if err != nil {
		return err
	}

	return nil
}
