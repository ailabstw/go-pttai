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
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func DiffOplogKeys(myKeys [][]byte, theirKeys [][]byte) ([][]byte, [][]byte, error) {
	keyMap := make(map[string]bool)
	for _, key := range theirKeys {
		keyMap[string(key)] = true
	}

	myExtraKeys := make([][]byte, 0)
	for _, key := range myKeys {
		keyStr := string(key)
		if !keyMap[keyStr] {
			myExtraKeys = append(myExtraKeys, key)
		}
	}

	keyMap = make(map[string]bool)
	for _, key := range myKeys {
		keyMap[string(key)] = true
	}

	theirExtraKeys := make([][]byte, 0)
	for _, key := range theirKeys {
		if !keyMap[string(key)] {
			theirExtraKeys = append(theirExtraKeys, key)
		}
	}

	return myExtraKeys, theirExtraKeys, nil
}

func GetOplogIter(db *pttdb.LDBBatch, dbOplogPrefix []byte, dbOplogIdxPrefix []byte, dbOplogMerklePrefix []byte, prefixID *types.PttID, logID *types.PttID, dbLock *types.LockMap, isLocked bool, status types.Status, listOrder pttdb.ListOrder) (iterator.Iterator, error) {

	return getOplogIterCore(db, dbOplogPrefix, dbOplogIdxPrefix, dbOplogMerklePrefix, prefixID, logID, dbLock, isLocked, status, listOrder)
}

func getOplogIterCore(db *pttdb.LDBBatch, dbOplogPrefix []byte, dbOplogIdxPrefix []byte, dbOplogMerklePrefix []byte, prefixID *types.PttID, logID *types.PttID, dbLock *types.LockMap, isLocked bool, status types.Status, listOrder pttdb.ListOrder) (iterator.Iterator, error) {

	switch status {
	case types.StatusInternalPending:
		dbOplogPrefix = dbPrefixToDBPrefixInternal(dbOplogPrefix)
	case types.StatusPending:
		dbOplogPrefix = dbPrefixToDBPrefixMaster(dbOplogPrefix)
	}

	prefix, err := DBPrefix(dbOplogPrefix, prefixID)
	if err != nil {
		return nil, err
	}

	if logID == nil {
		return db.DB().NewIteratorWithPrefix(nil, prefix, listOrder)
	}

	o := &Oplog{}
	o.SetDB(db, prefixID, dbOplogPrefix, dbOplogIdxPrefix, dbOplogMerklePrefix, dbLock)
	startKey, err := o.GetKey(logID, isLocked)
	if err != nil {
		return nil, err
	}

	return db.DB().NewIteratorWithPrefix(startKey, prefix, listOrder)
}

func CheckPreLog(oplog *Oplog, prelog *Oplog, existIDs map[types.PttID]*Oplog) error {
	if oplog.PreLogID == nil {
		existIDs[*oplog.ID] = oplog
		return nil
	}

	log, ok := existIDs[*oplog.PreLogID]
	if ok && log.MasterLogID != nil {
		existIDs[*oplog.ID] = oplog
		return nil
	}

	err := prelog.Get(oplog.PreLogID, false)
	if err != nil {
		return ErrInvalidOplog
	}

	if prelog.MasterLogID == nil {
		return ErrInvalidOplog
	}

	existIDs[*oplog.ID] = oplog
	return nil

}
