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
	"bytes"
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

type SyncNodeNameInfo struct {
	LogID    *types.PttID    `json:"pl,omitempty"`
	NodeName []byte          `json:"N,omitempty"`
	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`
}

type MyNode struct {
	V        types.Version
	ID       *types.PttID
	CreateTS types.Timestamp `json:"CT"`
	UpdateTS types.Timestamp `json:"UT"`

	Status types.Status `json:"S"`

	LastSeen types.Timestamp `json:"L"`

	NodeName         []byte            `json:"N,omitempty"`
	NodeNameLogID    *types.PttID      `json:"nl,omitempty"`
	SyncNodeNameInfo *SyncNodeNameInfo `json:"sn,omitempty"`

	NodeID *discover.NodeID `json:"NID"`

	NodeType pkgservice.NodeType `json:"NT"`
	Weight   uint32              `json:"W"`

	LogID *types.PttID `json:"pl,omitempty"`

	RaftID uint64 `json:"R"`

	Peer *pkgservice.PttPeer `json:"-"`
}

func NewMyNode(ts types.Timestamp, myID *types.PttID, nodeID *discover.NodeID, weight uint32) (*MyNode, error) {
	log.Debug("NewMyNode: start", "myID", myID, "nodeID", nodeID)
	raftID, err := nodeID.ToRaftID()
	if err != nil {
		return nil, err
	}

	return &MyNode{
		V:        types.CurrentVersion,
		ID:       myID,
		CreateTS: ts,
		UpdateTS: ts,

		Status: types.StatusInit,

		NodeID: nodeID,
		Weight: weight,
		RaftID: raftID,
	}, nil
}

func (m *MyNode) Save() ([]byte, error) {
	key, err := m.MarshalKey()
	if err != nil {
		return nil, err
	}

	marshaled, err := m.Marshal()
	if err != nil {
		return nil, err
	}

	err = dbMyNodes.Put(key, marshaled)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *MyNode) Delete(isLocked bool) error {
	key, err := m.MarshalKey()
	if err != nil {
		return err
	}

	return dbMyNodes.Delete(key)
}

func (m *MyNode) Get(myID *types.PttID, nodeID *discover.NodeID) error {
	m.ID = myID
	m.NodeID = nodeID
	key, err := m.MarshalKey()
	if err != nil {
		return err
	}

	theBytes, err := dbMyNodes.Get(key)
	if err != nil {
		log.Error("unable to Get", "nodeID", nodeID, "e", err)
		return err
	}

	if theBytes == nil {
		log.Error("bytes == nil", "nodeID", nodeID, "e", err)
		return leveldb.ErrNotFound
	}

	if len(theBytes) == 0 {
		log.Error("len(bytes) == 0", "nodeID", nodeID, "e", err)
		return leveldb.ErrNotFound
	}

	err = m.Unmarshal(theBytes)
	log.Debug("Get: after Get", "key", key, "theBytes", theBytes, "m", m)

	if err != nil {
		return err
	}

	if m.Status == types.StatusDeleted {
		log.Error("id deleted")
		return types.ErrInvalidID
	}

	return nil
}

func (m *MyNode) DBPrefix() ([]byte, error) {
	return append(DBMyNodePrefix, m.ID[:]...), nil
}

func (m *MyNode) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{DBMyNodePrefix, m.ID[:], m.NodeID[:]})
}

func (m *MyNode) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MyNode) Unmarshal(theBytes []byte) error {
	err := json.Unmarshal(theBytes, m)
	if err != nil {
		return err
	}

	return nil
}

func (m *MyNode) DeleteRawKey(key []byte) error {
	err := dbMyNodes.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

func (m *MyNode) IntegrateSyncNodeNameInfo(info *SyncNodeNameInfo) (*types.PttID, error) {
	var origLogID *types.PttID
	switch {
	case m.SyncNodeNameInfo == nil:
		m.SyncNodeNameInfo = info
		return nil, nil
	case info.Status != types.StatusInternalSync && m.SyncNodeNameInfo.Status > info.Status:
		return nil, nil
	case m.SyncNodeNameInfo.Status < info.Status:
		origLogID = m.SyncNodeNameInfo.LogID
		m.SyncNodeNameInfo = info
		return origLogID, nil
	case info.UpdateTS.IsLess(m.SyncNodeNameInfo.UpdateTS):
		return nil, nil
	case m.SyncNodeNameInfo.UpdateTS.IsLess(info.UpdateTS):
		origLogID = m.SyncNodeNameInfo.LogID
		m.SyncNodeNameInfo = info
		return origLogID, nil
	}

	cmp := bytes.Compare(m.SyncNodeNameInfo.LogID[:], info.LogID[:])
	if cmp < 0 {
		return nil, nil
	}

	origLogID = m.SyncNodeNameInfo.LogID
	m.SyncNodeNameInfo = info
	return origLogID, nil
}
