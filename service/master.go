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
	"github.com/ailabstw/go-pttai/pttdb"
)

type Master struct {
	*BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	TransferToID *types.PttID `json:"t,omitempty"`

	SyncInfo *SyncPersonInfo `json:"s,omitempty"`
}

func NewMaster(
	id *types.PttID,
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	db *pttdb.LDBBatch,
	dbLock *types.LockMap,
	fullDBPrefix []byte,
	fullDBIdxPrefix []byte,
) *Master {
	o := NewObject(id, createTS, creatorID, entityID, logID, status, db, dbLock, fullDBPrefix, fullDBIdxPrefix)

	return &Master{
		BaseObject: o,

		UpdateTS: createTS,
	}
}

func NewEmptyMaster() *Master {
	return &Master{BaseObject: &BaseObject{}}
}

func MastersToObjs(typedObjs []*Master) []Object {
	objs := make([]Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToMasters(objs []Object) []*Master {
	typedObjs := make([]*Master, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Master)
	}
	return typedObjs
}

/**********
 * SetObjDB
 **********/

func (pm *BaseProtocolManager) SetMasterObjDB(master *Master) {
	master.SetDB(pm.DB(), pm.DBObjLock(), pm.Entity().GetID(), pm.dbMasterPrefix, pm.dbMasterIdxPrefix)
}

func (m *Master) Save(isLocked bool) error {
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

	_, err = m.db.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (m *Master) NewEmptyObj() Object {
	return &Master{BaseObject: &BaseObject{EntityID: m.EntityID, db: m.db, dbLock: m.dbLock, fullDBPrefix: m.fullDBPrefix}}

}

func (m *Master) GetNewObjByID(id *types.PttID, isLocked bool) (Object, error) {
	newM := m.NewEmptyObj()
	err := newM.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newM, nil
}

func (m *Master) SetUpdateTS(ts types.Timestamp) {
	m.UpdateTS = ts
}

func (m *Master) GetUpdateTS() types.Timestamp {
	return m.UpdateTS
}

func (m *Master) GetByID(isLocked bool) error {
	var err error

	val, err := m.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return m.Unmarshal(val)
}

func (m *Master) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := m.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}
	return common.Concat([][]byte{m.fullDBPrefix, marshalTimestamp, m.ID[:]})
}

func (m *Master) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Master) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	return nil
}

/**********
 * Sync Info
 **********/

func (m *Master) GetSyncInfo() SyncInfo {
	if m.SyncInfo == nil {
		return nil
	}
	return m.SyncInfo
}

func (m *Master) SetSyncInfo(theSyncInfo SyncInfo) error {
	syncInfo, ok := theSyncInfo.(*SyncPersonInfo)
	if !ok {
		return ErrInvalidSyncInfo
	}

	m.SyncInfo = syncInfo

	return nil
}
