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

package service

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func NewObjectWithOplog(obj Object, oplog *BaseOplog) {
	obj.SetVersion(types.CurrentVersion)
	obj.SetID(oplog.ObjID)
	obj.SetCreateTS(oplog.UpdateTS)
	obj.SetUpdateTS(types.ZeroTimestamp)
	obj.SetStatus(types.StatusInternalSync)
	obj.SetCreatorID(oplog.CreatorID)
	obj.SetUpdaterID(oplog.CreatorID)
	obj.SetLogID(oplog.ID)
	obj.SetEntityID(oplog.dbPrefixID)
}

func SetNewObjectWithOplog(
	obj Object,
	oplog *BaseOplog,
) {
	obj.SetCreateTS(oplog.UpdateTS)
	obj.SetLogID(oplog.ID)
	obj.SetStatus(oplog.ToStatus())
	obj.SetUpdateTS(oplog.UpdateTS)
	obj.SetUpdaterID(oplog.CreatorID)
}

func SetUpdateObjectWithOplog(
	obj Object,
	oplog *BaseOplog,
) {
	obj.SetStatus(types.StatusAlive)

	obj.SetUpdateLogID(oplog.ID)
	obj.SetUpdateTS(oplog.UpdateTS)
	obj.SetUpdaterID(oplog.CreatorID)

}

func SetFailedObjectWithOplog(
	obj Object,
	oplog *BaseOplog,

	ts types.Timestamp,
) {

	obj.SetStatus(types.StatusFailed)

	obj.SetLogID(nil)
	obj.SetUpdateTS(ts)
	obj.SetSyncInfo(nil)
}

func SetDeleteObjectWithOplog(
	obj Object,
	status types.Status,
	oplog *BaseOplog,
) {
	obj.SetStatus(status)

	obj.SetLogID(oplog.ID)
	obj.SetUpdateTS(oplog.UpdateTS)
	obj.SetUpdaterID(oplog.CreatorID)
	obj.SetSyncInfo(nil)

	obj.RemoveMeta()

}

func AliveObjects(theObjs []Object) []Object {
	objs := make([]Object, 0, len(theObjs))
	for _, obj := range theObjs {
		if obj.GetStatus() == types.StatusAlive {
			objs = append(objs, obj)
		}
	}
	return objs
}

func GetObjList(obj Object, startID *types.PttID, limit int, listOrder pttdb.ListOrder, isLocked bool) ([]Object, error) {

	baseObj := obj.GetBaseObject()
	iter, err := baseObj.GetObjIterWithObj(startID, listOrder, isLocked)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	funcIter := pttdb.GetFuncIter(iter, listOrder)

	// for-loop
	var each Object
	var v []byte
	objs := make([]Object, 0)
	i := 0
	for funcIter() {
		if limit > 0 && i >= limit {
			break
		}

		k := iter.Key()
		log.Debug("GetObjList: (in-for-loop)", "i", i, "k", k)
		v = iter.Value()

		each = obj.NewEmptyObj()
		err = each.Unmarshal(v)
		if err != nil {
			continue
		}

		objs = append(objs, each)

		i++
	}

	return objs, nil

}

func (obj *BaseObject) GetObjIdxIterWithObj(startID *types.PttID, listOrder pttdb.ListOrder, isLocked bool) (iterator.Iterator, error) {

	prefix := obj.fullDBIdxPrefix

	if startID == nil {
		return obj.db.DB().NewIteratorWithPrefix(nil, prefix, listOrder)
	}

	o := obj.NewEmptyObj()
	o.SetID(startID)

	startKey, err := o.IdxKey()
	if err != nil {
		return nil, err
	}

	return obj.db.DB().NewIteratorWithPrefix(startKey, prefix, listOrder)
}

func (obj *BaseObject) GetCrossObjIterWithObj(startPrefix []byte, startID *types.PttID, listOrder pttdb.ListOrder, isLocked bool) (iterator.Iterator, error) {
	prefix := append(obj.fullDBPrefix, startPrefix...)

	if startID == nil {
		return obj.db.DB().NewIteratorWithPrefix(nil, prefix, listOrder)
	}

	o := obj.NewEmptyObj()
	startKey, err := o.GetKey(startID, isLocked)
	if err != nil {
		return nil, err
	}

	return obj.db.DB().NewIteratorWithPrefix(startKey, prefix, listOrder)

}

func (obj *BaseObject) GetObjIterWithObj(startID *types.PttID, listOrder pttdb.ListOrder, isLocked bool) (iterator.Iterator, error) {

	prefix := obj.fullDBPrefix

	log.Debug("GetObjIterWithObj: start", "prefix", prefix, "startID", startID)

	if startID == nil {
		return obj.db.DB().NewIteratorWithPrefix(nil, prefix, listOrder)
	}

	o := obj.NewEmptyObj()
	startKey, err := o.GetKey(startID, isLocked)
	if err != nil {
		return nil, err
	}

	return obj.db.DB().NewIteratorWithPrefix(startKey, prefix, listOrder)
}
