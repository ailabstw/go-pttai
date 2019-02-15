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
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func (spm *ServiceProtocolManager) GetFriendList(startingFriendID *types.PttID, limit int, listOrder pttdb.ListOrder) ([]*Friend, error) {
	iter, err := getFriendIter(startingFriendID, listOrder)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	iterFunc := pttdb.GetFuncIter(iter, listOrder)

	friendList := make([]*Friend, 0)

	i := 0
	for iterFunc() {
		if limit > 0 && i >= limit {
			break
		}

		k := iter.Key()
		log.Debug("GetFriendList (in-for-loop)", "k", k)
		v := iter.Value()

		eachFriend := NewEmptyFriend()
		err := eachFriend.Unmarshal(v)
		if err != nil {
			continue
		}

		ts, _ := eachFriend.LoadLastSeen()
		eachFriend.LastSeen = ts

		ts, _ = eachFriend.LoadMessageCreateTS()
		eachFriend.MessageCreateTS = ts

		friendList = append(friendList, eachFriend)

		i++
	}

	return friendList, nil
}

func getFriendIter(startingID *types.PttID, listOrder pttdb.ListOrder) (iterator.Iterator, error) {
	if startingID == nil {
		return dbFriend.DB().NewIteratorWithPrefix(nil, DBFriendPrefix, listOrder)
	}

	// key
	f := NewEmptyFriend()
	f.SetID(startingID)

	key, err := f.MarshalKey()
	if err != nil {
		return nil, err
	}

	// iter
	iter, err := dbFriend.DB().NewIteratorWithPrefix(key, DBFriendPrefix, listOrder)
	if err != nil {
		return nil, err
	}

	return iter, nil
}
