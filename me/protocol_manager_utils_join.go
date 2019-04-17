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
	"sync"

	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
)

func (pm *ProtocolManager) GetJoinRequest(hash *common.Address) (*pkgservice.JoinRequest, error) {

	// friend
	joinRequest, err := pm.getJoinRequestCore(hash, &pm.lockJoinFriendRequest, pm.joinFriendRequests)
	if err == nil {
		return joinRequest, nil
	}

	// me
	joinRequest, err = pm.getJoinRequestCore(hash, &pm.lockJoinMeRequest, pm.joinMeRequests)
	if err == nil {
		return joinRequest, nil
	}

	// content
	joinRequest, err = pm.getJoinRequestCore(hash, &pm.lockJoinBoardRequest, pm.joinBoardRequests)
	if err == nil {
		return joinRequest, nil
	}

	return nil, pkgservice.ErrInvalidMsg

}

func (pm *ProtocolManager) GetJoinType(hash *common.Address) (pkgservice.JoinType, error) {
	if pm.IsJoinMeKeyHash(hash) {
		return pkgservice.JoinTypeMe, nil
	}

	if pm.IsJoinFriendKeyHash(hash) {
		return pkgservice.JoinTypeFriend, nil
	}

	return pkgservice.JoinTypeInvalid, pkgservice.ErrInvalidData
}

func (pm *ProtocolManager) getJoinRequestCore(hash *common.Address, lock *sync.RWMutex, requests map[common.Address]*pkgservice.JoinRequest) (*pkgservice.JoinRequest, error) {

	lock.RLock()
	defer lock.RUnlock()

	joinRequest, ok := requests[*hash]
	if ok {
		return joinRequest, nil
	}

	return nil, pkgservice.ErrInvalidMsg
}

func (pm *ProtocolManager) Master0Hash() []byte {
	return nil
}
