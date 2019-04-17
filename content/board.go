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

package content

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

type Board struct {
	*pkgservice.BaseEntity `json:"e"`

	UpdateTS types.Timestamp `json:"UT"`

	Title []byte `json:"T,omitempty"`

	// get from other dbs
	LastSeen        types.Timestamp `json:"-"`
	ArticleCreateTS types.Timestamp `json:"-"`

	BoardMerkle *pkgservice.Merkle `json:"-"`
}

func NewEmptyBoard() *Board {
	return &Board{BaseEntity: &pkgservice.BaseEntity{SyncInfo: &pkgservice.BaseSyncInfo{}}}
}

func NewBoard(myID *types.PttID, ts types.Timestamp, ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager, dbLock *types.LockMap) (*Board, error) {

	id, err := pkgservice.NewPttIDWithMyID(myID)
	if err != nil {
		return nil, err
	}

	e := pkgservice.NewBaseEntity(id, ts, myID, types.StatusInit, dbBoard, dbLock)

	b := &Board{
		BaseEntity: e,
		UpdateTS:   ts,
	}

	err = b.Init(ptt, service, spm)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Board) GetUpdateTS() types.Timestamp {
	return b.UpdateTS
}

func (b *Board) SetUpdateTS(ts types.Timestamp) {
	b.UpdateTS = ts
}

func (b *Board) Init(ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager) error {

	b.SetDB(dbBoard, spm.GetDBLock())

	err := b.InitPM(ptt, service)
	if err != nil {
		return err
	}

	return nil
}

func (b *Board) InitPM(ptt pkgservice.Ptt, service pkgservice.Service) error {
	pm, err := NewProtocolManager(b, ptt, service)
	if err != nil {
		log.Error("InitPM: unable to NewProtocolManager", "e", err)
		return err
	}

	b.BaseEntity.Init(pm, ptt, service)

	return nil
}

func (b *Board) IdxKey() ([]byte, error) {
	return common.Concat([][]byte{DBBoardIdxPrefix, b.ID[:]})
}

func (b *Board) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := b.JoinTS.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{DBBoardPrefix, marshalTimestamp, b.ID[:]})

}

func (b *Board) Marshal() ([]byte, error) {
	return json.Marshal(b)
}

func (b *Board) Unmarshal(theBytes []byte) error {
	err := json.Unmarshal(theBytes, b)
	if err != nil {
		return err
	}

	// postprocess

	return nil
}

func (b *Board) Save(isLocked bool) error {
	if !isLocked {
		err := b.Lock()
		if err != nil {
			return err
		}
		defer b.Unlock()
	}

	key, err := b.MarshalKey()
	if err != nil {
		return err
	}

	marshaled, err := b.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := b.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: b.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{
			K: key,
			V: marshaled,
		},
	}

	_, err = dbBoard.ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (b *Board) SaveLastSeen(ts types.Timestamp) error {
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

	_, err = dbBoardCore.TryPut(key, marshaled, ts)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return err
	}

	return nil
}

func (b *Board) LoadLastSeen() (types.Timestamp, error) {
	key, err := b.MarshalLastSeenKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	data, err := dbBoardCore.Get(key)
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

func (b *Board) MarshalLastSeenKey() ([]byte, error) {
	return common.Concat([][]byte{DBBoardLastSeenPrefix, b.ID[:]})
}

func (b *Board) SaveArticleCreateTS(ts types.Timestamp) error {
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

	_, err = dbBoardCore.TryPut(key, marshaled, ts)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return err
	}

	return nil
}

func (b *Board) LoadArticleCreateTS() (types.Timestamp, error) {
	key, err := b.MarshalArticleCreateTSKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}
	data, err := dbBoardCore.Get(key)
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

func (b *Board) MarshalArticleCreateTSKey() ([]byte, error) {
	return common.Concat([][]byte{DBBoardArticleCreateTSPrefix, b.ID[:]})
}

func (b *Board) MarshalCommentCreateTSKey() ([]byte, error) {
	return common.Concat([][]byte{DBBoardCommentCreateTSPrefix, b.ID[:]})
}
