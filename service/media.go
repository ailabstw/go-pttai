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

package service

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type Media struct {
	*BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	SyncInfo *BaseSyncInfo `json:"s,omitempty"`

	MediaType MediaType `json:"T"`
	MediaData MediaData `json:"D,omitempty"`

	Buf []byte `json:"-"`
}

func NewEmptyMedia() *Media {
	return &Media{BaseObject: &BaseObject{}}
}

func NewMedia(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

) (*Media, error) {

	id, err := types.NewPttID()
	if err != nil {
		return nil, err
	}

	o := NewObject(id, createTS, creatorID, entityID, logID, status)

	return &Media{
		BaseObject: o,
		UpdateTS:   createTS,
	}, nil
}

func MediasToObjs(typedObjs []*Media) []Object {
	objs := make([]Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToMedias(objs []Object) []*Media {
	typedObjs := make([]*Media, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Media)
	}
	return typedObjs
}

func (m *Media) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = m.Lock()
		if err != nil {
			return err
		}
		defer m.Unlock()
	}

	key, err := m.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := m.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := m.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: m.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	_, err = m.db.ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

/*
func (m *Media) GetAndDelete(
	isLocked bool,

) error {
	var err error

	if !isLocked {
		err = m.Lock()
		if err != nil {
			return err
		}
		defer m.Unlock()
	}

	err = m.GetByID(true)
	if err != nil {
		return err
	}

	err = m.Delete(true)
	if err != nil {
		return err
	}

	return nil
}
*/

func (m *Media) DeleteAll(
	isLocked bool,
) error {

	var err error
	if !isLocked {
		err = m.Lock()
		if err != nil {
			return err
		}
		defer m.Unlock()
	}

	block := NewEmptyBlock()
	m.SetBlockDB(block)

	err = block.RemoveAll()
	if err != nil {
		return err
	}

	return m.Delete(true)
}

func (m *Media) SetBlockDB(block *Block) {
	fullDBPrefix := append(DBBlockInfoPrefix, m.EntityID[:]...)
	block.SetDB(m.DB(), fullDBPrefix, m.ID, nil)
}

func (m *Media) NewEmptyObj() Object {
	newObj := NewEmptyMedia()
	newObj.CloneDB(m.BaseObject)
	return newObj
}

func (pm *BaseProtocolManager) SetMediaDB(media *Media) {
	media.SetDB(pm.DB(), pm.DBObjLock(), pm.Entity().GetID(), pm.dbMediaPrefix, pm.dbMediaIdxPrefix, pm.SetBlockInfoDB, pm.SetMediaDB)
}

func (m *Media) GetNewObjByID(id *types.PttID, isLocked bool) (Object, error) {
	newObj := m.NewEmptyObj()
	newObj.SetID(id)
	err := newObj.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newObj, nil
}

func (m *Media) SetUpdateTS(ts types.Timestamp) {
	m.UpdateTS = ts
}

func (m *Media) GetUpdateTS() types.Timestamp {
	return m.UpdateTS
}

func (m *Media) GetByID(isLocked bool) error {
	var err error

	val, err := m.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return m.Unmarshal(val)
}

func (m *Media) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := m.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{m.fullDBPrefix, marshalTimestamp, m.ID[:]})
}

func (m *Media) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Media) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	return nil
}

/**********
 * Sync Info
 **********/

func (m *Media) GetSyncInfo() SyncInfo {
	if m.SyncInfo == nil {
		return nil
	}
	return m.SyncInfo
}

func (m *Media) SetSyncInfo(theSyncInfo SyncInfo) error {
	if theSyncInfo == nil {
		m.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*BaseSyncInfo)
	if !ok {
		return ErrInvalidSyncInfo
	}

	m.SyncInfo = syncInfo

	return nil
}

// get buf
func (m *Media) GetBuf() error {

	blockInfo := m.GetBlockInfo()
	if blockInfo == nil {
		return ErrInvalidBlock
	}
	if !blockInfo.IsAllGood {
		return ErrInvalidBlock
	}
	setBlockInfoDB := m.SetBlockInfoDB()
	setBlockInfoDB(blockInfo, m.ID)

	contentBlocks, err := GetContentBlockList(blockInfo, 0, false)
	if err != nil {
		return err
	}

	// blocks
	blocks := make([][]byte, blockInfo.NBlock*NScrambleInBlock)
	for _, eachContentBlocks := range contentBlocks {
		blocks = append(blocks, eachContentBlocks.Buf...)
	}

	buf, err := common.Concat(blocks)
	if err != nil {
		return err
	}

	m.Buf = buf

	return nil
}
