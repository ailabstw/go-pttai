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
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type JoinFriendEvent struct {
	JoinRequest *pkgservice.JoinRequest
}

func (pm *ProtocolManager) JoinFriend(joinRequest *pkgservice.JoinRequest) error {

	myInfo := pm.Entity().(*MyInfo)
	if myInfo.Status != types.StatusAlive {
		return nil
	}

	// lock
	pm.lockJoinFriendRequest.Lock()
	defer pm.lockJoinFriendRequest.Unlock()

	// hash-val
	hashVal := *joinRequest.Hash

	_, ok := pm.joinFriendRequests[hashVal]
	if ok {
		log.Error("JoinFriend: friend-request already exists", "hash", joinRequest.Hash)
		return types.ErrAlreadyExists
	}

	pm.joinFriendRequests[hashVal] = joinRequest

	pm.EventMux().Post(&JoinFriendEvent{JoinRequest: joinRequest})

	return nil
}

func (pm *ProtocolManager) SyncJoinFriendLoop() error {
	log.Debug("SyncJoinFriendLoop: Start")
	ticker := time.NewTicker(SyncJoinSeconds)
	defer ticker.Stop()

	pm.SyncJoinFriend()

loop:
	for {
		select {
		case <-ticker.C:
			pm.SyncJoinFriend()
		case <-pm.QuitSync():
			log.Debug("SyncJoinFriendLoop: QuitSync", "entity", pm.Entity().GetID())
			break loop
		}
	}

	return nil
}

func (pm *ProtocolManager) SyncJoinFriend() error {
	pm.lockJoinFriendRequest.Lock()
	defer pm.lockJoinFriendRequest.Unlock()

	now, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	toRemoveHashs := make([]*common.Address, 0)
	for hash, joinRequest := range pm.joinFriendRequests {
		if joinRequest.CreateTS.Ts < now.Ts-pkgservice.IntRenewJoinKeySeconds {
			log.Warn("SyncJoinFriend: expired", "joinRequest", joinRequest.CreateTS, "now", now)
			toRemoveHashs = append(toRemoveHashs, &hash)
			continue
		}

		if joinRequest.Status != pkgservice.JoinStatusPending {
			continue
		}

		pm.processJoinFriendEvent(joinRequest, true)
	}

	log.Debug("SyncJoinFriend: to remove hashs", "hashs", toRemoveHashs)
	for _, hash := range toRemoveHashs {
		delete(pm.joinFriendRequests, *hash)
	}

	return nil
}

func (pm *ProtocolManager) JoinFriendLoop() error {
	for obj := range pm.joinFriendSub.Chan() {
		ev, ok := obj.Data.(*JoinFriendEvent)
		if !ok {
			log.Error("JoinFriendLoop: unable to get JoinFriendEvent", "data", obj.Data)
			continue
		}

		err := pm.processJoinFriendEvent(ev.JoinRequest, false)
		if err != nil {
			log.Error("unable to process join friend event", "e", err)
		}
	}

	return nil
}

func (pm *ProtocolManager) processJoinFriendEvent(request *pkgservice.JoinRequest, isLocked bool) error {
	if !isLocked {
		pm.lockJoinFriendRequest.Lock()
		defer pm.lockJoinFriendRequest.Unlock()
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
