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
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type MyInfo struct {
	*pkgservice.BaseEntity `json:"e"`

	UpdateTS types.Timestamp `json:"UT"`

	ProfileID *types.PttID     `json:"PID"`
	Profile   *account.Profile `json:"-"`

	BoardID *types.PttID   `json:"BID"`
	Board   *content.Board `json:"-"`

	signKeyInfo     *pkgservice.KeyInfo
	nodeSignKeyInfo *pkgservice.KeyInfo

	NodeSignID *types.PttID `json:"-"`

	myKey   *ecdsa.PrivateKey
	nodeKey *ecdsa.PrivateKey

	validateKey *types.PttID
}

func NewEmptyMyInfo() *MyInfo {
	return &MyInfo{BaseEntity: &pkgservice.BaseEntity{SyncInfo: &pkgservice.BaseSyncInfo{}}}
}

func NewMyInfo(id *types.PttID, myKey *ecdsa.PrivateKey, ptt pkgservice.MyPtt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager, dbLock *types.LockMap) (*MyInfo, error) {
	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	e := pkgservice.NewBaseEntity(id, ts, id, types.StatusPending, dbMe, dbLock)

	m := &MyInfo{
		BaseEntity: e,
		UpdateTS:   ts,
		myKey:      myKey,
	}

	// new my node
	myNodeID := ptt.MyNodeID()
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

	err = m.Init(ptt, service, spm)
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

func (m *MyInfo) Init(thePtt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager) error {

	log.Debug("me.Init: start")
	ptt, ok := thePtt.(pkgservice.MyPtt)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	MyID := spm.(*ServiceProtocolManager).MyID
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
	m.NodeSignID = nodeSignID

	m.validateKey, err = types.NewPttID()
	if err != nil {
		return err
	}

	// my-key
	if m.myKey == nil {
		m.myKey, err = m.loadMyKey()
		if err != nil {
			if !reflect.DeepEqual(myID, MyID) {
				return nil
			}
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

	// set my entity
	if !reflect.DeepEqual(myID, MyID) {
		return nil
	}

	// profile
	accountSPM := service.(*Backend).accountBackend.SPM()
	if m.ProfileID != nil {
		profile := accountSPM.Entity(m.ProfileID)
		if profile == nil {
			return pkgservice.ErrInvalidEntity
		}
		m.Profile = profile.(*account.Profile)
	}

	// board
	contentSPM := service.(*Backend).contentBackend.SPM()
	if m.BoardID != nil {
		board := contentSPM.Entity(m.BoardID)
		if board == nil {
			return pkgservice.ErrInvalidEntity
		}
		m.Board = board.(*content.Board)
	}

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

	m.BaseEntity.Init(pm, ptt, service)

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

	err = dbMeCore.Put(key, marshaled)
	if err != nil {
		return err
	}

	return nil
}

func (m *MyInfo) MustSave(isLocked bool) error {
	if !isLocked {
		m.MustLock()
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

	err = dbMeCore.Put(key, marshaled)
	if err != nil {
		return err
	}

	return nil

}

// Remember to do InitPM when necessary.
/*
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
*/

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

func (m *MyInfo) GetProfile() pkgservice.Entity {
	return m.Profile
}

func (m *MyInfo) GetBoard() pkgservice.Entity {
	return m.Board
}

func (m *MyInfo) GetUserNodeID(id *types.PttID) (*discover.NodeID, error) {
	friendBackend := m.Service().(*Backend).friendBackend

	theFriend, err := friendBackend.SPM().(*friend.ServiceProtocolManager).GetFriendEntityByFriendID(id)
	if err != nil {
		return nil, err
	}
	if theFriend == nil {
		return nil, types.ErrInvalidID
	}

	friendPM := theFriend.PM().(*friend.ProtocolManager)

	return friendPM.GetUserNodeID()
}
