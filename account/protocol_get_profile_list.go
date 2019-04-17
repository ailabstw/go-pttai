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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func (spm *ServiceProtocolManager) GetProfileList(startingID *types.PttID, limit int) ([]*Profile, error) {
	iter, err := getProfileIter(startingID)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	profileList := make([]*Profile, 0)
	i := 0

	for iter.Next() {
		if limit > 0 && i >= limit {
			break
		}

		v := iter.Value()

		eachProfile := &Profile{}
		err = eachProfile.Unmarshal(v)
		if err != nil {
			continue
		}

		profileList = append(profileList, eachProfile)

		i++
	}

	return profileList, nil
}

func getProfileIter(startingID *types.PttID) (iterator.Iterator, error) {
	if startingID == nil {
		return dbAccount.DB().NewIteratorWithPrefix(nil, DBProfilePrefix, pttdb.ListOrderNext)
	}

	// key
	profile := NewEmptyProfile()
	profile.SetID(startingID)

	key, err := profile.MarshalKey()
	if err != nil {
		return nil, err
	}

	// iter
	iter, err := dbAccount.DB().NewIteratorWithPrefix(key, DBProfilePrefix, pttdb.ListOrderNext)
	if err != nil {
		return nil, err
	}

	return iter, nil
}
