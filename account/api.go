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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
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
 * UserOplog
 **********/

func (api *PrivateAPI) GetUserOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {
	return api.b.GetUserOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingUserOplogMasterList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {
	return api.b.GetPendingUserOplogMasterList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingUserOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {
	return api.b.GetPendingUserOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetUserOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetUserOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * UserNode
 **********/

func (api *PrivateAPI) GetUserNodeList(idStr string, startID string, limit int, listOrder pttdb.ListOrder) ([]*UserNode, error) {
	return api.b.GetUserNodeList([]byte(idStr), []byte(startID), limit, listOrder)
}

func (api *PrivateAPI) GetUserNodeInfo(idStr string) (*UserNodeInfo, error) {
	return api.b.GetUserNodeInfo([]byte(idStr))
}

func (api *PrivateAPI) RemoveUserNode(entityIDStr string, nodeIDStr string) (types.Bool, error) {
	return api.b.RemoveUserNode([]byte(entityIDStr), []byte(nodeIDStr))
}

/**********
 * Raw Data
 **********/

func (api *PrivateAPI) GetRawUserName(idStr string) (*UserName, error) {
	return api.b.GetRawUserName([]byte(idStr))
}

func (api *PrivateAPI) GetRawUserImg(idStr string) (*UserImg, error) {
	return api.b.GetRawUserImg([]byte(idStr))
}

func (api *PrivateAPI) GetRawProfile(idStr string) (*Profile, error) {
	return api.b.GetRawProfile([]byte(idStr))
}

/**********
 * MasterOplog
 **********/

func (api *PrivateAPI) GetMasterOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetMasterOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogMasterList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetPendingMasterOplogMasterList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetPendingMasterOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMasterOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetMasterOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MemberOplog
 **********/

func (api *PrivateAPI) GetMemberOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetMemberOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogMasterList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetPendingMemberOplogMasterList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetPendingMemberOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMemberOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetMemberOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * OpKeyOplog
 **********/

func (api *PrivateAPI) GetOpKeyOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.GetOpKeyOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.GetPendingOpKeyOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.GetPendingOpKeyOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

/**********
 * Master
 **********/

func (api *PrivateAPI) GetMasterListFromCache(entityID string) ([]*pkgservice.Master, error) {
	return api.b.GetMasterListFromCache([]byte(entityID))
}

func (api *PrivateAPI) GetMasterList(entityID string, startID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.Master, error) {
	return api.b.GetMasterList([]byte(entityID), []byte(startID), limit, listOrder)
}

/**********
 * Member
 **********/

func (api *PrivateAPI) GetMemberList(entityID string, startID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.Member, error) {
	return api.b.GetMemberList([]byte(entityID), []byte(startID), limit, listOrder)
}

/**********
 * Op
 **********/

func (api *PrivateAPI) ShowValidateKey() (*types.PttID, error) {
	return api.b.ShowValidateKey()
}

func (api *PrivateAPI) ValidateValidateKey(key string) (bool, error) {
	return api.b.ValidateValidateKey([]byte(key))
}

func (api *PrivateAPI) GetOpKeyInfos(entityID string) ([]*pkgservice.KeyInfo, error) {
	return api.b.GetOpKeys([]byte(entityID))
}

func (api *PrivateAPI) RevokeOpKey(entityID string, keyID string, myKey string) (bool, error) {
	return api.b.RevokeOpKey([]byte(entityID), []byte(keyID), []byte(myKey))
}

func (api *PrivateAPI) GetOpKeyInfosFromDB(entityID string) ([]*pkgservice.KeyInfo, error) {
	return api.b.GetOpKeysFromDB([]byte(entityID))
}

/**********
 * Peer
 **********/

func (api *PrivateAPI) CountPeers(profileID string) (int, error) {
	return api.b.CountPeers([]byte(profileID))
}

func (api *PrivateAPI) GetPeers(profileID string) ([]*pkgservice.BackendPeer, error) {
	return api.b.GetPeers([]byte(profileID))
}

/**********
 * Public
 **********/

type PublicAPI struct {
	b *Backend
}

func NewPublicAPI(b *Backend) *PublicAPI {
	return &PublicAPI{b}
}

func (api *PublicAPI) GetUserName(idStr string) (*BackendUserName, error) {
	return api.b.GetUserName([]byte(idStr))
}

/*
func (api *PublicAPI) GetUserNameList(idStr string, limit int, listOrder pttdb.ListOrder) ([]*BackendUserName, error) {
	return api.b.GetUserNameList([]byte(idStr), limit, listOrder)
}
*/

func (api *PublicAPI) GetUserNameByIDs(idStrs []string) (map[string]*BackendUserName, error) {
	idByteList := make([][]byte, len(idStrs))
	for i, idStr := range idStrs {
		idByteList[i] = []byte(idStr)
	}
	return api.b.GetUserNameByIDs(idByteList)
}

func (api *PublicAPI) GetUserImg(idStr string) (*BackendUserImg, error) {
	return api.b.GetUserImg([]byte(idStr))
}

func (api *PublicAPI) GetUserImgByIDs(idStrs []string) (map[string]*BackendUserImg, error) {
	idByteList := make([][]byte, len(idStrs))
	for i, idStr := range idStrs {
		idByteList[i] = []byte(idStr)
	}
	return api.b.GetUserImgByIDs(idByteList)
}
