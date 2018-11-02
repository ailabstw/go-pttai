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
	"sync"

	"github.com/ailabstw/go-pttai/common"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) GetJoinRequest(hash *common.Address) (*pkgservice.JoinRequest, error) {

	joinRequest, err := pm.getJoinRequestCore(hash, &pm.lockJoinMeRequest, pm.joinMeRequests)
	if err == nil {
		return joinRequest, nil
	}

	return nil, pkgservice.ErrInvalidMsg
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

func (pm *ProtocolManager) GetJoinMeRequests() (map[common.Address]*pkgservice.JoinRequest, *sync.RWMutex) {
	return pm.joinMeRequests, &pm.lockJoinMeRequest
}

func (pm *ProtocolManager) Master0Hash() []byte {
	return nil
}
