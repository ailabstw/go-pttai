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

func GetOplogList(oplog *BaseOplog, startID *types.PttID, limit int, listOrder pttdb.ListOrder, status types.Status, isLocked bool) ([]*BaseOplog, error) {

	iter, err := GetOplogIterWithOplog(oplog, startID, listOrder, status, isLocked)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	funcIter := pttdb.GetFuncIter(iter, listOrder)

	// for-loop
	var eachLog *BaseOplog
	oplogs := make([]*BaseOplog, 0)
	i := 0
	for funcIter() {
		if limit > 0 && i >= limit {
			break
		}

		v := iter.Value()

		eachLog = &BaseOplog{}
		err := eachLog.Unmarshal(v)
		if err != nil {
			continue
		}

		oplogs = append(oplogs, eachLog)

		i++
	}

	return oplogs, nil

}

func GetOplogIterWithOplog(oplog *BaseOplog, startID *types.PttID, listOrder pttdb.ListOrder, status types.Status, isLocked bool) (iterator.Iterator, error) {

	return getOplogIterCore(oplog.db, oplog.dbPrefix, oplog.dbIdxPrefix, oplog.dbMerklePrefix, oplog.dbPrefixID, startID, oplog.dbLock, isLocked, status, listOrder)

}

/*
func GetOplogIter(db *pttdb.LDBBatch, dbOplogPrefix []byte, dbOplogIdxPrefix []byte, dbOplogMerklePrefix []byte, prefixID *types.PttID, logID *types.PttID, dbLock *types.LockMap, isLocked bool, status types.Status, listOrder pttdb.ListOrder) (iterator.Iterator, error) {

	return getOplogIterCore(db, dbOplogPrefix, dbOplogIdxPrefix, dbOplogMerklePrefix, prefixID, logID, dbLock, isLocked, status, listOrder)
}
*/

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

	o := &BaseOplog{}
	o.SetDB(db, prefixID, dbOplogPrefix, dbOplogIdxPrefix, dbOplogMerklePrefix, dbLock)
	startKey, err := o.GetKey(logID, isLocked)
	if err != nil {
		return nil, err
	}

	return db.DB().NewIteratorWithPrefix(startKey, prefix, listOrder)
}

func getOplogsFromKeys(setDB func(oplog *BaseOplog), keys [][]byte) ([]*BaseOplog, error) {

	oplogs := make([]*BaseOplog, 0, len(keys))
	var oplog *BaseOplog
	for _, key := range keys {
		oplog = &BaseOplog{}
		setDB(oplog)
		err := oplog.Load(key)
		if err != nil {
			continue
		}

		oplogs = append(oplogs, oplog)
	}

	return oplogs, nil
}
