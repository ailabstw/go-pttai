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
	"bytes"
	"encoding/json"
	"unicode/utf8"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type SyncNameInfo struct {
	LogID    *types.PttID    `json:"pl,omitempty"`
	Name     []byte          `json:"N,omitempty"`
	BoardID  *types.PttID    `json:"bID"`
	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`
}

type UserName struct {
	V        types.Version
	ID       *types.PttID
	CreateTS types.Timestamp `json:"CT"`
	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`

	Name         []byte        `json:"N"`
	SyncNameInfo *SyncNameInfo `json:"sn,omitempty"`

	BoardID *types.PttID `json:"bID"`
	LogID   *types.PttID `json:"l"`

	dbLock *types.LockMap
}

func NewUserName(id *types.PttID, ts types.Timestamp) (*UserName, error) {
	return &UserName{
		V:        types.CurrentVersion,
		ID:       id,
		CreateTS: ts,
		UpdateTS: ts,
	}, nil
}

func (u *UserName) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserName) MarshalKey() ([]byte, error) {
	return append(DBUserNamePrefix, u.ID[:]...), nil
}

func (u *UserName) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserName) Save(isLocked bool) error {
	if !isLocked {
		err := u.Lock()
		if err != nil {
			return err
		}
		defer u.Unlock()
	}

	key, err := u.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := u.Marshal()
	if err != nil {
		return err
	}

	_, err = dbAccount.TryPut(key, marshaled, u.UpdateTS)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserName) Get(id *types.PttID, isLocked bool) error {
	if !isLocked {
		err := u.RLock()
		if err != nil {
			return err
		}
		defer u.RUnlock()
	}

	u.ID = id
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}

	theBytes, err := dbAccount.Get(key)
	if err != nil {
		return err
	}

	if theBytes == nil {
		return types.ErrInvalidID
	}

	err = u.Unmarshal(theBytes)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserName) SetName(name []byte) error {
	if !isValidName(name) {
		return ErrInvalidName
	}

	u.Name = name

	return nil
}

func (u *UserName) Update(name []byte, isLocked bool) error {
	err := u.SetName(name)
	if err != nil {
		return err
	}

	u.UpdateTS, err = types.GetTimestamp()
	if err != nil {
		return err
	}

	err = u.Save(isLocked)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserName) Delete(id *types.PttID, isLocked bool) error {
	if !isLocked {
		err := u.Lock()
		if err != nil {
			return err
		}
		defer u.Unlock()
	}

	u.ID = id
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}

	err = dbAccount.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserName) GetList(id *types.PttID, limit int, listOrder pttdb.ListOrder) ([]*UserName, error) {
	u.ID = id
	key, err := u.MarshalKey()
	if err != nil {
		return nil, err
	}
	userNames := make([]*UserName, 0)
	iter, err := dbAccount.NewIteratorWithPrefix(key, nil, listOrder)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	for i := 0; i < limit && iter.Next(); i++ {
		val := iter.Value()

		userName := &UserName{}
		err := userName.Unmarshal(val)
		if err != nil {
			continue
		}
		userNames = append(userNames, userName)
	}

	return userNames, nil
}

func isValidName(name []byte) bool {
	if utf8.RuneCount(name) > MaxNameLength {
		return false
	}

	return true

}

func (u *UserName) IntegrateSyncNameInfo(info *SyncNameInfo) (*types.PttID, error) {
	var origLogID *types.PttID

	switch {
	case u.SyncNameInfo == nil:
		u.SyncNameInfo = info
		return nil, nil
	case info.Status != types.StatusInternalSync && u.SyncNameInfo.Status > info.Status:
		return nil, nil
	case u.SyncNameInfo.Status < info.Status:
		origLogID = u.SyncNameInfo.LogID
		u.SyncNameInfo = info
		return origLogID, nil
	case info.UpdateTS.IsLess(u.SyncNameInfo.UpdateTS):
		return nil, nil
	case u.SyncNameInfo.UpdateTS.IsLess(info.UpdateTS):
		origLogID = u.SyncNameInfo.LogID
		u.SyncNameInfo = info
		return origLogID, nil
	}

	cmp := bytes.Compare(u.SyncNameInfo.LogID[:], info.LogID[:])
	if cmp < 0 {
		return nil, nil
	}

	origLogID = u.SyncNameInfo.LogID
	u.SyncNameInfo = info
	return origLogID, nil
}

func (u *UserName) SetDBLock(dbLock *types.LockMap) {
	u.dbLock = dbLock
}

func (u *UserName) Lock() error {
	return u.dbLock.Lock(u.ID)
}

func (u *UserName) Unlock() error {
	return u.dbLock.Unlock(u.ID)
}

func (u *UserName) RLock() error {
	return u.dbLock.RLock(u.ID)
}

func (u *UserName) RUnlock() error {
	return u.dbLock.RUnlock(u.ID)
}
