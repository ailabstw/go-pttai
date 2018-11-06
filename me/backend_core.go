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
	"github.com/syndtr/goleveldb/leveldb"
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

func (b *Backend) ShowMeURL() (*pkgservice.BackendJoinURL, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)
	myNodeID := b.myPtt.MyNodeID()

	keyInfo, err := pm.GetJoinKey()
	if err != nil {
		return nil, err
	}

	myUserName := &account.UserName{}
	err = myUserName.Get(myInfo.ID, true)
	if err == leveldb.ErrNotFound {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return pkgservice.MarshalBackendJoinURL(myInfo.ID, myNodeID, keyInfo, myUserName.Name, pkgservice.PathJoinMe)
}

/**********
 * Key
 **********/

func (b *Backend) ShowValidateKey() (*types.PttID, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	validateKey := myInfo.GetValidateKey()

	return validateKey, nil
}

func (b *Backend) ValidateValidateKey(keyBytes []byte) (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	validateKey := myInfo.GetValidateKey()
	theBytes, err := validateKey.MarshalText()
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(theBytes, keyBytes), nil
}

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

func (b *Backend) ShowURL() (*pkgservice.BackendJoinURL, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) JoinFriend(friendURL []byte) (*pkgservice.BackendJoinRequest, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) Get() (*BackendMyInfo, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	return MarshalBackendMyInfo(myInfo, b.myPtt), nil
}

func (b *Backend) GetRawMe() (*MyInfo, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	return myInfo, nil
}

func (b *Backend) GetRawMeByID(idBytes []byte) (*MyInfo, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(id)

	return entity.(*MyInfo), nil
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

func (b *Backend) GetMyNodes() ([]*MyNode, error) {
	pm := b.SPM().(*ServiceProtocolManager).MyInfo.PM().(*ProtocolManager)

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

func (b *Backend) GetFriendRequests() ([]*pkgservice.BackendJoinRequest, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) GetMeRequests() ([]*pkgservice.BackendJoinRequest, error) {
	pm := b.SPM().(*ServiceProtocolManager).MyInfo.PM().(*ProtocolManager)

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

func (b *Backend) CountPeers() (int, error) {
	pm := b.SPM().(*ServiceProtocolManager).MyInfo.PM().(*ProtocolManager)

	return pm.CountPeers()
}

func (b Backend) GetPeers() ([]*pkgservice.BackendPeer, error) {
	pm := b.SPM().(*ServiceProtocolManager).MyInfo.PM().(*ProtocolManager)

	peerList, err := pm.GetPeers()
	if err != nil {
		return nil, err
	}

	backendPeerList := make([]*pkgservice.BackendPeer, len(peerList))

	for i, peer := range peerList {
		backendPeerList[i] = pkgservice.PeerToBackendPeer(peer)
	}

	return backendPeerList, nil
}

func (b *Backend) GetMyBoard() (*content.BackendGetBoard, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) GetRaftStatus(idBytes []byte) (*RaftStatus, error) {
	var myInfo *MyInfo
	if len(idBytes) == 0 {
		myInfo = b.SPM().(*ServiceProtocolManager).MyInfo
	} else {
		myID := &types.PttID{}
		err := myID.UnmarshalText(idBytes)
		if err != nil {
			return nil, err
		}

		myInfo = b.SPM().Entity(myID).(*MyInfo)
	}

	if myInfo == nil {
		return nil, ErrInvalidMe
	}

	pm := myInfo.PM().(*ProtocolManager)
	return pm.GetRaftStatus()
}

func (b *Backend) GetTotalWeight() (uint32, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

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

func (b *Backend) BEGetJoinKeyInfos() ([]*pkgservice.KeyInfo, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	return myInfo.PM().JoinKeyInfos(), nil
}

func (b *Backend) BEGetOpKeyInfos() ([]*pkgservice.KeyInfo, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	return myInfo.PM().OpKeyInfoList(), nil
}

func (b *Backend) GetOpKeyInfosFromDB() ([]*pkgservice.KeyInfo, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	return myInfo.PM().GetOpKeyInfosFromDB()
}

/**********
 * MeOplog
 **********/

func (b *Backend) BEGetMeOplogList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetMeOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) BEGetPendingMeOplogMasterList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetMeOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) BEGetPendingMeOplogInternalList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetMeOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) BEGetMeOplogMerkleNodeList(level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

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

func (b *Backend) BEGetMasterOplogList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetMasterOplogList(logID, limit, listOrder, types.StatusAlive)
}

/**********
 * OpKeyOplog
 **********/

func (b *Backend) BEGetOpKeyOplogList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetOpKeyOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) BEGetPendingOpKeyOplogMasterList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetOpKeyOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) BEGetPendingOpKeyOplogInternalList(logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	var logID *types.PttID = nil
	if len(logIDBytes) != 0 {
		err := logID.Unmarshal(logIDBytes)
		if err != nil {
			return nil, err
		}
	}

	return pm.GetOpKeyOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) RevokeOpKey(keyIDBytes []byte, myKey []byte) (bool, error) {
	isValid, err := b.ValidateValidateKey(myKey)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, ErrInvalidMe
	}

	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo

	pm := myInfo.PM().(*ProtocolManager)

	keyID, err := types.UnmarshalTextPttID(keyIDBytes)
	if err != nil {
		return false, err
	}

	return pm.RevokeOpKeyInfo(keyID)
}

func (b *Backend) RequestRaftLead() (bool, error) {
	myInfo := b.SPM().(*ServiceProtocolManager).MyInfo
	pm := myInfo.PM().(*ProtocolManager)

	pm.ProposeRaftRequestLead()

	return false, nil
}
