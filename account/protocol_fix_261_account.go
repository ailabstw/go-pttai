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
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pm *ProtocolManager) Fix261Account() error {
	isFixed, err := pm.isFix261Account()
	if err != nil {
		return err
	}
	if isFixed {
		return nil
	}

	err = pm.ResetAccount()
	if err != nil {
		return err
	}

	err = pm.setFix261Account()
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) ResetAccount() error {
	entity := pm.Entity().(*Profile)
	spm := entity.Service().SPM().(*ServiceProtocolManager)
	userID := entity.GetCreatorID()

	_, errImg := spm.GetUserImgByID(userID)

	_, errName := spm.GetUserNameByID(userID)

	if errImg == nil && errName == nil {
		return nil
	}

	// remove oplog
	oplog := &pkgservice.BaseOplog{}
	pm.SetUserDB(oplog)

	iter, err := pkgservice.GetOplogIterWithOplog(oplog, nil, pttdb.ListOrderNext, types.StatusAlive, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	db := oplog.GetDB().DB()

	var key []byte
	var val []byte
	for iter.Next() {
		key = iter.Key()
		val = iter.Value()
		err = oplog.Unmarshal(val)
		if err != nil {
			db.Delete(key)
			continue
		}
		if oplog.Op == UserOpTypeCreateProfile {
			continue
		}

		oplog.Delete(true)
	}

	// clean object
	pm.CleanObject()

	return nil
}

func (pm *ProtocolManager) isFix261Account() (bool, error) {
	key, err := pm.marshalFix261AccountKey()
	if err != nil {
		return false, err
	}
	_, err = dbMeta.Get(key)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pm *ProtocolManager) marshalFix261AccountKey() ([]byte, error) {
	entityID := pm.Entity().GetID()
	return common.Concat([][]byte{DBFix261Prefix, entityID[:]})
}

func (pm *ProtocolManager) setFix261Account() error {
	key, err := pm.marshalFix261AccountKey()
	if err != nil {
		return err
	}

	dbMeta.Put(key, pttdb.ValueTrue)

	return nil
}
