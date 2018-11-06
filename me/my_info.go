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

package me

import (
	"crypto/ecdsa"
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

type MyInfo struct {
	*pkgservice.BaseEntity `json:"e"`

	UpdateTS types.Timestamp `json:"UT"`

	signKeyInfo     *pkgservice.KeyInfo
	nodeSignKeyInfo *pkgservice.KeyInfo

	nodeSignID *types.PttID

	myKey   *ecdsa.PrivateKey
	nodeKey *ecdsa.PrivateKey

	validateKey *types.PttID
}

func NewMyInfo(id *types.PttID, myKey *ecdsa.PrivateKey, ptt pkgservice.MyPtt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager, dbLock *types.LockMap) (*MyInfo, error) {
	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	e := pkgservice.NewBaseEntity(id, ts, id, types.StatusPending, dbMe, dbLock)

	m := &MyInfo{
		BaseEntity: e,
		UpdateTS:   ts,
		myKey:      myKey,
	}

	// nodeSignID
	myNodeID := ptt.MyNodeID()

	// new my node
	myNode, err := NewMyNode(ts, id, myNodeID, 1)
	if err != nil {
		return nil, err
	}

	myNode.Status = types.StatusAlive
	myNode.NodeType = ptt.MyNodeType()

	_, err = myNode.Save()
	if err != nil {
		return nil, err
	}

	err = m.Init(ptt, service, spm, id)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MyInfo) GetUpdateTS() types.Timestamp {
	return m.UpdateTS
}

func (m *MyInfo) SetUpdateTS(ts types.Timestamp) {
	m.UpdateTS = ts
}

func (m *MyInfo) Init(ptt pkgservice.MyPtt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager, MyID *types.PttID) error {
	m.SetDB(dbMe, spm.GetDBLock())

	err := m.InitPM(ptt, service)
	if err != nil {
		return err
	}

	myID := m.ID
	nodeKey := ptt.MyNodeKey()
	nodeID := ptt.MyNodeID()
	nodeSignID, err := setNodeSignID(nodeID, myID)

	m.nodeKey = nodeKey
	m.nodeSignID = nodeSignID

	m.validateKey, err = types.NewPttID()
	if err != nil {
		return err
	}

	// my-key
	if m.myKey == nil {
		m.myKey, err = m.loadMyKey()
		if err != nil {
			return err
		}
	}

	// sign-key
	err = m.CreateSignKeyInfo()
	if err != nil {
		return err
	}

	err = m.CreateNodeSignKeyInfo()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(myID, MyID) {
		return nil
	}

	// set my entity

	ptt.SetMyEntity(m)

	return nil
}

func (m *MyInfo) loadMyKey() (*ecdsa.PrivateKey, error) {
	cfg := m.Service().(*Backend).Config

	return cfg.GetDataPrivateKeyByID(m.ID)
}

func (m *MyInfo) InitPM(ptt pkgservice.MyPtt, service pkgservice.Service) error {
	pm, err := NewProtocolManager(m, ptt)
	if err != nil {
		log.Error("InitPM: unable to NewProtocolManager", "e", err)
		return err
	}

	userName := &account.UserName{}
	err = userName.Get(m.ID, true)
	if err == leveldb.ErrNotFound {
		err = nil
	}
	if err != nil {
		return err
	}
	name := userName.Name
	if name == nil {
		name = []byte{}
	}

	m.BaseEntity.Init(pm, string(name), ptt, service)

	return nil
}

func (m *MyInfo) MarshalKey() ([]byte, error) {
	key := append(DBMePrefix, m.ID[:]...)

	return key, nil
}

func (m *MyInfo) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MyInfo) Unmarshal(theBytes []byte) error {
	err := json.Unmarshal(theBytes, m)
	if err != nil {
		return err
	}

	// postprocess

	return nil
}

func (m *MyInfo) Save(isLocked bool) error {
	if !isLocked {
		err := m.Lock()
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

	_, err = dbMeCore.TryPut(key, marshaled, m.UpdateTS)

	if err != nil {
		return err
	}

	return nil
}

// Remember to do InitPM when necessary.
func (m *MyInfo) Get(id *types.PttID, ptt pkgservice.Ptt, service pkgservice.Service, contentBackend *content.Backend) error {
	m.ID = id

	key, err := m.MarshalKey()
	if err != nil {
		return err
	}

	theBytes, err := dbMeCore.Get(key)
	log.Debug("Get: after get from dbMe", "theBytes", theBytes, "e", err)
	if err != nil {
		return err
	}

	if len(theBytes) == 0 {
		return types.ErrInvalidID
	}

	err = m.Unmarshal(theBytes)
	log.Debug("Get: after Unmarshal", "theBytes", theBytes, "m", m, "e", err)
	if err != nil {
		return err
	}

	return nil
}

func (m *MyInfo) GetJoinRequest(hash *common.Address) (*pkgservice.JoinRequest, error) {
	return m.PM().(*ProtocolManager).GetJoinRequest(hash)
}

func (m *MyInfo) GetLenNodes() int {
	return len(m.PM().(*ProtocolManager).MyNodes)
}

func (m *MyInfo) IsValidInternalOplog(signInfos []*pkgservice.SignInfo) (*types.PttID, uint32, bool) {
	return m.PM().(*ProtocolManager).IsValidInternalOplog(signInfos)
}

func (m *MyInfo) MyPM() pkgservice.MyProtocolManager {
	return m.PM().(*ProtocolManager)
}

func (m *MyInfo) GetMasterKey() *ecdsa.PrivateKey {
	return m.myKey
}

func (m *MyInfo) GetValidateKey() *types.PttID {
	return m.validateKey
}

func (m *MyInfo) SetPendingDeleteSyncInfo(status types.Status, oplog *pkgservice.BaseOplog) {
	syncInfo := &pkgservice.BaseSyncInfo{
		LogID:     oplog.ID,
		UpdateTS:  oplog.UpdateTS,
		UpdaterID: oplog.CreatorID,
		Status:    status,
	}
	m.IntegrateSyncInfo(syncInfo)
}

func (m *MyInfo) RemoveSyncInfo(oplog *pkgservice.BaseOplog, opData pkgservice.OpData, syncInfo pkgservice.SyncInfo, info pkgservice.ProcessInfo) error {
	return types.ErrNotImplemented
}

func (m *MyInfo) UpdateDeleteInfo(oplog *pkgservice.BaseOplog, theInfo pkgservice.ProcessInfo) {
	info, ok := theInfo.(*ProcessMeInfo)
	if !ok {
		return
	}

	info.DeleteMeInfo[*m.ID] = oplog

	return
}
