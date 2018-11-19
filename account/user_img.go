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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncUserImgInfo struct {
	*pkgservice.BaseSyncInfo `json:"b"`

	ImgType ImgType `json:"T"`
	Width   uint16  `json:"W"`
	Height  uint16  `json:"H"`
	Str     string  `json:"I"`
}

func NewEmptySyncUserImgInfo() *SyncUserImgInfo {
	return &SyncUserImgInfo{BaseSyncInfo: &pkgservice.BaseSyncInfo{}}
}

func (s *SyncUserImgInfo) ToObject(theObj pkgservice.Object) error {
	obj, ok := theObj.(*UserImg)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	s.BaseSyncInfo.ToObject(obj)

	obj.ImgType = s.ImgType
	obj.Width = s.Width
	obj.Height = s.Height
	obj.Str = s.Str

	return nil
}

type UserImg struct {
	*pkgservice.BaseObject `json:"b"`
	UpdateTS               types.Timestamp `json:"UT"`

	SyncInfo *SyncUserImgInfo `json:"si,omitempty"`

	ImgType ImgType `json:"T"`
	Width   uint16  `json:"W"`
	Height  uint16  `json:"H"`
	Str     string  `json:"I"`
}

func NewUserImg(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,
) (*UserImg, error) {

	id := creatorID

	o := pkgservice.NewObject(id, createTS, creatorID, entityID, logID, status)

	return &UserImg{
		BaseObject: o,
		UpdateTS:   createTS,
	}, nil
}

func NewEmptyUserImg() *UserImg {
	return &UserImg{BaseObject: &pkgservice.BaseObject{}}
}

func UserImgsToObjs(typedObjs []*UserImg) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToUserImgs(objs []pkgservice.Object) []*UserImg {
	typedObjs := make([]*UserImg, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*UserImg)
	}
	return typedObjs
}

func AliveUserImgs(typedObjs []*UserImg) []*UserImg {
	objs := make([]*UserImg, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetUserImgDB(u *UserImg) {
	spm := pm.Entity().Service().SPM()

	u.SetDB(dbAccount, spm.DBObjLock(), pm.Entity().GetID(), pm.dbUserImgPrefix, pm.dbUserImgIdxPrefix, nil, nil)
}

func (spm *ServiceProtocolManager) SetUserImgDB(u *UserImg) {
	u.SetDB(dbAccount, spm.DBObjLock(), nil, DBUserImgPrefix, DBUserImgIdxPrefix, nil, nil)
}

func (u *UserImg) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = u.Lock()
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

	idxKey, err := u.IdxKey()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: u.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
	}

	_, err = u.DB().ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserImg) NewEmptyObj() pkgservice.Object {
	newU := NewEmptyUserImg()
	newU.CloneDB(u.BaseObject)
	return newU
}

func (u *UserImg) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newU := u.NewEmptyObj()
	newU.SetID(id)
	err := newU.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newU, nil
}

func (u *UserImg) SetUpdateTS(ts types.Timestamp) {
	u.UpdateTS = ts
}

func (u *UserImg) GetUpdateTS() types.Timestamp {
	return u.UpdateTS
}

func (u *UserImg) Get(isLocked bool) error {
	var err error

	if !isLocked {
		err = u.RLock()
		if err != nil {
			return err
		}
		defer u.RUnlock()
	}

	key, err := u.MarshalKey()
	if err != nil {
		return err
	}

	val, err := u.DB().DBGet(key)
	if err != nil {
		return err
	}

	return u.Unmarshal(val)
}

func (u *UserImg) GetByID(isLocked bool) error {
	var err error

	val, err := u.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return u.Unmarshal(val)
}

func (u *UserImg) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{DBUserImgPrefix, u.ID[:]})
}

func (u *UserImg) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserImg) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserImg) GetSyncInfo() pkgservice.SyncInfo {
	if u.SyncInfo == nil {
		return nil
	}
	return u.SyncInfo
}

func (u *UserImg) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		u.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*SyncUserImgInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	u.SyncInfo = syncInfo

	return nil
}
