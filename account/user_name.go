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

type SyncUserNameInfo struct {
	*pkgservice.BaseSyncInfo `json:"b"`

	Name []byte `json:"N,omitempty"`
}

func NewEmptySyncUserNameInfo() *SyncUserNameInfo {
	return &SyncUserNameInfo{BaseSyncInfo: &pkgservice.BaseSyncInfo{}}
}

func (s *SyncUserNameInfo) ToObject(theObj pkgservice.Object) error {
	obj, ok := theObj.(*UserName)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	s.BaseSyncInfo.ToObject(obj)

	obj.Name = s.Name

	return nil
}

type UserName struct {
	*pkgservice.BaseObject `json:"b"`
	UpdateTS               types.Timestamp   `json:"UT"`
	SyncInfo               *SyncUserNameInfo `json:"s,omitempty"`

	Name []byte `json:"N,omitempty"`
}

func NewUserName(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	name []byte,

) (*UserName, error) {

	id := creatorID

	o := pkgservice.NewObject(id, createTS, creatorID, entityID, logID, status)

	return &UserName{
		BaseObject: o,
		UpdateTS:   createTS,

		Name: name,
	}, nil
}

func NewEmptyUserName() *UserName {
	return &UserName{BaseObject: &pkgservice.BaseObject{}}
}

func UserNamesToObjs(typedObjs []*UserName) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToUserNames(objs []pkgservice.Object) []*UserName {
	typedObjs := make([]*UserName, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*UserName)
	}
	return typedObjs
}

func AliveUserNames(typedObjs []*UserName) []*UserName {
	objs := make([]*UserName, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetUserNameDB(u *UserName) {
	spm := pm.Entity().Service().SPM()

	u.SetDB(dbAccount, spm.DBObjLock(), pm.Entity().GetID(), pm.dbUserNamePrefix, pm.dbUserNameIdxPrefix, nil, nil)
}

func (spm *ServiceProtocolManager) SetUserNameDB(u *UserName) {
	u.SetDB(dbAccount, spm.DBObjLock(), nil, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)
}

func (u *UserName) Save(isLocked bool) error {
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

func (u *UserName) NewEmptyObj() pkgservice.Object {
	newU := NewEmptyUserName()
	newU.CloneDB(u.BaseObject)
	return newU
}

func (u *UserName) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newU := u.NewEmptyObj()
	newU.SetID(id)
	err := newU.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newU, nil
}

func (u *UserName) SetUpdateTS(ts types.Timestamp) {
	u.UpdateTS = ts
}

func (u *UserName) GetUpdateTS() types.Timestamp {
	return u.UpdateTS
}

func (u *UserName) Get(isLocked bool) error {
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

func (u *UserName) GetByID(isLocked bool) error {
	var err error

	val, err := u.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return u.Unmarshal(val)
}

func (u *UserName) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{DBUserNamePrefix, u.ID[:]})
}

func (u *UserName) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserName) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserName) GetSyncInfo() pkgservice.SyncInfo {
	if u.SyncInfo == nil {
		return nil
	}
	return u.SyncInfo
}

func (u *UserName) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		u.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*SyncUserNameInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	u.SyncInfo = syncInfo

	return nil
}
