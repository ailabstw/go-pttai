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
)

func (spm *ServiceProtocolManager) GetUserNameByID(id *types.PttID) (*UserName, error) {
	log.Debug("GetUserNameByID: start", "id", id)
	if id == nil {
		return nil, types.ErrInvalidID
	}
	u := NewEmptyUserName()
	spm.SetUserNameDB(u)
	u.SetID(id)
	err := u.Get(true)
	if err != nil {
		return nil, err
	}

	return u, nil
}
