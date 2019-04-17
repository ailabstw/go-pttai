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
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
)

func (pm *ProtocolManager) IsJoinMeKeyHash(hash *common.Address) bool {
	return pm.BaseProtocolManager.IsJoinKeyHash(hash)
}

func (pm *ProtocolManager) IsJoinMeRequests(hash *common.Address) bool {
	pm.lockJoinMeRequest.RLock()
	defer pm.lockJoinMeRequest.RUnlock()

	_, ok := pm.joinMeRequests[*hash]

	return ok
}

func (pm *ProtocolManager) GetMeRequests() ([]*pkgservice.JoinRequest, error) {
	pm.lockJoinMeRequest.RLock()
	defer pm.lockJoinMeRequest.RUnlock()

	lenRequests := len(pm.joinMeRequests)
	results := make([]*pkgservice.JoinRequest, 0, lenRequests)

	for _, joinRequest := range pm.joinMeRequests {
		results = append(results, joinRequest)
	}

	return results, nil
}

func (pm *ProtocolManager) RemoveMeRequests(hash []byte) (bool, error) {
	pm.lockJoinMeRequest.Lock()
	defer pm.lockJoinMeRequest.Unlock()

	addr := &common.Address{}
	copy(addr[:], hash)
	_, ok := pm.joinMeRequests[*addr]
	if !ok {
		return false, types.ErrAlreadyDeleted
	}

	delete(pm.joinMeRequests, *addr)

	return true, nil
}
