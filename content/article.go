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

type Article struct {
	*pkgservice.BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	SyncInfo *pkgservice.BaseSyncInfo `json:"s,omitempty"`

	Title []byte `json:"T,omitempty"`

	NPush *pkgservice.Count `json:"-"` // from other db-records
	NBoo  *pkgservice.Count `json:"-"` // from other db-records

	CommentCreateTS types.Timestamp `json:"-"` // from other db-records
	LastSeen        types.Timestamp `json:"-"` // from other db-records

}

func NewArticle(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	title []byte,

) (*Article, error) {

	id, err := types.NewPttID()
	if err != nil {
		return nil, err
	}

	o := pkgservice.NewObject(id, createTS, creatorID, entityID, logID, status)

	return &Article{
		BaseObject: o,

		UpdateTS: createTS,

		Title: title,
	}, nil
}

func NewEmptyArticle() *Article {
	return &Article{BaseObject: &pkgservice.BaseObject{}}
}

func ArticlesToObjs(typedObjs []*Article) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToArticles(objs []pkgservice.Object) []*Article {
	typedObjs := make([]*Article, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Article)
	}
	return typedObjs
}

func AliveArticles(typedObjs []*Article) []*Article {
	objs := make([]*Article, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetArticleDB(u *Article) {

	u.SetDB(dbBoard, pm.DBObjLock(), pm.Entity().GetID(), pm.dbArticlePrefix, pm.dbArticleIdxPrefix, pm.SetBlockInfoDB, pm.SetMediaDB)
}

func (a *Article) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = a.Lock()
		if err != nil {
			return err
		}
		defer a.Unlock()
	}

	key, err := a.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := a.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := a.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: a.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	_, err = a.DB().ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (a *Article) NewEmptyObj() pkgservice.Object {
	newObj := NewEmptyArticle()
	newObj.CloneDB(a.BaseObject)
	return newObj
}

func (a *Article) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newObj := a.NewEmptyObj()
	newObj.SetID(id)
	err := newObj.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newObj, nil
}

func (a *Article) SetUpdateTS(ts types.Timestamp) {
	a.UpdateTS = ts
}

func (a *Article) GetUpdateTS() types.Timestamp {
	return a.UpdateTS
}

func (a *Article) Get(isLocked bool) error {
	var err error

	if !isLocked {
		err = a.RLock()
		if err != nil {
			return err
		}
		defer a.RUnlock()
	}

	key, err := a.MarshalKey()
	if err != nil {
		return err
	}

	val, err := a.DB().DBGet(key)
	if err != nil {
		return err
	}

	return a.Unmarshal(val)
}

func (a *Article) GetByID(isLocked bool) error {
	var err error

	val, err := a.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return a.Unmarshal(val)
}

func (a *Article) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := a.CreateTS.Marshal()
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{a.FullDBPrefix(), marshalTimestamp, a.ID[:]})
}

func (a *Article) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Article) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, a)
}

func (a *Article) GetSyncInfo() pkgservice.SyncInfo {
	if a.SyncInfo == nil {
		return nil
	}
	return a.SyncInfo
}

func (a *Article) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		a.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*pkgservice.BaseSyncInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	a.SyncInfo = syncInfo

	return nil
}

func (a *Article) SaveLastSeen(ts types.Timestamp) error {
	a.LastSeen = ts

	key, err := a.MarshalLastSeenKey()
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

func (a *Article) LoadLastSeen() (types.Timestamp, error) {
	key, err := a.MarshalLastSeenKey()
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

func (a *Article) MarshalLastSeenKey() ([]byte, error) {
	return common.Concat([][]byte{DBArticleLastSeenPrefix, a.EntityID[:], a.ID[:]})
}

func (a *Article) SaveCommentCreateTS(ts types.Timestamp) error {
	a.CommentCreateTS = ts

	key, err := a.MarshalCommentCreateTSKey()
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

func (a *Article) LoadCommentCreateTS() (types.Timestamp, error) {
	key, err := a.MarshalCommentCreateTSKey()
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

func (a *Article) MarshalCommentCreateTSKey() ([]byte, error) {
	return common.Concat([][]byte{DBArticleCommentCreateTSPrefix, a.EntityID[:], a.ID[:]})
}

func (a *Article) IncreaseComment(commentID *types.PttID, commentType CommentType, ts types.Timestamp) error {
	a.SaveCommentCreateTS(ts)

	entityID := a.EntityID
	var count *pkgservice.Count
	var err error
	switch commentType {
	case CommentTypePush:
		count, err = a.LoadPush()
		if err != nil {
			count, err = pkgservice.NewCount(dbBoard, entityID, a.ID, DBPushPrefix, PCommentCount, true)
			log.Debug("IncreaseComment: CommentTypePush: after NewCount", "e", err)
			if err != nil {
				return err
			}
		}
		count.Add(commentID[:])
		count.Save()
	case CommentTypeBoo:
		count, err = a.LoadBoo()
		if err != nil {
			count, err = pkgservice.NewCount(dbBoard, entityID, a.ID, DBBooPrefix, PCommentCount, true)
			if err != nil {
				return err
			}
		}
		count.Add(commentID[:])
		count.Save()
	}

	return nil
}

func (a *Article) LoadPush() (*pkgservice.Count, error) {
	count, err := pkgservice.NewCount(dbBoard, a.EntityID, a.ID, DBPushPrefix, PCommentCount, false)
	if err != nil {
		return nil, err
	}
	err = count.Load()
	if err != nil {
		return nil, err
	}

	return count, nil
}

func (a *Article) LoadBoo() (*pkgservice.Count, error) {
	count, err := pkgservice.NewCount(dbBoard, a.EntityID, a.ID, DBBooPrefix, PCommentCount, false)
	if err != nil {
		return nil, err
	}
	err = count.Load()
	if err != nil {
		return nil, err
	}

	return count, err
}
