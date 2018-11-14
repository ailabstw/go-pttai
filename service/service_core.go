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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

/**********
 * MasterOplog
 **********/

func (svc *BaseService) GetMasterOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	return pm.GetMasterOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (svc *BaseService) GetPendingMasterOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	return pm.GetMasterOplogList(logID, limit, listOrder, types.StatusPending)
}

func (svc *BaseService) GetPendingMasterOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	return pm.GetMasterOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (svc *BaseService) GetMasterOplogMerkleNodeList(entityIDBytes []byte, level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendMerkleNode, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	merkleNodeList, err := pm.GetMasterOplogMerkleNodeList(level, startKey, limit, listOrder)
	if err != nil {
		return nil, err
	}

	results := make([]*BackendMerkleNode, len(merkleNodeList))
	for i, eachMerkleNode := range merkleNodeList {
		results[i] = MerkleNodeToBackendMerkleNode(eachMerkleNode)
	}

	return results, nil
}

/**********
 * Master List
 **********/

func (svc *BaseService) GetMasterListFromCache(entityIDBytes []byte) ([]*Master, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	return entity.PM().GetMasterListFromCache(false)
}

func (svc *BaseService) GetMasterList(entityIDBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*Master, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	startID, err := types.UnmarshalTextPttID(startIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	return entity.PM().GetMasterList(startID, limit, listOrder, false)
}

/**********
 * MemberOplog
 **********/

func (svc *BaseService) GetMemberOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MemberOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	return pm.GetMemberOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (svc *BaseService) GetPendingMemberOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MemberOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	return pm.GetMemberOplogList(logID, limit, listOrder, types.StatusPending)
}

func (svc *BaseService) GetPendingMemberOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*MemberOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	return pm.GetMemberOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (svc *BaseService) GetMemberOplogMerkleNodeList(entityIDBytes []byte, level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendMerkleNode, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	merkleNodeList, err := pm.GetMemberOplogMerkleNodeList(level, startKey, limit, listOrder)
	if err != nil {
		return nil, err
	}

	results := make([]*BackendMerkleNode, len(merkleNodeList))
	for i, eachMerkleNode := range merkleNodeList {
		results[i] = MerkleNodeToBackendMerkleNode(eachMerkleNode)
	}

	return results, nil
}

/**********
 * Member List
 **********/

func (svc *BaseService) GetMemberList(entityIDBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*Member, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	startID, err := types.UnmarshalTextPttID(startIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	return entity.PM().GetMemberList(startID, limit, listOrder, false)
}

/**********
 * OpKeyOplog
 **********/

func (svc *BaseService) GetOpKeyOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*OpKeyOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetOpKeyOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (svc *BaseService) GetPendingOpKeyOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*OpKeyOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetOpKeyOplogList(logID, limit, listOrder, types.StatusPending)
}

func (svc *BaseService) GetPendingOpKeyOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*OpKeyOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}
	return pm.GetOpKeyOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

/**********
 * Op
 **********/

func (svc *BaseService) ShowValidateKey() (*types.PttID, error) {
	myInfo := svc.Ptt().GetMyEntity()

	validateKey := myInfo.GetValidateKey()

	return validateKey, nil
}

func (svc *BaseService) ValidateValidateKey(keyBytes []byte) (bool, error) {
	myInfo := svc.Ptt().GetMyEntity()

	validateKey := myInfo.GetValidateKey()
	theBytes, err := validateKey.MarshalText()
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(theBytes, keyBytes), nil
}

func (svc *BaseService) RevokeOpKey(entityIDBytes []byte, keyIDBytes []byte, myKey []byte) (bool, error) {
	isValid, err := svc.ValidateValidateKey(myKey)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, ErrInvalidKey
	}

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return false, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return false, types.ErrInvalidID
	}

	pm := entity.PM()

	keyID, err := types.UnmarshalTextPttID(keyIDBytes)
	if err != nil {
		return false, err
	}

	return pm.RevokeOpKey(keyID)
}

func (svc *BaseService) GetOpKeys(entityIDBytes []byte) ([]*KeyInfo, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	pm := entity.PM()

	return pm.OpKeyList(), nil
}

func (svc *BaseService) GetOpKeysFromDB(entityIDBytes []byte) ([]*KeyInfo, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}

	pm := entity.PM()

	return pm.GetOpKeyListFromDB()
}

/**********
 * Peers
 **********/

func (svc *BaseService) CountPeers(entityIDBytes []byte) (int, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return 0, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return 0, types.ErrInvalidID
	}

	pm := entity.PM()

	return pm.CountPeers()
}

func (svc *BaseService) GetPeers(entityIDBytes []byte) ([]*BackendPeer, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := svc.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM()

	peerList, err := pm.GetPeers()
	if err != nil {
		return nil, err
	}

	backendPeerList := make([]*BackendPeer, len(peerList))

	for i, peer := range peerList {
		backendPeerList[i] = PeerToBackendPeer(peer)
	}

	return backendPeerList, nil
}
