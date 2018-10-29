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
	"github.com/ailabstw/go-pttai/common/types"
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

	err = pm.db.DB().Put(key, pm.newestMasterLogID[:])
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

func (pm *BaseProtocolManager) Master0Hash() []byte {
	return nil
}

func (pm *BaseProtocolManager) IsMaster(id *types.PttID) bool {
	return false
}
