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
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Comment struct {
	*pkgservice.BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	SyncInfo *pkgservice.BaseSyncInfo `json:"s,omitempty"`

	ArticleID        *types.PttID `json:"AID"`
	ArticleCreatorID *types.PttID `json:"aID"`

	CommentType CommentType `json:"t"`
}

func NewComment(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	articleID *types.PttID,
	articleCreatorID *types.PttID,
	commentType CommentType,

) (*Comment, error) {

	o := pkgservice.NewObject(entityID, createTS, creatorID, entityID, logID, status)

	return &Comment{
		BaseObject: o,

		UpdateTS: createTS,

		ArticleID:        articleID,
		ArticleCreatorID: articleCreatorID,
		CommentType:      commentType,
	}, nil
}

func NewEmptyComment() *Comment {
	return &Comment{BaseObject: &pkgservice.BaseObject{}}
}

func CommentsToObjs(typedObjs []*Comment) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToComments(objs []pkgservice.Object) []*Comment {
	typedObjs := make([]*Comment, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Comment)
	}
	return typedObjs
}

func AliveComments(typedObjs []*Comment) []*Comment {
	objs := make([]*Comment, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetCommentDB(u *Comment) {

	u.SetDB(dbBoard, pm.DBObjLock(), pm.Entity().GetID(), pm.dbCommentPrefix, pm.dbCommentIdxPrefix, pm.SetBlockInfoDB, pm.SetMediaDB)
}

func (c *Comment) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = c.Lock()
		if err != nil {
			return err
		}
		defer c.Unlock()
	}

	key, err := c.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := c.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := c.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: c.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	_, err = c.DB().ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (c *Comment) NewEmptyObj() pkgservice.Object {
	newObj := NewEmptyComment()
	newObj.CloneDB(c.BaseObject)
	return newObj
}

func (c *Comment) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newU := c.NewEmptyObj()
	newU.SetID(id)
	err := newU.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newU, nil
}

func (c *Comment) SetUpdateTS(ts types.Timestamp) {
	c.UpdateTS = ts
}

func (c *Comment) GetUpdateTS() types.Timestamp {
	return c.UpdateTS
}

func (c *Comment) Get(isLocked bool) error {
	var err error

	if !isLocked {
		err = c.RLock()
		if err != nil {
			return err
		}
		defer c.RUnlock()
	}

	key, err := c.MarshalKey()
	if err != nil {
		return err
	}

	val, err := c.DB().DBGet(key)
	if err != nil {
		return err
	}

	return c.Unmarshal(val)
}

func (c *Comment) GetByID(isLocked bool) error {
	var err error

	val, err := c.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return c.Unmarshal(val)
}

func (c *Comment) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := c.CreateTS.Marshal()
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{c.FullDBPrefix(), c.ArticleID[:], marshalTimestamp, c.ID[:]})
}

func (c *Comment) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Comment) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, c)
}

func (c *Comment) GetSyncInfo() pkgservice.SyncInfo {
	if c.SyncInfo == nil {
		return nil
	}
	return c.SyncInfo
}

func (c *Comment) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		c.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*pkgservice.BaseSyncInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	c.SyncInfo = syncInfo

	return nil
}
