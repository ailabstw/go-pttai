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
	"reflect"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (b *Backend) SetMyName(name []byte) (*account.UserName, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) SetMyNodeName(nodeIDBytes []byte, name []byte) (*MyNode, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) SetMyImage(imgStr string) (*account.UserImg, error) {
	return nil, types.ErrNotImplemented
}

/**********
 * Key
 **********/

func (b *Backend) ShowMyMasterKey() ([]byte, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	masterKey := myInfo.GetMasterKey()

	theBytes := crypto.FromECDSA(masterKey)

	return theBytes, nil
}

func (b *Backend) ValidateMyMasterKey(keyBytes []byte) (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	masterKey := myInfo.GetMasterKey()

	theBytes := crypto.FromECDSA(masterKey)

	return reflect.DeepEqual(keyBytes, theBytes), nil
}

func (b *Backend) ShowMyNodeKey() ([]byte, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	key := myInfo.GetNodeKey()

	theBytes := crypto.FromECDSA(key)

	return theBytes, nil

}

func (b *Backend) ValidateMyNodeKey(keyBytes []byte) (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	key := myInfo.GetNodeKey()

	theBytes := crypto.FromECDSA(key)

	return reflect.DeepEqual(keyBytes, theBytes), nil
}

func (b *Backend) ShowMySignKey() (*pkgservice.KeyInfo, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	key := myInfo.SignKey()

	return key, nil
}

func (b *Backend) RefreshMySignKey() (*pkgservice.KeyInfo, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	err := myInfo.CreateSignKeyInfo()
	if err != nil {
		return nil, err
	}

	key := myInfo.SignKey()

	return key, nil
}

func (b *Backend) ShowMyNodeSignKey() (*pkgservice.KeyInfo, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	key := myInfo.NodeSignKey()

	return key, nil
}

func (b *Backend) RefreshMyNodeSignKey() (*pkgservice.KeyInfo, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	err := myInfo.CreateNodeSignKeyInfo()
	if err != nil {
		return nil, err
	}

	key := myInfo.NodeSignKey()

	return key, nil
}

/**********
 * Join Me
 **********/

func (b *Backend) ShowMeURL() (*pkgservice.BackendJoinURL, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)
	myID := myInfo.ID
	myNodeID := b.myPtt.MyNodeID()

	keyInfo, err := pm.GetJoinKey()
	if err != nil {
		return nil, err
	}

	accountBackend := b.accountBackend
	myUserName, err := accountBackend.GetRawUserNameByID(myID)
	if err != nil {
		myUserName = account.NewEmptyUserName()
	}

	return pkgservice.MarshalBackendJoinURL(myID, myNodeID, keyInfo, myUserName.Name, pkgservice.PathJoinMe)
}

func (b *Backend) JoinMe(meURL []byte, myKeyBytes []byte) (*pkgservice.BackendJoinRequest, error) {

	joinRequest, err := pkgservice.ParseBackendJoinURL(meURL, pkgservice.PathJoinMe)
	log.Debug("JoinMe: after parse", "joinRequest", joinRequest, "e", err)
	if err != nil {
		return nil, err
	}

	myNodeID := b.myPtt.MyNodeID
	log.Debug("JoinMe: after parse", "joinRequest", joinRequest, "myNodeID", myNodeID, "joinNodeID", joinRequest.NodeID)
	if reflect.DeepEqual(myNodeID, joinRequest.NodeID) {
		return nil, ErrInvalidNode
	}

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)
	err = pm.JoinMe(joinRequest, myKeyBytes)
	if err != nil {
		return nil, err
	}

	backendJoinRequest := pkgservice.JoinRequestToBackendJoinRequest(joinRequest)

	return backendJoinRequest, nil
}

func (b *Backend) GetMeRequests(entityIDBytes []byte) ([]*pkgservice.BackendJoinRequest, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	joinMeRequests, lockJoinMeRequest := pm.GetJoinMeRequests()

	lockJoinMeRequest.RLock()
	defer lockJoinMeRequest.RUnlock()

	lenRequests := len(joinMeRequests)
	results := make([]*pkgservice.BackendJoinRequest, lenRequests)

	i := 0
	for _, joinRequest := range joinMeRequests {
		results[i] = pkgservice.JoinRequestToBackendJoinRequest(joinRequest)

		i++
	}

	return results, nil
}

func (b *Backend) GetJoinKeys(entityIDBytes []byte) ([]*pkgservice.KeyInfo, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.JoinKeyList(), nil
}

/**********
 * Friend
 **********/

func (b *Backend) ShowURL() (*pkgservice.BackendJoinURL, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)
	myNodeID := b.myPtt.MyNodeID()
	myID := myInfo.ID

	keyInfo, err := pm.GetJoinFriendKey()
	if err != nil {
		return nil, err
	}

	accountBackend := b.accountBackend
	myUserName, err := accountBackend.GetRawUserNameByID(myID)
	if err != nil {
		myUserName = account.NewEmptyUserName()
	}

	return pkgservice.MarshalBackendJoinURL(myID, myNodeID, keyInfo, myUserName.Name, pkgservice.PathJoinFriend)
}

