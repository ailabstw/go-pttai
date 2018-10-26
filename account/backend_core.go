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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (b *Backend) GetRawUserName(idBytes []byte) (*UserName, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	u := &UserName{}
	err = u.Get(id, true)
	log.Debug("GetRawUserName", "id", id, "u", u, "e", err)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (b *Backend) GetUserName(idBytes []byte) (*BackendUserName, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	u := &UserName{}
	err = u.Get(id, true)
	if err != nil {
		return nil, err
	}

	return userNameToBackendUserName(u), nil
}

func (b *Backend) GetUserNameList(idTextBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendUserName, error) {
	id, err := types.UnmarshalTextPttID(idTextBytes)
	if err != nil {
		return nil, err
	}

	userName := &UserName{}
	userNameList, err := userName.GetList(id, limit, listOrder)
	if err != nil {
		return nil, err
	}

	backendUserNameList := make([]*BackendUserName, len(userNameList))
	for i, eachUserName := range userNameList {
		backendUserNameList[i] = userNameToBackendUserName(eachUserName)
	}

	return backendUserNameList, nil
}

func (b *Backend) GetUserNameByIDs(idByteList [][]byte) (map[string]*BackendUserName, error) {
	backendUserNames := make(map[string]*BackendUserName)
	for _, idBytes := range idByteList {
		id, err := types.UnmarshalTextPttID(idBytes)
		if err != nil {
			continue
		}

		u := &UserName{}
		err = u.Get(id, true)
		if err != nil {
			continue
		}

		backendUserNames[string(idBytes)] = userNameToBackendUserName(u)
	}

	return backendUserNames, nil
}

func (b *Backend) GetRawUserImg(idBytes []byte) (*UserImg, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	u := &UserImg{}
	err = u.Get(id, true)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (b *Backend) GetUserImg(idBytes []byte) (*BackendUserImg, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	u := &UserImg{}
	err = u.Get(id, true)
	if err != nil {
		return nil, err
	}

	return userImgToBackendUserImg(u), nil
}

func (b *Backend) GetUserImgByIDs(idByteList [][]byte) (map[string]*BackendUserImg, error) {
	backendUserImgs := make(map[string]*BackendUserImg)
	for _, idBytes := range idByteList {
		id, err := types.UnmarshalTextPttID(idBytes)
		if err != nil {
			continue
		}

		u := &UserImg{}
		err = u.Get(id, true)
		if err != nil {
			continue
		}

		backendUserImgs[string(idBytes)] = userImgToBackendUserImg(u)
	}

	return backendUserImgs, nil
}
