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

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

/**********
 * OpKey
 **********/

func (m *MyInfo) NewOpKeyInfo(entityID *types.PttID, setOpKeyObjDB func(k *pkgservice.KeyInfo)) (*pkgservice.KeyInfo, error) {

	key, err := pkgservice.NewOpKeyInfo(entityID, m.ID, m.myKey)
	if err != nil {
		return nil, err
	}
	setOpKeyObjDB(key)

	return key, nil
}

/**********
 * Sign
 **********/

func (m *MyInfo) Sign(oplog *pkgservice.BaseOplog) error {
	signKey := m.SignKey()

	return oplog.Sign(signKey)
}

func (m *MyInfo) InternalSign(oplog *pkgservice.BaseOplog) error {
	nodeSignID := m.NodeSignID
	nodeSignKey := m.NodeSignKey()

	return oplog.InternalSign(nodeSignID, nodeSignKey)
}

func (m *MyInfo) MasterSign(oplog *pkgservice.BaseOplog) error {
	signKey := m.SignKey()

	return oplog.MasterSign(m.ID, signKey)
}

func (m *MyInfo) MyMasterSign(oplog *pkgservice.BaseOplog) error {

	nodeSignID := m.NodeSignID
	nodeSignKey := m.NodeSignKey()

	return oplog.MasterSign(nodeSignID, nodeSignKey)
}

func (m *MyInfo) SignBlock(block *pkgservice.Block) error {

	signKey := m.SignKey()

	return block.Sign(signKey)
}

/**********
 * SignKey
 **********/

func (m *MyInfo) CreateSignKeyInfo() error {
	keyInfo, err := pkgservice.NewSignKeyInfo(m.ID, m.myKey)
	if err != nil {
		return err
	}

	m.signKeyInfo = keyInfo

	return nil
}

func (m *MyInfo) SignKey() *pkgservice.KeyInfo {
	if m.signKeyInfo.Count > NRenewSignKey {
		m.CreateSignKeyInfo()
	}

	return m.signKeyInfo
}

func (m *MyInfo) GetMyKey() *ecdsa.PrivateKey {
	return m.myKey
}

/**********
 * NodeSignKey
 **********/

func (m *MyInfo) CreateNodeSignKeyInfo() error {
	keyInfo, err := pkgservice.NewSignKeyInfo(m.NodeSignID, m.nodeKey)
	if err != nil {
		return err
	}

	m.nodeSignKeyInfo = keyInfo

	return nil
}

func (m *MyInfo) GetNodeSignID() *types.PttID {
	return m.NodeSignID
}

func (m *MyInfo) NodeSignKey() *pkgservice.KeyInfo {
	if m.nodeSignKeyInfo.Count > NRenewSignKey {
		m.CreateNodeSignKeyInfo()
	}
	return m.nodeSignKeyInfo
}

func (m *MyInfo) GetNodeKey() *ecdsa.PrivateKey {
	return m.nodeKey
}
