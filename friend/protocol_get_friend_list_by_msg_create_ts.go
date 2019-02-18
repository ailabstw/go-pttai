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
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func (spm *ServiceProtocolManager) GetFriendListByMsgCreateTS(startingTS types.Timestamp, limit int, listOrder pttdb.ListOrder) ([]*Friend, error) {
	iter, err := getFriendByMsgCreateTSIter(startingTS, listOrder)

	if err != nil {
		return nil, err
	}
	defer iter.Release()

	iterFunc := pttdb.GetFuncIter(iter, listOrder)

	friendList := make([]*Friend, 0)

	i := 0
	var k []byte
	var entityID *types.PttID
	var f *Friend
	for iterFunc() {
		if limit > 0 && i >= limit {
			break
		}

		k = iter.Key()
		entityID = msgCreateTSKeyToEntityID(k)

		f = spm.Entity(entityID).(*Friend)
		if f == nil {
			continue
		}

		friendList = append(friendList, f)

		i++

	}

	return friendList, nil
}

func getFriendByMsgCreateTSIter(startingTS types.Timestamp, listOrder pttdb.ListOrder) (iterator.Iterator, error) {
	if startingTS == types.ZeroTimestamp {
		return dbFriend.DB().NewIteratorWithPrefix(nil, DBMessageCreateTS2Prefix, listOrder)
	}

	key, err := getFriendIterMarshalKey(startingTS, listOrder)
	if err != nil {
		return nil, err
	}

	// iter
	iter, err := dbFriend.DB().NewIteratorWithPrefix(key, DBMessageCreateTS2Prefix, listOrder)
	if err != nil {
		return nil, err
	}

	return iter, nil
}

func getFriendIterMarshalKey(ts types.Timestamp, listOrder pttdb.ListOrder) ([]byte, error) {
	f := NewEmptyFriend()
	if listOrder == pttdb.ListOrderPrev {
		copy(f.ID[:], types.MaxID[:])
	}

	key, err := f.MarshalMessageCreateTSKey2(ts)
	if err != nil {
		return nil, err
	}

	return key, nil
}
