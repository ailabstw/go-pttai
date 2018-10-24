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
	"github.com/syndtr/goleveldb/leveldb"
)

func (spm *ServiceProtocolManager) SetName(ts types.Timestamp, userID *types.PttID, name []byte, boardID *types.PttID, oplogID *types.PttID, status types.Status) (*UserName, error) {

	u := &UserName{ID: userID}

	err := u.Get(userID, true)
	if err == leveldb.ErrNotFound {
		err = nil
		u, err = NewUserName(userID, ts)
		u.BoardID = boardID
		u.Status = types.StatusInit
	}
	if err != nil {
		return nil, err
	}

	if status == types.StatusAlive {
		u.Name = name
		u.UpdateTS = ts
		u.BoardID = boardID
		u.LogID = oplogID
		u.Status = status
		u.SyncNameInfo = nil
	} else {
		u.IntegrateSyncNameInfo(&SyncNameInfo{
			LogID:    oplogID,
			Name:     name,
			BoardID:  boardID,
			UpdateTS: ts,
			Status:   status,
		})
	}

	err = u.Save(true)
	if err != nil {
		return nil, err
	}

	return u, nil
}
