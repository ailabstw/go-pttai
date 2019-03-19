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
	"bytes"
	"encoding/json"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Friend struct {
	*pkgservice.BaseEntity `json:"e"`

	UpdateTS types.Timestamp `json:"UT"`

	Friend0ID *types.PttID `json:"f0"`
	Friend1ID *types.PttID `json:"f1"`
	FriendID  *types.PttID `json:"f"`

	BoardID *types.PttID   `json:"BID,omitempty"`
	Board   *content.Board `json:"-"`

	ProfileID *types.PttID     `json:"PID,omitempty"`
	Profile   *account.Profile `json:"-"`

	// get from other dbs
	LastSeen        types.Timestamp `json:"-"`
	MessageCreateTS types.Timestamp `json:"-"`
}

func NewEmptyFriend() *Friend {
	return &Friend{BaseEntity: &pkgservice.BaseEntity{SyncInfo: &pkgservice.BaseSyncInfo{}}}
}

func NewFriend(friendID *types.PttID, ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager, dbLock *types.LockMap) (*Friend, error) {

	myID := ptt.GetMyEntity().GetID()
	id, err := pkgservice.NewPttIDWithMixedIDs([]*types.PttID{myID, friendID})
	if err != nil {
		return nil, err
	}

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	e := pkgservice.NewBaseEntity(id, ts, myID, types.StatusInit, dbFriend, dbLock)

	var friend0ID *types.PttID
	var friend1ID *types.PttID
	if bytes.Compare(myID[:], friendID[:]) <= 0 {
		friend0ID = myID
		friend1ID = friendID
	} else {
		friend0ID = friendID
		friend1ID = myID
	}

	f := &Friend{
		BaseEntity: e,
		UpdateTS:   ts,

		Friend0ID: friend0ID,
		Friend1ID: friend1ID,
		FriendID:  friendID,
	}

	err = f.Init(ptt, service, spm)
	if err != nil {
		return nil, err
	}

	log.Debug("NewFriend: done", "createTS", f.CreateTS)

	return f, nil
}

func (f *Friend) GetUpdateTS() types.Timestamp {
	return f.UpdateTS
}

func (f *Friend) SetUpdateTS(ts types.Timestamp) {
	f.UpdateTS = ts
}

func (f *Friend) Init(ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager) error {

	f.SetDB(dbFriend, spm.GetDBLock())

	// friend-id

	err := f.InitPM(ptt, service)
	if err != nil {
		return err
	}

	// profile
	accountSPM := service.(*Backend).accountBackend.SPM()
	if f.ProfileID != nil {
		profile := accountSPM.Entity(f.ProfileID)
		if profile == nil {
			return pkgservice.ErrInvalidEntity
		}
		f.Profile = profile.(*account.Profile)
	}

	// board
	contentSPM := service.(*Backend).contentBackend.SPM()
	if f.BoardID != nil {
		board := contentSPM.Entity(f.BoardID)
		if board == nil {
			return pkgservice.ErrInvalidEntity
		}
		f.Board = board.(*content.Board)
	}

	return nil
}

func (f *Friend) InitPM(ptt pkgservice.Ptt, service pkgservice.Service) error {
	pm, err := NewProtocolManager(f, ptt, service)
	if err != nil {
		return err
	}

	f.BaseEntity.Init(pm, ptt, service)

	return nil
}

func (f *Friend) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := f.JoinTS.Marshal()
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{DBFriendPrefix, marshalTimestamp, f.ID[:]})
}

func (f *Friend) IdxKey() ([]byte, error) {
	return common.Concat([][]byte{DBFriendIdxPrefix, f.ID[:]})
}

func (f *Friend) Idx2Key() ([]byte, error) {
	return common.Concat([][]byte{DBFriendIdx2Prefix, f.FriendID[:]})
}

func (f *Friend) Marshal() ([]byte, error) {
	return json.Marshal(f)
}

func (f *Friend) Unmarshal(theBytes []byte) error {
	err := json.Unmarshal(theBytes, f)
	if err != nil {
		return err
	}

	// postprocess

	return nil
}

