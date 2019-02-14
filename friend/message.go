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

package friend

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Message struct {
	*pkgservice.BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	SyncInfo *pkgservice.BaseSyncInfo `json:"s,omitempty"`
}

func NewMessage(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

) (*Message, error) {

	id, err := types.NewPttID()
	if err != nil {
		return nil, err
	}

	o := pkgservice.NewObject(id, createTS, creatorID, entityID, logID, status)

	return &Message{
		BaseObject: o,
		UpdateTS:   createTS,
	}, nil
}

func NewEmptyMessage() *Message {
	return &Message{BaseObject: &pkgservice.BaseObject{}}
}

func MessagesToObjs(typedObjs []*Message) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToMessages(objs []pkgservice.Object) []*Message {
	typedObjs := make([]*Message, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*Message)
	}
	return typedObjs
}

func AliveMessages(typedObjs []*Message) []*Message {
	objs := make([]*Message, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetMessageDB(m *Message) {
	m.SetDB(dbFriend, pm.DBObjLock(), pm.Entity().GetID(), pm.dbMessagePrefix, pm.dbMessageIdxPrefix, pm.SetBlockInfoDB, pm.SetMediaDB)
}

func (m *Message) Save(isLocked bool) error {
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

	log.Debug("Message.Save: to ForcePutAll", "idxKey", idxKey, "key", kvs[0].K)

	_, err = m.DB().ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (m *Message) NewEmptyObj() pkgservice.Object {
	newObj := NewEmptyMessage()
	newObj.CloneDB(m.BaseObject)
	return newObj
}

func (m *Message) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newObj := m.NewEmptyObj()
	newObj.SetID(id)
	err := newObj.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newObj, nil
}

func (m *Message) SetUpdateTS(ts types.Timestamp) {
	m.UpdateTS = ts
}

func (m *Message) GetUpdateTS() types.Timestamp {
	return m.UpdateTS
}

func (m *Message) Get(isLocked bool) error {
	var err error

	if !isLocked {
		err = m.RLock()
		if err != nil {
			return err
		}
		defer m.RUnlock()
	}

	key, err := m.MarshalKey()
	if err != nil {
		return err
	}

	val, err := m.DB().DBGet(key)
	if err != nil {
		return err
	}

	return m.Unmarshal(val)
}

func (m *Message) GetByID(isLocked bool) error {
	var err error

	val, err := m.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return m.Unmarshal(val)
}

func (m *Message) MarshalKey() ([]byte, error) {
	marshalTimestamp, err := m.CreateTS.Marshal()
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{m.FullDBPrefix(), marshalTimestamp, m.ID[:]})
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, m)
}

func (m *Message) GetSyncInfo() pkgservice.SyncInfo {
	if m.SyncInfo == nil {
		return nil
	}
	return m.SyncInfo
}

func (m *Message) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	if theSyncInfo == nil {
		m.SyncInfo = nil
		return nil
	}

	syncInfo, ok := theSyncInfo.(*pkgservice.BaseSyncInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	m.SyncInfo = syncInfo

	return nil
}

func (m *Message) DeleteAll(isLocked bool) error {
	var err error
	if !isLocked {
		err = m.Lock()
		if err != nil {
			return err
		}
		defer m.Unlock()
	}

	// block-info
	blockInfo := m.GetBlockInfo()
	setBlockInfoDB := m.SetBlockInfoDB()
	setBlockInfoDB(blockInfo, m.ID)

	blockInfo.Remove(false)

	// delete
	m.Delete(true)

	return nil
}
