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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
)

func (pm *ProtocolManager) ApproveJoin(joinEntity *pkgservice.JoinEntity, keyInfo *pkgservice.KeyInfo, peer *pkgservice.PttPeer) (*pkgservice.KeyInfo, interface{}, error) {

	hash := keyInfo.Hash

	if pm.IsJoinFriendKeyHash(hash) {
		log.Debug("ApproveJoin: to approve join friend")
		return pm.ApproveJoinFriend(joinEntity, keyInfo, peer)
	}

	if pm.IsJoinMeKeyHash(hash) {
		return pm.ApproveJoinMe(joinEntity, keyInfo, peer)
	}

	return nil, nil, pkgservice.ErrInvalidData
}

func (m *MyInfo) HandleApproveJoin(dataBytes []byte, hash *common.Address, joinRequest *pkgservice.JoinRequest, peer *pkgservice.PttPeer) error {
	return m.PM().(*ProtocolManager).HandleApproveJoin(dataBytes, hash, joinRequest, peer)
}

func (pm *ProtocolManager) HandleApproveJoin(dataBytes []byte, hash *common.Address, joinRequest *pkgservice.JoinRequest, peer *pkgservice.PttPeer) error {

	var err error
	switch {
	case pm.IsJoinFriendRequests(hash):
		err = pm.HandleApproveJoinFriend(dataBytes, joinRequest, peer)
	case pm.IsJoinMeRequests(hash):
		err = pm.HandleApproveJoinMe(dataBytes, joinRequest, peer)
	case pm.IsJoinBoardRequests(hash):
		log.Debug("HandleApproveJoin: is join-board request", "hash", hash)
		err = pm.HandleApproveJoinBoard(dataBytes, joinRequest, peer)
	}

	return err
}
