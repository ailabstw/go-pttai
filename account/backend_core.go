// Copyright 2019 The go-pttai Authors
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
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (b *Backend) GetRawProfile(entityIDBytes []byte) (*Profile, error) {

	entity, err := b.EntityIDToEntity(entityIDBytes)
	if err != nil {
		return nil, err
	}

	profile := entity.(*Profile)

	return profile, nil
}

/**********
 * UserOplog
 **********/

func (b *Backend) GetUserOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.(*ProtocolManager).GetUserOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) GetPendingUserOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.(*ProtocolManager).GetUserOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) GetPendingUserOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.(*ProtocolManager).GetUserOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) GetUserOplogMerkleNodeList(entityIDBytes []byte, level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}

	merkleNodeList, err := pm.(*ProtocolManager).GetUserOplogMerkleNodeList(level, startKey, limit, listOrder)
	if err != nil {
		return nil, err
	}

	results := make([]*pkgservice.BackendMerkleNode, len(merkleNodeList))
	for i, eachMerkleNode := range merkleNodeList {
		results[i] = pkgservice.MerkleNodeToBackendMerkleNode(eachMerkleNode)
	}

	return results, nil
}

func (b *Backend) ForceSyncUserMerkle(entityIDBytes []byte) (bool, error) {
	thePM, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return false, err
	}
	pm := thePM.(*ProtocolManager)

	return pm.ForceSyncUserMerkle()
}

/**********
 * User Node
 **********/

func (b *Backend) GetUserNodeList(entityIDBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserNode, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}

	startID, err := types.UnmarshalTextPttID(startIDBytes, true)
	if err != nil {
		return nil, err
	}

	return pm.(*ProtocolManager).GetUserNodeList(startID, limit, listOrder, false)
}

/*
func (b *Backend) GetUserNodeInfo(entityIDBytes []byte) (*UserNodeInfo, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.(*ProtocolManager).GetUserNodeInfo(), nil
}
*/

func (b *Backend) RemoveUserNode(entityIDBytes []byte, nodeIDBytes []byte) (types.Bool, error) {

	pm, err := b.EntityIDToPM(entityIDBytes)
	if err != nil {
		return false, err
	}

	nodeID, err := discover.BytesID(nodeIDBytes)
	if err != nil {
		return false, err
	}

	err = pm.(*ProtocolManager).RemoveUserNode(&nodeID)
	if err != nil {
		return false, err
	}

	return true, nil

}

/**********
 * User Name
 **********/

func (b *Backend) GetRawUserName(idBytes []byte) (*UserName, error) {

	id, err := types.UnmarshalTextPttID(idBytes, false)
	if err != nil {
		return nil, err
	}

	return b.GetRawUserNameByID(id)

}

func (b *Backend) GetRawUserNameByID(id *types.PttID) (*UserName, error) {

	spm := b.SPM().(*ServiceProtocolManager)
	return spm.GetUserNameByID(id)
}

func (b *Backend) GetUserName(idBytes []byte) (*BackendUserName, error) {

	u, err := b.GetRawUserName(idBytes)
	if err != nil {
		return nil, err
	}

	return userNameToBackendUserName(u), nil
}

func (b *Backend) GetUserNameByIDs(idByteList [][]byte) (map[string]*BackendUserName, error) {

	backendUserNames := make(map[string]*BackendUserName)

	var u *BackendUserName
	var err error

	for _, idBytes := range idByteList {

		u, err = b.GetUserName(idBytes)
		if err != nil {
			continue
		}

		backendUserNames[string(idBytes)] = u
	}

	return backendUserNames, nil
}

/**********
 * User Img
 **********/

func (b *Backend) GetRawUserImg(idBytes []byte) (*UserImg, error) {

	id, err := types.UnmarshalTextPttID(idBytes, false)
	if err != nil {
		return nil, err
	}

	return b.GetRawUserImgByID(id)

}

func (b *Backend) GetRawUserImgByID(id *types.PttID) (*UserImg, error) {

	spm := b.SPM().(*ServiceProtocolManager)
	return spm.GetUserImgByID(id)
}

func (b *Backend) GetUserImg(idBytes []byte) (*BackendUserImg, error) {

	u, err := b.GetRawUserImg(idBytes)
	if err != nil {
		return nil, err
	}

	return userImgToBackendUserImg(u), nil
}

func (b *Backend) GetUserImgByIDs(idByteList [][]byte) (map[string]*BackendUserImg, error) {
	backendUserImgs := make(map[string]*BackendUserImg)

	var u *BackendUserImg
	var err error
	for _, idBytes := range idByteList {

		u, err = b.GetUserImg(idBytes)
		if err != nil {
			continue
		}

		backendUserImgs[string(idBytes)] = u
	}

	return backendUserImgs, nil
}

/**********
 * Name Card
 **********/

func (b *Backend) GetRawNameCard(idBytes []byte) (*NameCard, error) {

	id, err := types.UnmarshalTextPttID(idBytes, false)
	if err != nil {
		return nil, err
	}

	return b.GetRawNameCardByID(id)

}

func (b *Backend) GetRawNameCardByID(id *types.PttID) (*NameCard, error) {

	spm := b.SPM().(*ServiceProtocolManager)
	return spm.GetNameCardByID(id)
}

func (b *Backend) GetNameCard(idBytes []byte) (*BackendNameCard, error) {

	u, err := b.GetRawNameCard(idBytes)
	if err != nil {
		return nil, err
	}

	return userNameToBackendNameCard(u), nil
}

func (b *Backend) GetNameCardByIDs(idByteList [][]byte) (map[string]*BackendNameCard, error) {

	backendNameCards := make(map[string]*BackendNameCard)

	var u *BackendNameCard
	var err error

	for _, idBytes := range idByteList {

		u, err = b.GetNameCard(idBytes)
		if err != nil {
			continue
		}

		backendNameCards[string(idBytes)] = u
	}

	return backendNameCards, nil
}
