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
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncTitleInfo struct {
	*pkgservice.BaseSyncInfo `json:"b"`

	Title []byte `json:"T,omitempty"`
}

func NewEmptySyncTitleInfo() *SyncTitleInfo {
	return &SyncTitleInfo{BaseSyncInfo: &pkgservice.BaseSyncInfo{}}
}

func (s *SyncTitleInfo) ToObject(theObj pkgservice.Object) error {
	obj, ok := theObj.(*Title)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	s.BaseSyncInfo.ToObject(obj)

	obj.Title = s.Title

	return nil
}

type Title struct {
	*pkgservice.BaseObject `json:"b"`
	UpdateTS               types.Timestamp `json:"UT"`

	SyncInfo *SyncTitleInfo `json:"s,omitempty"`

	Title []byte `json:"T,omitempty"`
}

func NewTitle(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	title []byte,

) (*Title, error) {

	o := pkgservice.NewObject(entityID, createTS, creatorID, entityID, logID, status)

	return &Title{
		BaseObject: o,

		UpdateTS: createTS,

		Title: title,
	}, nil
}

func NewEmptyTitle() *Title {
	return &Title{BaseObject: &pkgservice.BaseObject{}}
}

func TitlesToObjs(typedObjs []*Title) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToTitles(objs []pkgservice.Object) []*Title {
	typedObjs := make([]*Title, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Title)
	}
	return typedObjs
}

func AliveTitles(typedObjs []*Title) []*Title {
	objs := make([]*Title, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetTitleDB(u *Title) {

	u.SetDB(dbBoard, pm.DBObjLock(), pm.Entity().GetID(), pm.dbTitlePrefix, pm.dbTitleIdxPrefix, nil, nil)
}

func (t *Title) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = t.Lock()
		if err != nil {
			return err
		}
		defer t.Unlock()
	}

	key, err := t.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := t.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := t.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: t.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	_, err = t.DB().ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (t *Title) NewEmptyObj() pkgservice.Object {
	newU := NewEmptyTitle()
	newU.CloneDB(t.BaseObject)
	return newU
}

func (t *Title) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newU := t.NewEmptyObj()
	newU.SetID(id)
	err := newU.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newU, nil
}

func (t *Title) SetUpdateTS(ts types.Timestamp) {
	t.UpdateTS = ts
}

func (t *Title) GetUpdateTS() types.Timestamp {
	return t.UpdateTS
}

func (t *Title) Get(isLocked bool) error {
	var err error

	if !isLocked {
		err = t.RLock()
		if err != nil {
			return err
		}
		defer t.RUnlock()
	}

	key, err := t.MarshalKey()
	if err != nil {
		return err
	}

	val, err := t.DB().DBGet(key)
	if err != nil {
		return err
	}

	return t.Unmarshal(val)
}

func (t *Title) GetByID(isLocked bool) error {
	var err error

	val, err := t.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return t.Unmarshal(val)
}

func (t *Title) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{t.FullDBPrefix(), t.ID[:]})
}

func (t *Title) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Title) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, t)
}

func (t *Title) GetSyncInfo() pkgservice.SyncInfo {
	if t.SyncInfo == nil {
		return nil
	}
	return t.SyncInfo
}

func (t *Title) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		t.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*SyncTitleInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	t.SyncInfo = syncInfo

	return nil
}
