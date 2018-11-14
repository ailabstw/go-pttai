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
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (b *Backend) GetFriend(entityIDBytes []byte) (*BackendGetFriend, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	theFriend := b.SPM().Entity(entityID).(*Friend)
	if theFriend == nil {
		return nil, types.ErrInvalidID
	}

	accountBackend := b.accountBackend
	userName, err := accountBackend.GetRawUserNameByID(theFriend.FriendID)
	if err != nil {
		userName = account.NewEmptyUserName()
	}

	return friendToBackendGetFriend(theFriend, userName), nil
}

func (b *Backend) GetRawFriend(entityIDBytes []byte) (*Friend, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	f := b.SPM().Entity(entityID).(*Friend)
	if f == nil {
		return nil, types.ErrInvalidID
	}

	return f, nil
}

func (b *Backend) GetFriendByFriendID(friendIDBytes []byte) (*BackendGetFriend, error) {
	friendID, err := types.UnmarshalTextPttID(friendIDBytes)
	if err != nil {
		return nil, err
	}
	if friendID == nil {
		return nil, types.ErrInvalidID
	}

	theFriend, err := b.SPM().(*ServiceProtocolManager).GetFriendByFriendID(friendID)
	if err != nil {
		return nil, err
	}

	accountBackend := b.accountBackend
	userName, err := accountBackend.GetRawUserNameByID(friendID)
	if err != nil {
		return nil, err
	}

	return friendToBackendGetFriend(theFriend, userName), nil
}

func (b *Backend) RemoveFriend(idBytes []byte) (bool, error) {
	return false, types.ErrNotImplemented
}

func (b *Backend) GetFriendList(startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetFriend, error) {
	startID, err := types.UnmarshalTextPttID(startIDBytes)
	if err != nil {
		return nil, err
	}

	friendList, err := b.SPM().(*ServiceProtocolManager).GetFriendList(startID, limit, listOrder)
	if err != nil {
		return nil, err
	}

	accountBackend := b.accountBackend
	backendFriendList := make([]*BackendGetFriend, len(friendList))
	var userName *account.UserName
	for i, f := range friendList {
		userName, err = accountBackend.GetRawUserNameByID(f.FriendID)
		if err != nil {
			userName = account.NewEmptyUserName()
		}
		backendFriendList[i] = friendToBackendGetFriend(f, userName)
	}

	return backendFriendList, nil

}

func (b *Backend) GetFriendOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetFriendOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) GetPendingFriendOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetFriendOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) GetPendingFriendOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetFriendOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) GetFriendOplogMerkleNodeList(entityIDBytes []byte, level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	merkleNodeList, err := pm.GetFriendOplogMerkleNodeList(level, startKey, limit, listOrder)
	if err != nil {
		return nil, err
	}

	results := make([]*pkgservice.BackendMerkleNode, len(merkleNodeList))
	for i, eachMerkleNode := range merkleNodeList {
		results[i] = pkgservice.MerkleNodeToBackendMerkleNode(eachMerkleNode)
	}

	return results, nil
}

func (b *Backend) CreateArticle(friendIDBytes []byte, article [][]byte, mediaIDStrs []string) (*BackendCreateMessage, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) GetArticleList(boardIDBytes []byte, startingArticleIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetMessage, error) {

	return nil, types.ErrNotImplemented

}

func (b *Backend) GetArticleBlockList(boardIDBytes []byte, articleIDBytes []byte, subContentIDBytes []byte, contentType pkgservice.ContentType, blockID uint32, limit int) ([]*pkgservice.ArticleBlock, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) MarkFriendSeen(entityIDBytes []byte) (types.Timestamp, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return types.ZeroTimestamp, types.ErrInvalidID
	}
	if entityID == nil {
		return types.ZeroTimestamp, types.ErrInvalidID
	}

	f := b.SPM().Entity(entityID).(*Friend)
	if f == nil {
		return types.ZeroTimestamp, types.ErrInvalidID
	}

	pm := f.PM().(*ProtocolManager)

	return pm.SaveLastSeen(types.ZeroTimestamp)
}
