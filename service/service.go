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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/ethereum/go-ethereum/rpc"
)

type Backend interface {
	// master-oplog

	GetMasterOplogList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error)
	GetPendingMasterOplogMasterList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error)
	GetPendingMasterOplogInternalList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error)
	GetMasterOplogMerkleNodeList(idBytes []byte, level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendMerkleNode, error)

	// master
	GetMasterListFromCache(idBytes []byte) ([]*Master, error)
	GetMasterList(idBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*Master, error)

	// member-oplog

	GetMemberOplogList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*MemberOplog, error)
	GetPendingMemberOplogMasterList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*MemberOplog, error)
	GetPendingMemberOplogInternalList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*MemberOplog, error)
	GetMemberOplogMerkleNodeList(idBytes []byte, level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendMerkleNode, error)

	// member

	GetMemberList(idBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*Master, error)

	// op-key-oplog

	GetOpKeyOplogList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*OpKeyOplog, error)
	GetPendingOpKeyOplogMasterList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*OpKeyOplog, error)
	GetPendingOpKeyOplogInternalList(idBytes []byte, logID []byte, limit int, listOrder pttdb.ListOrder) ([]*OpKeyOplog, error)

	// op-key

	ShowValidateKey() (*types.PttID, error)
	ValidateValidateKey(keyBytes []byte) (bool, error)

	GetOpKeyInfos([]*KeyInfo, error)
	RevokeOpKey(keyIDBytes []byte, myKeyBytes []byte) (bool, error)
	GetOpKeyInfosFromDB() ([]*KeyInfo, error)

	// peers
	CountPeers() (int, error)
	GetPeers() ([]*BackendPeer, error)
}

type Service interface {
	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []rpc.API

	// Start is called after all services have been constructed and the networking
	// layer was not initialized yet.
	Prestart() error

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error

	SPM() ServiceProtocolManager

	Name() string

	Ptt() Ptt
}

type MyService interface {
	Service
}

/*
BaseService implements the base-type of Service
*/
type BaseService struct {
	spm ServiceProtocolManager
	ptt Ptt
}

func NewBaseService(ptt Ptt, spm ServiceProtocolManager) (*BaseService, error) {
	return &BaseService{ptt: ptt, spm: spm}, nil
}

func (svc *BaseService) APIs() []rpc.API {
	return nil
}

func (svc *BaseService) Prestart() error {
	return svc.SPM().Prestart()
}

func (svc *BaseService) Start() error {
	return svc.SPM().Start()
}

func (svc *BaseService) Stop() error {
	return svc.SPM().Stop()
}

func (svc *BaseService) SPM() ServiceProtocolManager {
	return svc.spm
}

func (svc *BaseService) Ptt() Ptt {
	return svc.ptt
}
