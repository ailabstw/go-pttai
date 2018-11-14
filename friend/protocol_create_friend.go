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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (spm *ServiceProtocolManager) CreateFriend(friendID *types.PttID) (*Friend, error) {

	data := &FriendOpCreateFriend{
		FriendID: friendID,
	}
	entity, err := spm.CreateEntity(data, FriendOpTypeCreateFriend, spm.NewFriend, spm.NewFriendOplogWithTS, nil, nil)
	log.Debug("CreateFriend: after CreateEntity", "e", err)
	if err != nil {
		return nil, err
	}

	f, ok := entity.(*Friend)
	if !ok {
		return nil, pkgservice.ErrInvalidEntity
	}
	f.Status = types.StatusToBeSynced
	f.Save(false)

	return f, nil
}

func (spm *ServiceProtocolManager) NewFriend(theData pkgservice.CreateData, ptt pkgservice.Ptt, service pkgservice.Service) (pkgservice.Entity, pkgservice.OpData, error) {

	data, ok := theData.(*FriendOpCreateFriend)
	if !ok {
		return nil, nil, pkgservice.ErrInvalidData
	}

	f, err := NewFriend(data.FriendID, ptt, service, spm, spm.GetDBLock())
	if err != nil {
		return nil, nil, err
	}

	log.Debug("spm.NewFriend: to return", "f", f.UpdateTS)

	return f, data, nil
}
