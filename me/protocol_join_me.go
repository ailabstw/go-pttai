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
	"reflect"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type JoinMeEvent struct {
	JoinMeRequest *pkgservice.JoinRequest
}

func (pm *ProtocolManager) JoinMe(joinRequest *pkgservice.JoinRequest, myKeyBytes []byte) error {
	myInfo := pm.Entity().(*MyInfo)
	if myInfo.Status != types.StatusAlive {
		return nil
	}

	log.Debug("JoinMe: start", "joinRequest", joinRequest)

	// lock
	pm.lockJoinMeRequest.Lock()
	defer pm.lockJoinMeRequest.Unlock()

	// already with other nodes
	if len(pm.MyNodes) > 1 {
		return ErrAlreadyMyNode
	}

	// me
	myKey, err := types.UnmarshalTextPttID(myKeyBytes)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(myKey, myInfo.GetValidateKey()) {
		return ErrInvalidMe
	}

	// hash-val
	hashVal := *joinRequest.Hash

	_, ok := pm.joinMeRequests[hashVal]
	if ok {
		log.Error("JoinMe: request already exists", "hash", joinRequest.Hash)
		return types.ErrAlreadyExists
	}

	pm.joinMeRequests[hashVal] = joinRequest

	pm.EventMux().Post(&JoinMeEvent{JoinMeRequest: joinRequest})

	return nil
}

func (pm *ProtocolManager) SyncJoinMeLoop() error {
	log.Debug("SyncJoinMeLoop: Start")
	ticker := time.NewTicker(SyncJoinSeconds)
	defer ticker.Stop()

	pm.SyncJoinMe()

loop:
	for {
		select {
		case <-ticker.C:
			pm.SyncJoinMe()
		case <-pm.QuitSync():
			log.Debug("SyncJoinMeLoop: QuitSync", "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
			break loop
		}
	}

	return nil
}

func (pm *ProtocolManager) SyncJoinMe() error {
	pm.lockJoinMeRequest.Lock()
	defer pm.lockJoinMeRequest.Unlock()

	now, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	toRemoveHashs := make([]*common.Address, 0)
	for hash, joinRequest := range pm.joinMeRequests {
		if joinRequest.CreateTS.Ts < now.Ts-pkgservice.IntRenewJoinKeySeconds {
			toRemoveHashs = append(toRemoveHashs, &hash)
			continue
		}

		if joinRequest.Status != pkgservice.JoinStatusPending {
			continue
		}

		pm.EventMux().Post(&JoinMeEvent{JoinMeRequest: joinRequest})
	}

	for _, hash := range toRemoveHashs {
		delete(pm.joinMeRequests, *hash)
	}

	return nil
}

func (pm *ProtocolManager) JoinMeLoop() error {
	for obj := range pm.joinMeSub.Chan() {
		ev, ok := obj.Data.(*JoinMeEvent)
		if !ok {
			log.Error("JoinMeLoop: unable to get JoinMeEvent", "data", obj.Data)
			continue
		}

		err := pm.processJoinMeEvent(ev.JoinMeRequest)
		if err != nil {
			log.Error("unable to process join me event", "e", err)
		}
	}

	return nil
}

func (pm *ProtocolManager) processJoinMeEvent(request *pkgservice.JoinRequest) error {
	pm.lockJoinMeRequest.Lock()
	defer pm.lockJoinMeRequest.Unlock()

	if request.Status != pkgservice.JoinStatusPending {
		return pkgservice.ErrInvalidStatus
	}

	hash, key, challenge := request.Hash, request.Key, request.Challenge

	log.Debug("processJoinMeEvent: TryJoin")

	ptt := pm.myPtt
	err := ptt.TryJoin(challenge, hash, key, request)
	if err != nil {
		return err
	}

	return nil
}
