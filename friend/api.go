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
 * Op
 **********/

func (api *PrivateAPI) CreateMessage(entityID string, message [][]byte, mediaIDs []string) (*BackendCreateMessage, error) {
	return api.b.CreateMessage(
		[]byte(entityID),
		message,
		mediaIDs,
	)
}

func (api *PrivateAPI) DeleteFriend(entityID string) (bool, error) {
	return api.b.DeleteFriend([]byte(entityID))
}

func (api *PrivateAPI) MarkFriendSeen(entityID string) (types.Timestamp, error) {
	return api.b.MarkFriendSeen([]byte(entityID))
}

/**********
 * Get Friend
 **********/

func (api *PrivateAPI) GetFriend(entityID string) (*BackendGetFriend, error) {
	return api.b.GetFriend([]byte(entityID))
}

func (api *PrivateAPI) GetRawFriend(entityID string) (*Friend, error) {
	return api.b.GetRawFriend([]byte(entityID))
}

func (api *PrivateAPI) GetFriendByFriendID(friendID string) (*BackendGetFriend, error) {
	return api.b.GetFriendByFriendID([]byte(friendID))
}

func (api *PrivateAPI) GetFriendList(startingFriendID string, limit int) ([]*BackendGetFriend, error) {
	return api.b.GetFriendList(
		[]byte(startingFriendID),
		limit,
		pttdb.ListOrderNext,
	)
}

func (api *PrivateAPI) GetFriendListByMsgCreateTS(tsBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetFriend, error) {
	return api.b.GetFriendListByMsgCreateTS(
		tsBytes,
		limit,
		listOrder,
	)
}

/**********
 * Get Message
 **********/

func (api *PrivateAPI) GetMessageList(entityID string, startingMessageID string, limit int, listOrder pttdb.ListOrder) ([]*BackendGetMessage, error) {
	return api.b.GetMessageList(
		[]byte(entityID),
		[]byte(startingMessageID),
		limit,
		listOrder,
	)
}

func (api *PrivateAPI) GetMessageBlockList(entityID string, messageID string, dummy0 string, dummy1 pkgservice.ContentType, dummy2 uint32, limit uint32) ([]*BackendMessageBlock, error) {
	return api.b.GetMessageBlockList([]byte(entityID), []byte(messageID), limit)
}

/**********
 * FriendOplog
 **********/

func (api *PrivateAPI) GetFriendOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {
	return api.b.GetFriendOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingFriendOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {
	return api.b.GetPendingFriendOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingFriendOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {
	return api.b.GetPendingFriendOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetFriendOplogMerkleNodeList(entityID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetFriendOplogMerkleNodeList([]byte(entityID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MasterOplog
 **********/

func (api *PrivateAPI) GetMasterOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetMasterOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetPendingMasterOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetPendingMasterOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMasterOplogMerkleNodeList(entityID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetMasterOplogMerkleNodeList([]byte(entityID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MemberOplog
 **********/

func (api *PrivateAPI) GetMemberOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetMemberOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetPendingMemberOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetPendingMemberOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMemberOplogMerkleNodeList(entityID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetMemberOplogMerkleNodeList([]byte(entityID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
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

func (api *PrivateAPI) CountPeers(entityID string) (int, error) {
	return api.b.CountPeers([]byte(entityID))
}

func (api *PrivateAPI) GetPeers(entityID string) ([]*pkgservice.BackendPeer, error) {
	return api.b.GetPeers([]byte(entityID))
}

func (api *PrivateAPI) ForceSync(entityID string) (bool, error) {
	return api.b.ForceSync([]byte(entityID))
}

func (api *PrivateAPI) ForceOpKey(entityID string) (bool, error) {
	return api.b.ForceOpKey([]byte(entityID))
}
