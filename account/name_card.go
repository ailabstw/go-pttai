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

package account

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncNameCardInfo struct {
	*pkgservice.BaseSyncInfo `json:"b"`

	Card []byte `json:"C,omitempty"`
}

func NewEmptySyncNameCardInfo() *SyncNameCardInfo {
	return &SyncNameCardInfo{BaseSyncInfo: &pkgservice.BaseSyncInfo{}}
}

func (s *SyncNameCardInfo) ToObject(theObj pkgservice.Object) error {
	obj, ok := theObj.(*NameCard)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	s.BaseSyncInfo.ToObject(obj)

	obj.Card = s.Card

	return nil
}

type NameCard struct {
	*pkgservice.BaseObject `json:"b"`
	UpdateTS               types.Timestamp   `json:"UT"`
	SyncInfo               *SyncNameCardInfo `json:"s,omitempty"`

	Card []byte `json:"C,omitempty"`
}

func NewNameCard(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	card []byte,

) (*NameCard, error) {

	id := creatorID

	o := pkgservice.NewObject(id, createTS, creatorID, entityID, logID, status)

	return &NameCard{
		BaseObject: o,
		UpdateTS:   createTS,

		Card: card,
	}, nil
}

func NewEmptyNameCard() *NameCard {
	return &NameCard{BaseObject: &pkgservice.BaseObject{}}
}

func NameCardsToObjs(typedObjs []*NameCard) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToNameCards(objs []pkgservice.Object) []*NameCard {
	typedObjs := make([]*NameCard, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*NameCard)
	}
	return typedObjs
}

func AliveNameCards(typedObjs []*NameCard) []*NameCard {
	objs := make([]*NameCard, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetNameCardDB(u *NameCard) {
	spm := pm.Entity().Service().SPM()

	u.SetDB(dbAccount, spm.DBObjLock(), pm.Entity().GetID(), pm.dbNameCardPrefix, pm.dbNameCardIdxPrefix, nil, nil)
}

func (spm *ServiceProtocolManager) SetNameCardDB(u *NameCard) {
	u.SetDB(dbAccount, spm.DBObjLock(), nil, DBNameCardPrefix, DBNameCardIdxPrefix, nil, nil)
}

func (u *NameCard) Save(isLocked bool) error {
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

func (u *NameCard) NewEmptyObj() pkgservice.Object {
	newU := NewEmptyNameCard()
	newU.CloneDB(u.BaseObject)
	return newU
}

func (u *NameCard) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newU := u.NewEmptyObj()
	newU.SetID(id)
	err := newU.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newU, nil
}

func (u *NameCard) SetUpdateTS(ts types.Timestamp) {
	u.UpdateTS = ts
}

func (u *NameCard) GetUpdateTS() types.Timestamp {
	return u.UpdateTS
}

func (u *NameCard) Get(isLocked bool) error {
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

func (u *NameCard) GetByID(isLocked bool) error {
	var err error

	val, err := u.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return u.Unmarshal(val)
}

func (u *NameCard) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{DBNameCardPrefix, u.ID[:]})
}

func (u *NameCard) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *NameCard) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *NameCard) GetSyncInfo() pkgservice.SyncInfo {
	if u.SyncInfo == nil {
		return nil
	}
	return u.SyncInfo
}

func (u *NameCard) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		u.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*SyncNameCardInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	u.SyncInfo = syncInfo

	return nil
}
