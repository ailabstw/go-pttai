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

package account

import (
	"encoding/json"
	"math/rand"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
)

type UserNode struct {
	ID     *types.PttID
	NodeID *discover.NodeID `json:"NID"`

	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`

	LogID *types.PttID `json:"l"`
}

func NewUserNode() (*UserNode, error) {
	return &UserNode{}, nil
}

func (u *UserNode) Save() error {
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := u.Marshal()
	if err != nil {
		return err
	}

	_, err = dbAccountCore.TryPut(key, marshaled, u.UpdateTS)
	if err != nil {
		return err
	}

	if u.Status != types.StatusAlive {
		return nil
	}

	// is-to-save-idx
	count, err := u.Count()
	if err != nil {
		return err
	}

	if count <= 0 {
		count = 1
	}

	randInt := rand.Intn(count)
	if randInt != 0 {
		return nil
	}

	// to save-idx
	idxKey, err := u.IdxKey()
	if err != nil {
		return err
	}

	err = dbAccountCore.Put(idxKey, marshaled)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserNode) Get(userID *types.PttID) error {
	u.ID = userID
	idxKey, err := u.IdxKey()
	if err != nil {
		return err
	}

	val, err := dbAccountCore.Get(idxKey)
	if err != nil {
		return err
	}

	err = u.Unmarshal(val)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserNode) Count() (int, error) {
	prefix, err := u.Prefix()
	if err != nil {
		return 0, err
	}

	iter, err := dbAccountCore.NewIteratorWithPrefix(nil, prefix, pttdb.ListOrderNext)
	if err != nil {
		return 0, err
	}
	defer iter.Release()

	count := 0
	for iter.Next() {
		count++
	}
	return count, nil
}

func (u *UserNode) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserNode) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserNode) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{DBUserNodePrefix, u.ID[:], u.NodeID[:]})
}

func (u *UserNode) IdxKey() ([]byte, error) {
	return common.Concat([][]byte{DBUserNodeIdxPrefix, u.ID[:]})
}

func (u *UserNode) Prefix() ([]byte, error) {
	return common.Concat([][]byte{DBUserNodePrefix, u.ID[:]})
}
