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

package service

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

type Member struct {
	*BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	TransferToID *types.PttID `json:"t,omitempty"`

	SyncInfo *SyncPersonInfo `json:"s,omitempty"`
}

func NewMember(
	id *types.PttID,
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,
) *Member {

	o := NewObject(id, createTS, creatorID, entityID, logID, status)

	return &Member{
		BaseObject: o,

		UpdateTS: createTS,
	}
}

func NewEmptyMember() *Member {
	return &Member{BaseObject: &BaseObject{}}
}

func MembersToObjs(typedObjs []*Member) []Object {
	objs := make([]Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToMembers(objs []Object) []*Member {
	typedObjs := make([]*Member, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Member)
	}
	return typedObjs
}

func (m *Member) Save(isLocked bool) error {
	var err error

	log.Debug("Member.Save: start", "id", m.ID, "entityID", m.EntityID)

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

func (m *Member) NewEmptyObj() Object {
	obj := m.BaseObject.NewEmptyObj()
	return &Member{BaseObject: obj}
}

func (pm *BaseProtocolManager) SetMemberObjDB(member *Member) {
	member.SetDB(pm.DB(), pm.DBObjLock(), pm.Entity().GetID(), pm.dbMemberPrefix, pm.dbMemberIdxPrefix, nil, nil)
}

func (m *Member) GetNewObjByID(id *types.PttID, isLocked bool) (Object, error) {
	newM := m.NewEmptyObj()
	newM.SetID(id)
	err := newM.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newM, nil
}

func (m *Member) SetUpdateTS(ts types.Timestamp) {
	m.UpdateTS = ts
}

func (m *Member) GetUpdateTS() types.Timestamp {
	return m.UpdateTS
}

func (m *Member) GetByID(isLocked bool) error {
	var err error

	val, err := m.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return m.Unmarshal(val)
}

func (m *Member) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := m.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{m.fullDBPrefix, marshalTimestamp, m.ID[:]})
}

func (m *Member) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Member) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	return nil
}

/**********
 * Sync Info
 **********/

func (m *Member) GetSyncInfo() SyncInfo {
	if m.SyncInfo == nil {
		return nil
	}
	return m.SyncInfo
}

func (m *Member) SetSyncInfo(theSyncInfo SyncInfo) error {
	if theSyncInfo == nil {
		m.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*SyncPersonInfo)
	if !ok {
		return ErrInvalidSyncInfo
	}

	m.SyncInfo = syncInfo

	return nil
}