func (b *Backend) JoinFriend(friendURL []byte) (*pkgservice.BackendJoinRequest, error) {
	joinRequest, err := pkgservice.ParseBackendJoinURL(friendURL, pkgservice.PathJoinFriend)
	if err != nil {
		return nil, err
	}

	myNodeID := b.myPtt.MyNodeID
	if reflect.DeepEqual(myNodeID, joinRequest.NodeID) {
		return nil, ErrInvalidNode
	}

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)
	err = pm.JoinFriend(joinRequest)
	if err != nil {
		return nil, err
	}

	backendJoinRequest := pkgservice.JoinRequestToBackendJoinRequest(joinRequest)

	return backendJoinRequest, nil
}

func (b *Backend) GetFriendRequests(entityIDBytes []byte) ([]*pkgservice.BackendJoinRequest, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	joinFriendRequests, err := pm.GetFriendRequests()
	if err != nil {
		return nil, err
	}

	theList := make([]*pkgservice.BackendJoinRequest, len(joinFriendRequests))
	for i, request := range joinFriendRequests {
		theList[i] = pkgservice.JoinRequestToBackendJoinRequest(request)
	}
	return theList, nil
}

/**********
 * MyInfo
 **********/

func (b *Backend) Get() (*BackendMyInfo, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	return MarshalBackendMyInfo(myInfo, b.myPtt), nil
}

func (b *Backend) GetRawMe(entityIDBytes []byte) (*MyInfo, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	return entity.(*MyInfo), nil
}

func (b *Backend) GetMyIDStr() (string, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	myIDBytes, err := myInfo.ID.MarshalText()
	if err != nil {
		return "", err
	}
	return string(myIDBytes), nil
}

func (b *Backend) GetMeList() ([]*BackendMyInfo, error) {
	entities := b.SPM().Entities()

	myInfoList := make([]*BackendMyInfo, 0, len(entities))
	var myInfo *BackendMyInfo
	for _, entity := range entities {
		myInfo = MarshalBackendMyInfo(entity.(*MyInfo), b.myPtt)
		myInfoList = append(myInfoList, myInfo)
	}

	return myInfoList, nil
}

func (b *Backend) Revoke(myKey []byte) (bool, error) {
	isValid, err := b.ValidateValidateKey(myKey)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, ErrInvalidMe
	}

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	pm := myInfo.PM().(*ProtocolManager)

	err = pm.DeleteMe()
	if err != nil {
		return false, err
	}

	return true, nil
}

/**********
 * My Nodes
 **********/

func (b *Backend) GetMyNodes(entityIDBytes []byte) ([]*MyNode, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	pm.RLockMyNodes()
	defer pm.RUnlockMyNodes()

	myNodeList := make([]*MyNode, len(pm.MyNodes))

	i := 0
	for _, node := range pm.MyNodes {
		myNodeList[i] = node

		i++
	}

	return myNodeList, nil
}

/**********
 * Raft and Node
 **********/

func (b *Backend) RequestRaftLead() (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	pm.ProposeRaftRequestLead()

	return false, nil
}

func (b *Backend) GetRaftStatus(entityIDBytes []byte) (*RaftStatus, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetRaftStatus()
}

func (b *Backend) GetTotalWeight(entityIDBytes []byte) (uint32, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return 0, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return 0, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.totalWeight, nil
}

func (b *Backend) RemoveNode(nodeIDStr string) (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	nodeID, err := discover.HexID(nodeIDStr)
	if err != nil {
		return false, err
	}

	pm := myInfo.PM().(*ProtocolManager)
	err = pm.RevokeNode(&nodeID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Backend) ForceRemoveNode(nodeIDStr string) (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	nodeID, err := discover.HexID(nodeIDStr)
	if err != nil {
		return false, err
	}

	pm := myInfo.PM().(*ProtocolManager)
	err = pm.ForceProposeRaftRemoveNode(&nodeID)
	return false, err
}

/**********
 * MeOplog
 **********/

func (b *Backend) GetMeOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetMeOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) GetPendingMeOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetMeOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) GetPendingMeOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetMeOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) GetMeOplogMerkleNodeList(entityIDBytes []byte, level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	merkleNodeList, err := pm.GetMeOplogMerkleNodeList(level, startKey, limit, listOrder)
	if err != nil {
		return nil, err
	}

	results := make([]*pkgservice.BackendMerkleNode, len(merkleNodeList))
	for i, eachMerkleNode := range merkleNodeList {
		results[i] = pkgservice.MerkleNodeToBackendMerkleNode(eachMerkleNode)
	}

	return results, nil
}

/**********
 * MasterOplog
 **********/

func (b *Backend) GetMyMasterOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetMyMasterOplogList(logID, limit, listOrder, types.StatusAlive)
}

/**********
 * Board
 **********/

func (b *Backend) GetBoard(entityIDBytes []byte) (*content.BackendGetBoard, error) {
	return nil, types.ErrNotImplemented
}

/**********
 * Profile
 **********/

func (b *Backend) GetMyProfile(entityIDBytes []byte) (*account.Profile, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	return entity.(*MyInfo).Profile, nil

}
