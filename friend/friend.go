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

	BoardID *types.PttID `json:"BID,omitempty"`

	ProfileID *types.PttID     `json:"PID,omitempty"`
	Profile   *account.Profile `json:"-"`

	// get from other dbs
	LastSeen        types.Timestamp `json:"-"`
	ArticleCreateTS types.Timestamp `json:"-"`
}

func NewEmptyFriend() *Friend {
	return &Friend{BaseEntity: &pkgservice.BaseEntity{}}
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

	return nil
}

func (f *Friend) InitPM(ptt pkgservice.Ptt, service pkgservice.Service) error {
	pm, err := NewProtocolManager(f, ptt)
	if err != nil {
		return err
	}

	f.BaseEntity.Init(pm, ptt, service)

	return nil
}

func (f *Friend) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{DBFriendPrefix, f.ID[:]})
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

	err = f.Unmarshal(val)
	if err != nil {
		return err
	}

	return nil
}

func (b *Friend) SaveLastSeen(ts types.Timestamp) error {
	b.LastSeen = ts

	key, err := b.MarshalLastSeenKey()
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

func (b *Friend) LoadLastSeen() (types.Timestamp, error) {
	key, err := b.MarshalLastSeenKey()
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

func (b *Friend) MarshalLastSeenKey() ([]byte, error) {
	return common.Concat([][]byte{content.DBBoardLastSeenPrefix, b.ID[:]})
}

func (b *Friend) SaveArticleCreateTS(ts types.Timestamp) error {
	b.ArticleCreateTS = ts

	key, err := b.MarshalArticleCreateTSKey()
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

func (b *Friend) LoadArticleCreateTS() (types.Timestamp, error) {
	key, err := b.MarshalArticleCreateTSKey()
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

func (b *Friend) MarshalArticleCreateTSKey() ([]byte, error) {
	return common.Concat([][]byte{content.DBBoardArticleCreateTSPrefix, b.ID[:]})
}
