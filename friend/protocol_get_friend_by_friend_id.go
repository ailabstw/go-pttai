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

package friend

import "github.com/ailabstw/go-pttai/common/types"

func (spm *ServiceProtocolManager) GetFriendByFriendID(friendID *types.PttID) (*Friend, error) {
	f := NewEmptyFriend()

	err := f.GetByFriendID(friendID)
	if err != nil {
		return nil, err
	}

	ts, _ := f.LoadLastSeen()
	f.LastSeen = ts

	ts, err = f.LoadMessageCreateTS()
	f.MessageCreateTS = ts

	return f, nil
}

func (spm *ServiceProtocolManager) GetFriendEntityByFriendID(friendID *types.PttID) (*Friend, error) {
	f := NewEmptyFriend()

	err := f.GetByFriendID(friendID)
	if err != nil {
		return nil, err
	}

	entity := spm.Entity(f.ID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	return entity.(*Friend), nil
}
