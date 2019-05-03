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

/*
type UserNodeInfo struct {
	ID         *types.PttID
	UserNodeID *types.PttID `json:"nid"`
	NUserNode  int          `json:"n"`
}

func NewUserNodeInfo(
	id *types.PttID,
	userNodeID *types.PttID,
	nUserNode int,
) (*UserNodeInfo, error) {
	return &UserNodeInfo{
		ID:         id,
		UserNodeID: userNodeID,
		NUserNode:  nUserNode,
	}, nil
}

func (u *UserNodeInfo) Save() error {
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}

	marshaled, err := u.Marshal()
	if err != nil {
		return err
	}

	err = dbAccount.DB().Put(key, marshaled)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserNodeInfo) Get(id *types.PttID) error {
	u.ID = id
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}
	val, err := dbAccount.DB().Get(key)
	if err != nil {
		return err
	}
	return u.Unmarshal(val)
}

func (u *UserNodeInfo) Delete() error {
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}
	err = dbAccount.DB().Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserNodeInfo) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserNodeInfo) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserNodeInfo) MarshalKey() ([]byte, error) {
	return append(DBUserNodeInfoPrefix, u.ID[:]...), nil
}
*/
