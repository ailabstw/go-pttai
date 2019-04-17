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
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

type JoinBoardEvent struct {
	JoinRequest *pkgservice.JoinRequest
}

func (pm *ProtocolManager) JoinBoard(joinRequest *pkgservice.JoinRequest) error {

	myInfo := pm.Entity().(*MyInfo)
	if myInfo.Status != types.StatusAlive {
		return nil
	}

	// lock
	pm.lockJoinBoardRequest.Lock()
	defer pm.lockJoinBoardRequest.Unlock()

	// hash-val
	hashVal := *joinRequest.Hash

	_, ok := pm.joinBoardRequests[hashVal]
	if ok {
		return types.ErrAlreadyExists
	}

	pm.joinBoardRequests[hashVal] = joinRequest

	pm.EventMux().Post(&JoinBoardEvent{JoinRequest: joinRequest})

	return nil
}

func (pm *ProtocolManager) SyncJoinBoardLoop() error {
	log.Debug("SyncJoinBoardLoop: Start")
	ticker := time.NewTicker(SyncJoinSeconds)
	defer ticker.Stop()

	pm.SyncJoinBoard()

loop:
	for {
		select {
		case <-ticker.C:
			pm.SyncJoinBoard()
		case <-pm.QuitSync():
			log.Info("SyncJoinBoardLoop: QuitSync", "entity", pm.Entity().IDString())
			break loop
		}
	}

	return nil
}

func (pm *ProtocolManager) SyncJoinBoard() error {
	pm.lockJoinBoardRequest.Lock()
	defer pm.lockJoinBoardRequest.Unlock()

	now, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	toRemoveHashs := make([]*common.Address, 0)
	for _, joinRequest := range pm.joinBoardRequests {
		if joinRequest.CreateTS.Ts < now.Ts-pkgservice.IntRenewJoinKeySeconds {
			log.Warn("SyncJoinBoard: expired", "joinRequest", joinRequest.CreateTS, "now", now)
			toRemoveHashs = append(toRemoveHashs, joinRequest.Hash)
			continue
		}

		if joinRequest.Status != pkgservice.JoinStatusPending {
			continue
		}

		pm.processJoinBoardEvent(joinRequest, true)
	}

	for _, hash := range toRemoveHashs {
		delete(pm.joinBoardRequests, *hash)
	}

	return nil
}

/**********
 * BroadcastLoop
 **********/

func (pm *ProtocolManager) JoinBoardLoop() {
	for obj := range pm.joinBoardSub.Chan() {
		ev, ok := obj.Data.(*JoinBoardEvent)
		if !ok {
			continue
		}

		err := pm.processJoinBoardEvent(ev.JoinRequest, false)
		if err != nil {
			log.Error("Unable to process join board event", "data", ev, "e", err)
		}
	}
}

func (pm *ProtocolManager) processJoinBoardEvent(request *pkgservice.JoinRequest, isLocked bool) error {
	if !isLocked {
		pm.lockJoinBoardRequest.Lock()
		defer pm.lockJoinBoardRequest.Unlock()
	}

	if request.Status != pkgservice.JoinStatusPending {
		return pkgservice.ErrInvalidStatus
	}

	hash, key, challenge := request.Hash, request.Key, request.Challenge

	ptt := pm.Ptt()
	err := ptt.TryJoin(challenge, hash, key, request)
	if err != nil {
		return err
	}

	return nil
}
