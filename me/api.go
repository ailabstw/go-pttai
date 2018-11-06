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
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type PrivateAPI struct {
	b *Backend
}

func NewPrivateAPI(b *Backend) *PrivateAPI {
	return &PrivateAPI{b}
}

/**********
 * Set
 **********/

func (api *PrivateAPI) SetMyName(name []byte) (*account.UserName, error) {
	return api.b.SetMyName(name)
}

func (api *PrivateAPI) SetMyNodeName(nodeID string, name []byte) (*MyNode, error) {
	return api.b.SetMyNodeName([]byte(nodeID), name)
}

func (api *PrivateAPI) SetMyImage(imgStr string) (*account.UserImg, error) {
	return api.b.SetMyImage(imgStr)
}

/**********
 * Revoke
 **********/

func (api *PrivateAPI) Revoke(keyBytes []byte) error {
	return api.b.Revoke(keyBytes)
}

/**********
 * JoinMe
 **********/

func (api *PrivateAPI) ShowMeURL() (*pkgservice.BackendJoinURL, error) {
	return api.b.ShowMeURL()
}

func (api *PrivateAPI) JoinMe(meURL string, myKey string, dummy bool) (*pkgservice.BackendJoinRequest, error) {

	return api.b.JoinMe([]byte(meURL), []byte(myKey))
}

func (api *PrivateAPI) GetJoinKeyInfos() ([]*pkgservice.KeyInfo, error) {
	return api.b.BEGetJoinKeyInfos()
}

/*
GetMeRequests get the me-requests from me to the others.
*/
func (api *PrivateAPI) GetMeRequests() ([]*pkgservice.BackendJoinRequest, error) {
	return api.b.GetMeRequests()
}

/**********
 * JoinFriend
 **********/

func (api *PrivateAPI) JoinFriend(friendURL string) (*pkgservice.BackendJoinRequest, error) {
	return api.b.JoinFriend([]byte(friendURL))
}

/*
GetFriendRequests get the friend-requests from me to the others.
*/
func (api *PrivateAPI) GetFriendRequests() ([]*pkgservice.BackendJoinRequest, error) {
	return api.b.GetFriendRequests()
}

/**********
 * Op
 **********/

func (api *PrivateAPI) GetOpKeyInfos() ([]*pkgservice.KeyInfo, error) {
	return api.b.BEGetOpKeyInfos()
}

func (api *PrivateAPI) RevokeOpKey(keyID string, myKey string) (bool, error) {
	return api.b.RevokeOpKey([]byte(keyID), []byte(myKey))
}

func (api *PrivateAPI) GetOpKeyInfosFromDB() ([]*pkgservice.KeyInfo, error) {
	return api.b.GetOpKeyInfosFromDB()
}

/**********
 * Get
 **********/

func (api *PrivateAPI) GetMyBoard() (*content.BackendGetBoard, error) {
	return api.b.GetMyBoard()
}

func (api *PrivateAPI) GetRawMe() (*MyInfo, error) {
	return api.b.GetRawMe()
}

func (api *PrivateAPI) GetRawMeByID(id string) (*MyInfo, error) {
	return api.b.GetRawMeByID([]byte(id))
}

/**********
 * Peer
 **********/

func (api *PrivateAPI) CountPeers() (int, error) {
	return api.b.CountPeers()
}

func (api *PrivateAPI) GetPeers() ([]*pkgservice.BackendPeer, error) {
	return api.b.GetPeers()
}

/**********
 * Raft / Node
 **********/

func (api *PrivateAPI) GetRaftStatus(id string) (*RaftStatus, error) {
	return api.b.GetRaftStatus([]byte(id))
}

func (api *PrivateAPI) ForceRemoveNode(nodeID string) (bool, error) {
	return api.b.ForceRemoveNode(nodeID)
}

func (api *PrivateAPI) GetMyNodes() ([]*MyNode, error) {
	return api.b.GetMyNodes()
}

func (api *PrivateAPI) GetTotalWeight() uint32 {
	return api.b.GetTotalWeight()
}

/**********
 * MeOplog
 **********/

func (api *PrivateAPI) GetMeOplogList(logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {
	return api.b.BEGetMeOplogList([]byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMeOplogMasterList(logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {
	return api.b.BEGetPendingMeOplogMasterList([]byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMeOplogInternalList(logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {
	return api.b.BEGetPendingMeOplogInternalList([]byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMeOplogMerkleNodeList(level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.BEGetMeOplogMerkleNodeList(pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MasterOplog
 **********/

func (api *PrivateAPI) GetMasterOplogList(logID string, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {
	return api.b.BEGetMasterOplogList([]byte(logID), limit, listOrder)
}

/**********
 * OpKeyOplog
 **********/

func (api *PrivateAPI) GetOpKeyOplogList(logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.BEGetOpKeyOplogList([]byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogMasterList(logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.BEGetPendingOpKeyOplogMasterList([]byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogInternalList(logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.BEGetPendingOpKeyOplogInternalList([]byte(logID), limit, listOrder)
}

/**********
 * Key
 **********/

func (api *PrivateAPI) ShowMyKey() (*types.PttID, error) {
	return api.b.ShowValidateKey()
}

func (api *PrivateAPI) ValidateMyKey(key string) (bool, error) {
	return api.b.ValidateValidateKey([]byte(key))
}

func (api *PrivateAPI) ShowMyMasterKey() ([]byte, error) {
	return api.b.ShowMyMasterKey()
}

func (api *PrivateAPI) ValidateMyMasterKey(masterKeyBytes []byte) (bool, error) {
	return api.b.ValidateMyMasterKey(masterKeyBytes)
}

func (api *PrivateAPI) ShowMyNodeKey() ([]byte, error) {
	return api.b.ShowMyNodeKey()
}

func (api *PrivateAPI) ValidateMyNodeKey(nodeKeyBytes []byte) (bool, error) {
	return api.b.ValidateMyNodeKey(nodeKeyBytes)
}

func (api *PrivateAPI) ShowMySignKey() (*pkgservice.KeyInfo, error) {
	return api.b.ShowMySignKey()
}

func (api *PrivateAPI) RefreshMySignKey() (*pkgservice.KeyInfo, error) {
	return api.b.RefreshMySignKey()
}

func (api *PrivateAPI) ShowMyNodeSignKey() (*pkgservice.KeyInfo, error) {
	return api.b.ShowMyNodeSignKey()
}

func (api *PrivateAPI) RefreshMyNodeSignKey() (*pkgservice.KeyInfo, error) {
	return api.b.RefreshMyNodeSignKey()
}

/**********
 * public
 **********/

type PublicAPI struct {
	b *Backend
}

func NewPublicAPI(b *Backend) *PublicAPI {
	return &PublicAPI{b}
}

func (api *PublicAPI) Get() (*BackendMyInfo, error) {
	return api.b.Get()
}

func (api *PublicAPI) ShowURL() (*pkgservice.BackendJoinURL, error) {
	return api.b.ShowURL()
}

func (api *PublicAPI) GetMeList() ([]*BackendMyInfo, error) {
	return api.b.GetMeList()
}
