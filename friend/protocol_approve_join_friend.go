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

package friend

import (
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ApproveJoin struct {
	Friend    *Friend             `json:"f"`
	OpKeyInfo *pkgservice.KeyInfo `json:"O"`
}

func (pm *ProtocolManager) ApproveJoinFriend(joinEntity *pkgservice.JoinEntity, keyInfo *pkgservice.KeyInfo, peer *pkgservice.PttPeer) (*pkgservice.KeyInfo, interface{}, error) {

	// friend
	f := pm.Entity().(*Friend)

	// register pending peer
	if peer.UserID == nil {
		peer.UserID = joinEntity.ID
	}
	pm.RegisterPendingPeer(peer)

	// op-key
	opKeyInfo, err := pm.GetNewestOpKey(false)
	if err != nil {
		return nil, nil, err
	}

	approveJoin := &ApproveJoin{
		Friend:    f,
		OpKeyInfo: opKeyInfo,
	}
	return opKeyInfo, approveJoin, nil
}
