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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type UserNode struct {
	*pkgservice.BaseObject `json:"b"`
	UpdateTS               types.Timestamp `json:"UT"`

	UserID *types.PttID     `json:"UID"`
	NodeID *discover.NodeID `json:"NID"`

	SyncInfo *pkgservice.BaseSyncInfo `json:"s"`

	fullDBIdx2Prefix []byte
}

func NewUserNode(
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	db *pttdb.LDBBatch,
	dbLock *types.LockMap,
	fullDBPrefix []byte,
	fullDBIdxPrefix []byte,

	userID *types.PttID,
	nodeID *discover.NodeID,

	fullDBIdx2Prefix []byte,
) (*UserNode, error) {

	key, err := nodeID.Pubkey()
	if err != nil {
		return nil, err
	}

	id, err := types.NewPttIDWithPubkeyAndRefID(key, userID)
	if err != nil {
		return nil, err
	}

	o := pkgservice.NewObject(id, createTS, creatorID, entityID, logID, status, db, dbLock, fullDBPrefix, fullDBIdxPrefix)

	return &UserNode{
		BaseObject: o,
		UpdateTS:   createTS,

		UserID: userID,
		NodeID: nodeID,

		fullDBIdx2Prefix: fullDBIdx2Prefix,
	}, nil
}

func NewEmptyUserNode() *UserNode {
	return &UserNode{BaseObject: &pkgservice.BaseObject{}}
}

func UserNodesToObjs(typedObjs []*UserNode) []pkgservice.Object {
	objs := make([]pkgservice.Object, len(typedObjs))
	for i, obj := range typedObjs {
		objs[i] = obj
	}
	return objs
}

func ObjsToUserNodes(objs []pkgservice.Object) []*UserNode {
	typedObjs := make([]*UserNode, len(objs))
	for i, obj := range objs {
		typedObjs[i] = obj.(*UserNode)
	}
	return typedObjs
}

func AliveUserNodes(typedObjs []*UserNode) []*UserNode {
	objs := make([]*UserNode, 0, len(typedObjs))
	for _, obj := range typedObjs {
		if obj.Status == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (pm *ProtocolManager) SetUserNodeDB(u *UserNode) {
	u.SetDB(pm.DB(), pm.DBObjLock(), pm.Entity().GetID(), pm.dbUserNodePrefix, pm.dbUserNodeIdxPrefix)
	u.fullDBIdx2Prefix = pm.dbUserNodeIdx2Prefix
}

func (u *UserNode) Save(isLocked bool) error {
	var err error

	if !isLocked {
		err = u.Lock()
		if err != nil {
			return err
		}
		defer u.Unlock()
	}

	key, err := u.MarshalKey()
	if err != nil {
		return err
	}
	marshaled, err := u.Marshal()
	if err != nil {
		return err
	}

	idxKey, err := u.IdxKey()
	if err != nil {
		return err
	}

	idx2Key, err := u.Idx2Key()
	if err != nil {
		return err
	}

	idx := &pttdb.Index{Keys: [][]byte{key, idx2Key}, UpdateTS: u.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{K: key, V: marshaled},
		&pttdb.KeyVal{K: idx2Key, V: key},
	}

	log.Debug("UserNode.Save: to TryPut", "entityID", u.EntityID, "idxKey", idxKey, "key", key)

	_, err = u.DB().TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserNode) NewEmptyObj() pkgservice.Object {
	newU := NewEmptyUserNode()
	newU.CloneDB(u.BaseObject)
	newU.fullDBIdx2Prefix = u.fullDBIdx2Prefix
	return newU
}

func (u *UserNode) GetNewObjByID(id *types.PttID, isLocked bool) (pkgservice.Object, error) {
	newU := u.NewEmptyObj()
	err := newU.GetByID(isLocked)
	if err != nil {
		return nil, err
	}
	return newU, nil
}

func (u *UserNode) SetUpdateTS(ts types.Timestamp) {
	u.UpdateTS = ts
}

func (u *UserNode) GetUpdateTS() types.Timestamp {
	return u.UpdateTS
}

func (u *UserNode) GetByID(isLocked bool) error {
	var err error

	val, err := u.GetValueByID(isLocked)
	if err != nil {
		return err
	}

	return u.Unmarshal(val)
}

func (u *UserNode) MarshalKey() ([]byte, error) {
	return common.Concat([][]byte{u.FullDBPrefix(), u.UserID[:], u.ID[:]})
}

func (u *UserNode) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserNode) Unmarshal(theBytes []byte) error {
	return json.Unmarshal(theBytes, u)
}

func (u *UserNode) GetSyncInfo() pkgservice.SyncInfo {
	if u.SyncInfo == nil {
		return nil
	}
	return u.SyncInfo
}

func (u *UserNode) SetSyncInfo(theSyncInfo pkgservice.SyncInfo) error {
	syncInfo, ok := theSyncInfo.(*pkgservice.BaseSyncInfo)
	if !ok {
		return pkgservice.ErrInvalidData
	}
	u.SyncInfo = syncInfo

	return nil
}

/**********
 * idx2key
 **********/

func (u *UserNode) Idx2Key() ([]byte, error) {
	key := append(u.fullDBIdx2Prefix, u.NodeID[:]...)
	log.Debug("UserNode.Idx2Key", "nodeID", u.NodeID, "key", key)
	return key, nil
}

func (u *UserNode) GetIDByNodeID(nodeID *discover.NodeID) (*types.PttID, error) {
	u.NodeID = nodeID
	idx2Key, err := u.Idx2Key()
	if err != nil {
		return nil, err
	}
	key, err := u.DB().GetKeyBy2ndIdxKey(idx2Key)
	if err != nil {
		return nil, err
	}

	id := &types.PttID{}
	copy(id[:], key[len(u.FullDBPrefix())+types.SizePttID:])

	return id, nil
}
