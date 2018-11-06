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

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pm *BaseProtocolManager) SetNewestMasterLogID(id *types.PttID) error {
	pm.newestMasterLogID = id
	return pm.saveNewestMasterLogID()
}

func (pm *BaseProtocolManager) GetNewestMasterLogID() *types.PttID {
	return pm.newestMasterLogID
}

func (pm *BaseProtocolManager) saveNewestMasterLogID() error {
	e := pm.Entity()
	entityID := e.GetID()

	key, err := DBPrefix(DBNewestMasterLogIDPrefix, entityID)
	if err != nil {
		return err
	}

	err = pm.DB().DB().Put(key, pm.newestMasterLogID[:])
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) loadNewestMasterLogID() (*types.PttID, error) {
	e := pm.Entity()

	entityID := e.GetID()

	key, err := DBPrefix(DBNewestMasterLogIDPrefix, entityID)
	if err != nil {
		return nil, err
	}

	val, err := pm.db.DBGet(key)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	id := &types.PttID{}
	copy(id[:], val)

	return id, nil
}

func (pm *BaseProtocolManager) MasterLog0Hash() []byte {
	return pm.masterLog0Hash
}

func (pm *BaseProtocolManager) SetMasterLog0Hash(theBytes []byte) error {
	if pm.masterLog0Hash != nil {
		return nil
	}
	pm.masterLog0Hash = theBytes
	return pm.saveMasterLog0Hash()
}

func (pm *BaseProtocolManager) saveMasterLog0Hash() error {
	e := pm.Entity()
	entityID := e.GetID()

	key, err := DBPrefix(DBMasterLog0HashPrefix, entityID)
	if err != nil {
		return err
	}

	err = pm.DB().DB().Put(key, pm.masterLog0Hash)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) loadMasterLog0Hash() ([]byte, error) {
	e := pm.Entity()

	entityID := e.GetID()

	key, err := DBPrefix(DBMasterLog0HashPrefix, entityID)
	if err != nil {
		return nil, err
	}

	val, err := pm.db.DBGet(key)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (pm *BaseProtocolManager) IsMaster(id *types.PttID, isLocked bool) bool {
	return pm.isMaster(id, isLocked)
}

func (pm *BaseProtocolManager) defaultIsMaster(id *types.PttID, isLocked bool) bool {
	master, err := pm.GetMaster(id, isLocked)
	if err != nil {
		return false
	}
	return master.Status == types.StatusAlive
}

func (pm *BaseProtocolManager) GetMaster(id *types.PttID, isLocked bool) (*Master, error) {
	if !isLocked {
		pm.lockMaster.RLock()
		defer pm.lockMaster.RUnlock()
	}

	master, ok := pm.masters[*id]
	if !ok {
		return nil, types.ErrInvalidID
	}

	return master, nil
}

func (pm *BaseProtocolManager) RegisterMaster(master *Master, isLocked bool, isSkipPtt bool) error {
	if !isLocked {
		pm.lockMaster.Lock()
		defer pm.lockMaster.Unlock()
	}

	log.Debug("RegisterMaster", "masterID", master.ID, "myID", pm.Ptt().GetMyEntity().GetID(), "status", master.Status, "entity", pm.Entity().GetID())

	if master.Status != types.StatusAlive {
		return nil
	}

	pm.masters[*master.ID] = master

	if isSkipPtt {
		return nil
	}

	return pm.Ptt().RegisterEntityPeerWithOtherUserID(pm.Entity(), master.ID, PeerTypeImportant, false)
}

func (pm *BaseProtocolManager) UnregisterMaster(master *Master, isLocked bool) error {
	if !isLocked {
		pm.lockMaster.Lock()
		defer pm.lockMaster.Unlock()
	}

	delete(pm.masters, *master.ID)

	myID := pm.Ptt().GetMyEntity().GetID()

	log.Debug("UnregisterMaster", "masterID", master.ID, "myID", myID)

	if reflect.DeepEqual(myID, master.ID) {
		return nil
	}

	return pm.UnregisterPeerByOtherUserID(master.ID, false, false)
}

func (pm *BaseProtocolManager) loadMasters() ([]*Master, error) {

	master := NewEmptyMaster()
	pm.SetMasterObjDB(master)

	iter, err := master.BaseObject.GetObjIterWithObj(nil, pttdb.ListOrderNext, false)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	toRemoveKeys := make([][]byte, 0)
	masters := make([]*Master, 0)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		master = NewEmptyMaster()
		err = master.Unmarshal(val)
		if err != nil {
			log.Warn("loadMasters: unable to unmarshal", "key", key, "e", err)
			toRemoveKeys = append(toRemoveKeys, common.CloneBytes(key))
			continue
		}

		pm.SetMasterObjDB(master)

		masters = append(masters, master)
	}

	// to remove
	master = NewEmptyMaster()
	pm.SetMasterObjDB(master)
	entity := pm.Entity()
	for _, key := range toRemoveKeys {
		err = master.DeleteKey(key)
		if err != nil {
			log.Error("loadMasters: unable to delete key", "entity", entity.Name(), "e", err)
		}
	}

	return masters, nil
}

func (pm *BaseProtocolManager) registerMasters(masters []*Master, isLocked bool) {
	if !isLocked {
		pm.lockMaster.Lock()
		defer pm.lockMaster.Unlock()
	}
	// to register
	for _, master := range masters {
		if master.Status != types.StatusAlive {
			continue
		}
		pm.RegisterMaster(master, true, true)
	}
}
