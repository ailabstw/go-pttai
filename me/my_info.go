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
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

type MyInfo struct {
	*pkgservice.BaseEntity `json:"-"`

	V        types.Version
	ID       *types.PttID
	CreateTS types.Timestamp `json:"CT"`
	UpdateTS types.Timestamp `json:"UT"`

	Status types.Status `json:"S"`

	LogID *types.PttID `json:"l"`

	OwnerID *types.PttID `json:"o"`

	signKeyInfo     *pkgservice.KeyInfo
	nodeSignKeyInfo *pkgservice.KeyInfo

	nodeSignID *types.PttID

	myKey   *ecdsa.PrivateKey
	nodeKey *ecdsa.PrivateKey

	meOplogMerkle     *pkgservice.Merkle
	masterOplogMerkle *pkgservice.Merkle

	validateKey *types.PttID
}

func NewMyInfo(id *types.PttID, myKey *ecdsa.PrivateKey, ptt pkgservice.MyPtt, service pkgservice.Service) (*MyInfo, error) {
	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	m := &MyInfo{
		V:        types.CurrentVersion,
		ID:       id,
		CreateTS: ts,
		UpdateTS: ts,
		myKey:    myKey,
	}
	m.Status = types.StatusPending

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

	err = m.Init(ptt, service, id)
	if err != nil {
		return nil, err
	}
	m.OwnerID = id

	return m, nil
}

func (m *MyInfo) Init(ptt pkgservice.MyPtt, service pkgservice.Service, MyID *types.PttID) error {
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

	// merkle

	m.meOplogMerkle, err = pkgservice.NewMerkle(DBMeOplogPrefix, DBMeMerkleOplogPrefix, myID, dbOplog)
	if err != nil {
		return err
	}

	m.masterOplogMerkle, err = pkgservice.NewMerkle(DBMasterOplogPrefix, DBMasterMerkleOplogPrefix, myID, dbOplog)
	if err != nil {
		return err
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

	baseEntity, err := pkgservice.NewBaseEntity(pm, string(name), ptt, service)
	if err != nil {
		log.Error("InitPM: unable to NewBaseEntity", "e", err)
		return err
	}

	m.BaseEntity = baseEntity

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

func (m *MyInfo) Save() error {
	key, err := m.MarshalKey()
	if err != nil {
		return err
	}

	marshaled, err := m.Marshal()
	if err != nil {
		return err
	}

	_, err = dbMe.TryPut(key, marshaled, m.UpdateTS)

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

	theBytes, err := dbMe.Get(key)
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

/*
Revoke intends to Revoke the id.

    1. check me
    2. check whether revoke channel is busy.
    3. mark deleted in local-db.
    4. broadcast to the network.
    5. stop node.
    6. clear node.key-store.
    7. exit.
*/

func (m *MyInfo) Revoke(keyBytes []byte) error {
	// check me
	key, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return err
	}

	if !m.ID.IsSameKey(key) {
		return ErrInvalidMe
	}

	log.Info("Same Key. To revoke", "m.ID", m.ID, "key", key)

	m.Status = types.StatusDeleted

	err = m.Save()
	if err != nil {
		return err
	}

	return nil
}

func (m *MyInfo) GetID() *types.PttID {
	return m.ID
}

func (m *MyInfo) GetCreateTS() types.Timestamp {
	return m.CreateTS
}

func (m *MyInfo) DB() *pttdb.LDBBatch {
	return dbMeBatch
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

func (m *MyInfo) GetStatus() types.Status {
	return m.Status
}

func (m *MyInfo) MyPM() pkgservice.MyProtocolManager {
	return m.PM().(*ProtocolManager)
}

func (m *MyInfo) MeOplogMerkle() *pkgservice.Merkle {
	return m.meOplogMerkle
}

func (m *MyInfo) MasterOplogMerkle() *pkgservice.Merkle {
	return m.masterOplogMerkle
}

func (m *MyInfo) GetMasterKey() *ecdsa.PrivateKey {
	return m.myKey
}

func (m *MyInfo) GetValidateKey() *types.PttID {
	return m.validateKey
}
