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

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type SyncImgInfo struct {
	LogID   *types.PttID `json:"pl,omitempty"`
	ImgType ImgType      `json:"T"`
	Width   uint16       `json:"W"`
	Height  uint16       `json:"H"`
	Str     string       `json:"I"`
	BoardID *types.PttID `json:"bID"`

	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`
}

type UserImg struct {
	V        types.Version
	ID       *types.PttID
	CreateTS types.Timestamp `json:"CT"`
	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`

	ImgType     ImgType      `json:"T"`
	Width       uint16       `json:"W"`
	Height      uint16       `json:"H"`
	Str         string       `json:"I"`
	SyncImgInfo *SyncImgInfo `json:"si,omitempty"`

	BoardID *types.PttID `json:"bID"`
	LogID   *types.PttID `json:"l"`

	dbLock *types.LockMap
}

func NewUserImg(id *types.PttID, ts types.Timestamp) (*UserImg, error) {
	return &UserImg{
		V:        types.CurrentVersion,
		ID:       id,
		CreateTS: ts,
		UpdateTS: ts,
	}, nil
}

func (u *UserImg) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserImg) MarshalKey() ([]byte, error) {
	return append(DBUserImgPrefix, u.ID[:]...), nil
}

func (u *UserImg) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserImg) Save(isLocked bool) error {
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

func (u *UserImg) Get(id *types.PttID, isLocked bool) error {
	u.ID = id
	key, err := u.MarshalKey()
	if err != nil {
		return err
	}

	theBytes, err := dbAccount.Get(key)
	//log.Debug("Get: after dbAccount", "theBytes", theBytes, "e", err)
	if err != nil {
		return err
	}

	if theBytes == nil {
		return types.ErrInvalidID
	}

	if len(theBytes) == 0 {
		return types.ErrInvalidID
	}

	log.Debug("to Unmarshal", "u", u, "e", err)
	err = u.Unmarshal(theBytes)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserImg) Delete(id *types.PttID, isLocked bool) error {
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

func (u *UserImg) GetList(id *types.PttID, limit int) ([]*UserImg, error) {
	u.ID = id
	key, err := u.MarshalKey()
	if err != nil {
		return nil, err
	}

	userImgs := make([]*UserImg, 0)
	iter, err := dbAccount.NewIteratorWithPrefix(key, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	for i := 0; i < limit && iter.Next(); i++ {
		val := iter.Value()

		userImg := &UserImg{}
		err := userImg.Unmarshal(val)
		if err != nil {
			continue
		}
		userImgs = append(userImgs, userImg)
	}

	return userImgs, nil
}

func (u *UserImg) IntegrateSyncImgInfo(info *SyncImgInfo) (*types.PttID, error) {
	var origLogID *types.PttID

	switch {
	case u.SyncImgInfo == nil:
		u.SyncImgInfo = info
		return nil, nil
	case info.Status != types.StatusInternalSync && u.SyncImgInfo.Status > info.Status:
		return nil, nil
	case u.SyncImgInfo.Status < info.Status:
		origLogID = u.SyncImgInfo.LogID
		u.SyncImgInfo = info
		return origLogID, nil
	case info.UpdateTS.IsLess(u.SyncImgInfo.UpdateTS):
		return nil, nil
	case u.SyncImgInfo.UpdateTS.IsLess(info.UpdateTS):
		origLogID = u.SyncImgInfo.LogID
		u.SyncImgInfo = info
		return origLogID, nil
	}

	cmp := bytes.Compare(u.SyncImgInfo.LogID[:], info.LogID[:])
	if cmp < 0 {
		return nil, nil
	}

	origLogID = u.SyncImgInfo.LogID
	u.SyncImgInfo = info
	return origLogID, nil
}

func (u *UserImg) SetDBLock(dbLock *types.LockMap) {
	u.dbLock = dbLock
}

func (u *UserImg) Lock() error {
	return u.dbLock.Lock(u.ID)
}

func (u *UserImg) Unlock() error {
	return u.dbLock.Unlock(u.ID)
}

func (u *UserImg) RLock() error {
	return u.dbLock.RLock(u.ID)
}

func (u *UserImg) RUnlock() error {
	return u.dbLock.RUnlock(u.ID)
}
