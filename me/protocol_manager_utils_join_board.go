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

func (pm *ProtocolManager) IsJoinBoardRequests(hash *common.Address) bool {
	pm.lockJoinBoardRequest.RLock()
	defer pm.lockJoinBoardRequest.RUnlock()

	_, ok := pm.joinBoardRequests[*hash]

	return ok
}

func (pm *ProtocolManager) GetBoardRequests() ([]*pkgservice.JoinRequest, error) {
	pm.lockJoinBoardRequest.RLock()
	defer pm.lockJoinBoardRequest.RUnlock()

	theList := make([]*pkgservice.JoinRequest, len(pm.joinBoardRequests))
	i := 0
	for _, request := range pm.joinBoardRequests {
		theList[i] = request
		i++
	}
	return theList, nil
}

func (pm *ProtocolManager) RemoveBoardRequests(hash []byte) (bool, error) {
	pm.lockJoinBoardRequest.Lock()
	defer pm.lockJoinBoardRequest.Unlock()

	addr := &common.Address{}
	copy(addr[:], hash)
	_, ok := pm.joinBoardRequests[*addr]
	if !ok {
		return false, types.ErrAlreadyDeleted
	}

	delete(pm.joinBoardRequests, *addr)

	return true, nil
}