func (f *Friend) Save(isLocked bool) error {
	if !isLocked {
		err := f.Lock()
		if err != nil {
			return err
		}
		defer f.Unlock()
	}

	key, err := f.MarshalKey()
	if err != nil {
		return err
	}

	marshaled, err := f.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := f.IdxKey()
	if err != nil {
		return err
	}

	idx2Key, err := f.Idx2Key()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key, idx2Key}, UpdateTS: f.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{
			K: key,
			V: marshaled,
		},
		&pttdb.KeyVal{
			K: idx2Key,
			V: key,
		},
	}

	_, err = dbFriend.ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (f *Friend) GetByFriendID(friendID *types.PttID) error {
	f.FriendID = friendID

	idx2Key, err := f.Idx2Key()
	if err != nil {
		return err
	}

	val, err := dbFriend.GetBy2ndIdxKey(idx2Key)
	log.Debug("GetByFriendID: after GetBy2ndIdxKey", "e", err)
	if err != nil {
		return err
	}

	err = f.Unmarshal(val)
	if err != nil {
		return err
	}

	return nil
}

func (f *Friend) SaveLastSeen(ts types.Timestamp) error {
	f.LastSeen = ts

	key, err := f.MarshalLastSeenKey()
	if err != nil {
		return err
	}
	val := &pttdb.DBable{
		UpdateTS: ts,
	}
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = dbFriendCore.TryPut(key, marshaled, ts)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return err
	}

	return nil
}

func (f *Friend) LoadLastSeen() (types.Timestamp, error) {
	key, err := f.MarshalLastSeenKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	data, err := dbFriendCore.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			err = nil
		}
		return types.ZeroTimestamp, err
	}

	val := &pttdb.DBable{}
	err = json.Unmarshal(data, val)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return val.UpdateTS, nil
}

func (f *Friend) MarshalLastSeenKey() ([]byte, error) {
	return common.Concat([][]byte{DBLastSeenPrefix, f.ID[:]})
}

func (f *Friend) SaveMessageCreateTS(ts types.Timestamp) error {
	f.MessageCreateTS = ts

	key, err := f.MarshalMessageCreateTSKey()
	if err != nil {
		return err
	}
	val := &pttdb.DBable{
		UpdateTS: ts,
	}
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = dbFriendCore.TryPut(key, marshaled, ts)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return err
	}

	err = f.SaveMessageCreateTS2(ts)
	if err != nil {
		return err
	}

	return nil
}

func (f *Friend) LoadMessageCreateTS() (types.Timestamp, error) {
	ts, err := f.LoadMessageCreateTS2()
	if err == nil {
		return ts, nil
	}

	key, err := f.MarshalMessageCreateTSKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	data, err := dbFriendCore.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			err = nil
		}
		return types.ZeroTimestamp, err
	}

	val := &pttdb.DBable{}
	err = json.Unmarshal(data, val)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return val.UpdateTS, nil
}

func (f *Friend) MarshalMessageCreateTSKey() ([]byte, error) {
	return common.Concat([][]byte{DBMessageCreateTSPrefix, f.ID[:]})
}

func (f *Friend) SaveMessageCreateTS2(ts types.Timestamp) error {
	f.MessageCreateTS = ts

	idxKey, err := f.MarshalMessageCreateTSIdxKey()
	if err != nil {
		return err
	}

	key, err := f.MarshalMessageCreateTSKey2(ts)
	if err != nil {
		return err
	}

	val := &pttdb.DBable{
		UpdateTS: ts,
	}

	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: ts}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	_, err = dbFriend.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return err
	}

	return nil
}

func (f *Friend) MarshalMessageCreateTSIdxKey() ([]byte, error) {
	return common.Concat([][]byte{DBMessageCreateTSIdxPrefix, f.ID[:]})
}

func (f *Friend) MarshalMessageCreateTSKey2(ts types.Timestamp) ([]byte, error) {
	marshaledTimestamp, err := ts.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{DBMessageCreateTS2Prefix, marshaledTimestamp, f.ID[:]})
}

func (f *Friend) LoadMessageCreateTS2() (types.Timestamp, error) {
	idxKey, err := f.MarshalMessageCreateTSIdxKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	val, err := dbFriend.GetByIdxKey(idxKey, 0)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	dbable := &pttdb.DBable{}
	err = dbable.Unmarshal(val)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return dbable.UpdateTS, nil
}

func msgCreateTSKeyToEntityID(key []byte) *types.PttID {
	theID := &types.PttID{}

	offset := pttdb.SizeDBKeyPrefix + types.SizeTimestamp
	copy(theID[:], key[offset:])

	return theID
}
