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

package pttdb

import (
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func GetIterByID(db *LDBDatabase, prefix []byte, idxPrefix []byte, startID *types.PttID, listOrder ListOrder) (iterator.Iterator, error) {
	if startID == nil {
		return db.NewIteratorWithPrefix(nil, prefix, listOrder)
	}

	idxKey, err := common.Concat([][]byte{idxPrefix, startID[:]})
	if err != nil {
		return nil, err
	}

	key, err := db.Get(idxKey)
	if err != nil {
		return nil, err
	}

	return db.NewIteratorWithPrefix(key, prefix, listOrder)
}

func GetFuncIter(iter iterator.Iterator, listOrder ListOrder) func() bool {
	switch listOrder {
	case ListOrderNext:
		return iter.Next
	case ListOrderPrev:
		return iter.Prev
	}

	return nil
}
