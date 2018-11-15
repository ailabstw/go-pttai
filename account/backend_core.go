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

func (b *Backend) GetRawProfile(idBytes []byte) (*Profile, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	profile := b.SPM().Entity(id).(*Profile)

	return profile, nil
}

/**********
 * UserOplog
 **********/

func (b *Backend) GetUserOplogList(profileIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {

	profileID, err := types.UnmarshalTextPttID(profileIDBytes)
	if err != nil {
		return nil, err
	}
	if profileID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(profileID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetUserOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) GetPendingUserOplogMasterList(profileIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {

	profileID, err := types.UnmarshalTextPttID(profileIDBytes)
	if err != nil {
		return nil, err
	}
	if profileID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(profileID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetUserOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) GetPendingUserOplogInternalList(profileIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserOplog, error) {

	profileID, err := types.UnmarshalTextPttID(profileIDBytes)
	if err != nil {
		return nil, err
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(profileID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetUserOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) GetUserOplogMerkleNodeList(profileIDBytes []byte, level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	profileID, err := types.UnmarshalTextPttID(profileIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(profileID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	merkleNodeList, err := pm.GetUserOplogMerkleNodeList(level, startKey, limit, listOrder)
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
 * User Node
 **********/

func (b *Backend) GetUserNodeList(idBytes []byte, startIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*UserNode, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	var startID *types.PttID
	if len(startIDBytes) != 0 {
		startID, err = types.UnmarshalTextPttID(startIDBytes)
		if err != nil {
			return nil, err
		}
	}

	profile := b.SPM().Entity(id).(*Profile)
	if profile == nil {
		return nil, types.ErrInvalidID
	}

	return profile.PM().(*ProtocolManager).GetUserNodeList(startID, limit, listOrder, false)
}

func (b *Backend) GetUserNodeInfo(idBytes []byte) (*UserNodeInfo, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	profile := b.SPM().Entity(id).(*Profile)
	if profile == nil {
		return nil, types.ErrInvalidID
	}

	return profile.PM().(*ProtocolManager).GetUserNodeInfo(), nil
}

/**********
 * User Name
 **********/

func (b *Backend) GetRawUserName(idBytes []byte) (*UserName, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
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
	for _, idBytes := range idByteList {

		u, err := b.GetUserName(idBytes)
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
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	u := &UserImg{}
	err = u.Get(id, true)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (b *Backend) GetUserImg(idBytes []byte) (*BackendUserImg, error) {
	id, err := types.UnmarshalTextPttID(idBytes)
	if err != nil {
		return nil, err
	}

	u := &UserImg{}
	err = u.Get(id, true)
	if err != nil {
		return nil, err
	}

	return userImgToBackendUserImg(u), nil
}

func (b *Backend) GetUserImgByIDs(idByteList [][]byte) (map[string]*BackendUserImg, error) {
	backendUserImgs := make(map[string]*BackendUserImg)
	for _, idBytes := range idByteList {
		id, err := types.UnmarshalTextPttID(idBytes)
		if err != nil {
			continue
		}

		u := &UserImg{}
		err = u.Get(id, true)
		if err != nil {
			continue
		}

		backendUserImgs[string(idBytes)] = userImgToBackendUserImg(u)
	}

	return backendUserImgs, nil
}
