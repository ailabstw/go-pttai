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
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (b *Backend) GetFriend(entityIDBytes []byte) (*BackendGetFriend, error) {
	theFriend, err := b.GetRawFriend(entityIDBytes)
	if err != nil {
		return nil, err
	}

	accountBackend := b.accountBackend
	userName, err := accountBackend.GetRawUserNameByID(theFriend.FriendID)
	if err != nil {
		userName = account.NewEmptyUserName()
	}

	return friendToBackendGetFriend(theFriend, userName), nil
}

func (b *Backend) GetRawFriend(entityIDBytes []byte) (*Friend, error) {

	entity, err := b.EntityIDToEntity(entityIDBytes)
	if err != nil {
		return nil, err
	}

	return entity.(*Friend), nil
}

func (b *Backend) GetFriendByFriendID(friendIDBytes []byte) (*BackendGetFriend, error) {
	friendID, err := types.UnmarshalTextPttID(friendIDBytes, false)
	if err != nil {
		return nil, err
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

func (b *Backend) DeleteFriend(entityIDBytes []byte) (bool, error) {
	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return false, err
	}
	pm := thePM.(*ProtocolManager)

	err = pm.DeleteFriend()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (b *Backend) GetFriendList(startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetFriend, error) {

	startID, err := types.UnmarshalTextPttID(startIDBytes, true)
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

func (b *Backend) GetFriendListByMsgCreateTS(tsBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetFriend, error) {

	ts := types.ZeroTimestamp
	var err error
	if len(tsBytes) != 0 {
		ts, err = types.UnmarshalTimestamp(tsBytes)
		if err != nil {
			return nil, err
		}
	}

	friendList, err := b.SPM().(*ServiceProtocolManager).GetFriendListByMsgCreateTS(ts, limit, listOrder)
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

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.GetFriendOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) GetPendingFriendOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.GetFriendOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) GetPendingFriendOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*FriendOplog, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.GetFriendOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) GetFriendOplogMerkleNodeList(entityIDBytes []byte, level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

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

func (b *Backend) CreateMessage(entityIDBytes []byte, message [][]byte, mediaIDStrs []string) (*BackendCreateMessage, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

	lenMediaIDs := len(mediaIDStrs)
	var mediaIDs []*types.PttID = nil
	var eachMediaID *types.PttID
	if len(mediaIDStrs) != 0 {
		mediaIDs = make([]*types.PttID, lenMediaIDs)
		for i, mediaIDStr := range mediaIDStrs {
			eachMediaID, err = types.UnmarshalTextPttID([]byte(mediaIDStr), false)
			if err != nil {
				return nil, err
			}
			mediaIDs[i] = eachMediaID
		}
	}

	theMessage, err := pm.CreateMessage(message, mediaIDs)
	log.Debug("CreateMessage: after CreateMessage", "e", err)
	if err != nil {
		return nil, err
	}

	return messageToBackendCreateMessage(theMessage), nil
}

func (b *Backend) GetMessageList(entityIDBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetMessage, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

	startID, err := types.UnmarshalTextPttID(startIDBytes, true)
	if err != nil {
		return nil, err
	}

	messageList, err := pm.GetMessageList(startID, limit, listOrder, true)

	backendMessageList := make([]*BackendGetMessage, len(messageList))
	for i, message := range messageList {
		backendMessageList[i] = messageToBackendGetMessage(message)
	}

	return backendMessageList, nil
}

func (b *Backend) GetMessageBlockList(entityIDBytes []byte, msgIDBytes []byte, limit uint32) ([]*BackendMessageBlock, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}
	pm := thePM.(*ProtocolManager)

	msgID, err := types.UnmarshalTextPttID(msgIDBytes, false)
	if err != nil {
		return nil, err
	}
	if msgID == nil {
		return nil, types.ErrInvalidID
	}

	msg, contentBlocks, err := pm.GetMessageBlockList(msgID, limit)
	if err != nil {
		return nil, err
	}

	blockInfo := msg.GetBlockInfo()
	if blockInfo == nil {
		return nil, pkgservice.ErrInvalidBlock
	}
	blockInfoID := blockInfo.ID

	backendMsgBlocks := make([]*BackendMessageBlock, len(contentBlocks))
	for i, contentBlock := range contentBlocks {
		backendMsgBlocks[i] = contentBlockToBackendMessageBlock(msg, blockInfoID, contentBlock)
	}

	return backendMsgBlocks, nil
}

func (b *Backend) MarkFriendSeen(entityIDBytes []byte) (types.Timestamp, error) {

	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return types.ZeroTimestamp, err
	}
	pm := thePM.(*ProtocolManager)

	return pm.SaveLastSeen(types.ZeroTimestamp)
}
