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
 * MasterOplog
 **********/

func (api *PrivateAPI) GetMasterOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.BEGetMasterOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogMasterList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.BEGetPendingMasterOplogMasterList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.BEGetPendingMasterOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMasterOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.BEGetMasterOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MemberOplog
 **********/

func (api *PrivateAPI) GetMemberOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.BEGetMemberOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogMemberList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.BEGetPendingMemberOplogMemberList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.BEGetPendingMemberOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMemberOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.BEGetMemberOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * UserOplog
 **********/

func (api *PrivateAPI) GetUserOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {
	return api.b.BEGetUserOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingUserOplogMemberList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {
	return api.b.BEGetPendingUserOplogMemberList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingUserOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {
	return api.b.BEGetPendingUserOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetUserOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.BEGetUserOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
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

func (api *PrivateAPI) GetMasterList(idStr string) ([]*pkgservice.Master, error) {
	return api.b.GetMasterList([]byte(idStr))
}

func (api *PrivateAPI) GetMasterListFromDB(idStr string, startID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.Master, error) {
	return api.b.GetMasterListFromDB([]byte(idStr), []byte(startID), limit, listOrder)
}

func (api *PrivateAPI) GetMemberList(idStr string, startID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.Member, error) {
	return api.b.GetMemberList([]byte(idStr), []byte(startID), limit, listOrder)
}

type PublicAPI struct {
	b *Backend
}

func NewPublicAPI(b *Backend) *PublicAPI {
	return &PublicAPI{b}
}

func (api *PublicAPI) GetUserName(idStr string) (*BackendUserName, error) {
	return api.b.GetUserName([]byte(idStr))
}

func (api *PublicAPI) GetUserNameList(idStr string, limit int, listOrder pttdb.ListOrder) ([]*BackendUserName, error) {
	return api.b.GetUserNameList([]byte(idStr), limit, listOrder)
}

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
