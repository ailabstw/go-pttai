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

package me

import (
	"encoding/binary"
	"math"
	"sync"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/ailabstw/go-pttai/raft"
	pb "github.com/ailabstw/go-pttai/raft/raftpb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type RaftStorage struct {
	lock sync.RWMutex

	lastIdx  uint64
	firstIdx uint64

	hardState pb.HardState
	snapshot  pb.Snapshot

	myID *types.PttID
}

func NewRaftStorage(isClean bool, myID *types.PttID) (*RaftStorage, error) {
	if isClean {
		return NewRaftStorageWithClean(myID)
	}

	rs := &RaftStorage{myID: myID}

	// firstIdx
	iter, err := rs.GetIter(0)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	isItered := false

	var idx uint64
	for iter.Next() {
		isItered = true
		key := iter.Key()
		idx, err = rs.GetIdxByKey(key)
		if err != nil {
			return nil, err
		}
		rs.firstIdx = idx + 1

		break
	}

	if !isItered {
		rs.SaveEntry(pb.Entry{}, false)
		rs.firstIdx = 1
		rs.lastIdx = 0
		rs.SetHardState(pb.HardState{})
		rs.ApplySnapshot(pb.Snapshot{})
		return rs, nil
	}

	log.Debug("NewRaftStorage", "firstIdx", rs.firstIdx)

	// lastIdx
	iter, err = rs.GetPrevIter(math.MaxUint64)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	isItered = false
	for iter.Prev() {
		isItered = true
		key := iter.Key()
		rs.lastIdx, err = rs.GetIdxByKey(key)
		if err != nil {
			return nil, err
		}
		break
	}

	log.Debug("NewRaftStorage: after iter lastIdx", "isItered", isItered, "lastIdx", rs.lastIdx)

	if !isItered {
		return nil, ErrInvalidEntry
	}

	// hard-state
	rs.hardState, err = rs.LoadHardState()
	log.Debug("NewRaftStorage: after load hard-state", "e", err, "hardState", rs.hardState)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	// snapshot
	rs.snapshot, err = rs.LoadSnapshot()
	log.Debug("NewRaftStorage: after load snapshot", "e", err)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return rs, nil
}

func NewRaftStorageWithClean(myID *types.PttID) (*RaftStorage, error) {
	err := CleanRaftStorage(myID, nil, true)
	if err != nil {
		return nil, err
	}

	rs := &RaftStorage{myID: myID}

	rs.SaveEntry(pb.Entry{}, false)
	rs.firstIdx = 1
	rs.lastIdx = 0
	rs.SetHardState(pb.HardState{})
	rs.ApplySnapshot(pb.Snapshot{})
	return rs, nil
}

func CleanRaftStorage(myID *types.PttID, rs *RaftStorage, isLocked bool) error {
	if rs == nil {
		rs = &RaftStorage{myID: myID}
	}

	if !isLocked {
		rs.Lock()
		defer rs.Unlock()
	}

	iter, err := rs.GetIter(0)
	if err != nil {
		return err
	}

	for iter.Next() {
		key := iter.Key()
		dbRaft.Delete(key)
	}

	return nil
}

func (rs *RaftStorage) InitialState() (pb.HardState, pb.ConfState, error) {

	return rs.hardState, rs.snapshot.Metadata.ConfState, nil
}

func (rs *RaftStorage) SetHardState(st pb.HardState) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	data, err := st.Marshal()
	if err != nil {
		return err
	}

	key, err := rs.MarshalKeyRaftHardState()
	if err != nil {
		return err
	}

	err = dbRaft.Put(key, data)
	if err != nil {
		return err
	}

	rs.hardState = st

	return nil
}

func (rs *RaftStorage) LoadHardState() (pb.HardState, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	hs := pb.HardState{}
	key, err := rs.MarshalKeyRaftHardState()
	if err != nil {
		return hs, err
	}

	data, err := dbRaft.Get(key)
	if err != nil {
		return hs, err
	}

	err = hs.Unmarshal(data)
	if err != nil {
		return hs, err
	}

	return hs, nil
}

func (rs *RaftStorage) Entries(startIdx uint64, endIdx uint64, maxSize uint64) ([]pb.Entry, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	if endIdx <= startIdx {
		return []pb.Entry{}, nil
	}

	iter, err := rs.GetIter(startIdx)
	if err != nil {
		return nil, err
	}

	ents := make([]pb.Entry, 0, endIdx-startIdx)

	i := startIdx
	for iter.Next() {
		if i == endIdx {
			break
		}

		val := iter.Value()
		ent := pb.Entry{}
		err := ent.Unmarshal(val)
		if err != nil {
			return nil, ErrInvalidEntry
		}

		if ent.Index != i {
			return nil, ErrInvalidEntry
		}

		ents = append(ents, ent)

		i++
	}

	if i != endIdx {
		return nil, ErrInvalidEntry
	}

	return ents, nil
}

func (rs *RaftStorage) Term(idx uint64) (uint64, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	key, err := rs.MarshalKey(idx)
	if err != nil {
		return 0, err
	}

	val, err := dbRaft.Get(key)
	if err != nil {
		return 0, err
	}

	ent := pb.Entry{}
	err = ent.Unmarshal(val)
	if err != nil {
		return 0, err
	}

	return ent.Term, nil
}

func (rs *RaftStorage) LastIndex() (uint64, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	return rs.lastIdx, nil
}

func (rs *RaftStorage) FirstIndex() (uint64, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	return rs.firstIdx, nil
}

func (rs *RaftStorage) Snapshot() (pb.Snapshot, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	return rs.snapshot, nil
}

func (rs *RaftStorage) ApplySnapshot(snap pb.Snapshot) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	rsIndex := rs.snapshot.Metadata.Index
	snapIndex := snap.Metadata.Index
	if rsIndex >= snapIndex {
		return raft.ErrSnapOutOfDate
	}

	data, err := snap.Marshal()
	if err != nil {
		return err
	}

	key, err := rs.MarshalKeyRaftSnapshot()
	if err != nil {
		return err
	}

	err = dbRaft.Put(key, data)
	if err != nil {
		return err
	}

	rs.snapshot = snap

	rs.firstIdx = snapIndex + 1
	rs.lastIdx = snapIndex

	rs.Compact(snapIndex, true)

	return nil
}

func (rs *RaftStorage) LoadSnapshot() (pb.Snapshot, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	snapshot := pb.Snapshot{}

	key, err := rs.MarshalKeyRaftSnapshot()
	if err != nil {
		return snapshot, err
	}

	data, err := dbRaft.Get(key)
	if err != nil {
		return snapshot, err
	}

	err = snapshot.Unmarshal(data)
	if err != nil {
		return snapshot, err
	}

	return snapshot, nil
}

func (rs *RaftStorage) CreateSnapshot(i uint64, cs *pb.ConfState, data []byte) (pb.Snapshot, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	if i <= rs.snapshot.Metadata.Index {
		return pb.Snapshot{}, raft.ErrSnapOutOfDate
	}

	if i > rs.lastIdx {
		return pb.Snapshot{}, ErrInvalidEntry
	}

	rs.snapshot.Metadata.Index = i
	term, err := rs.GetTermByIdx(i)
	if err != nil {
		return pb.Snapshot{}, err
	}
	rs.snapshot.Metadata.Term = term
	rs.snapshot.Data = data

	return rs.snapshot, nil
}

func (rs *RaftStorage) Compact(idx uint64, isLocked bool) error {
	if !isLocked {
		rs.lock.Lock()
		defer rs.lock.Unlock()
	}

	if idx == 0 {
		return nil
	}

	if idx < rs.firstIdx {
		return nil
	}

	if idx > rs.lastIdx {
		return nil
	}

	iter, err := rs.GetPrevIter(idx - 1)
	if err != nil {
		return err
	}

	toRemoveKeys := make([][]byte, 0, idx-rs.firstIdx+1)
	for iter.Prev() {
		key := common.CloneBytes(iter.Key())
		toRemoveKeys = append(toRemoveKeys, key)
	}

	for _, key := range toRemoveKeys {
		dbRaft.Delete(key)
	}

	return nil
}

func (rs *RaftStorage) Append(ents []pb.Entry) error {
	lenEnts := len(ents)
	if lenEnts == 0 {
		return nil
	}

	rs.lock.Lock()
	defer rs.lock.Unlock()

	rsFirstIdx := rs.firstIdx
	lastIdx := ents[lenEnts-1].Index
	if lastIdx < rs.firstIdx {
		return nil
	}

	if rsFirstIdx > ents[0].Index {
		ents = ents[rsFirstIdx-ents[0].Index:]
	}

	var err error
	for _, ent := range ents {
		err = rs.SaveEntry(ent, true)
		if err != nil {
			return err
		}
	}
	rs.lastIdx = ents[len(ents)-1].Index

	return nil

}

func (rs *RaftStorage) SaveEntry(ent pb.Entry, isLocked bool) error {
	if !isLocked {
		rs.lock.Lock()
		defer rs.lock.Unlock()
	}

	key, err := rs.MarshalKey(ent.Index)
	if err != nil {
		return err
	}

	data, err := ent.Marshal()
	if err != nil {
		return err
	}

	err = dbRaft.Put(key, data)
	if err != nil {
		return err
	}

	return nil

}

func (rs *RaftStorage) GetEntry(idx uint64) (pb.Entry, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	ent := pb.Entry{}
	key, err := rs.MarshalKey(idx)
	if err != nil {
		return ent, err
	}
	data, err := dbRaft.Get(key)
	if err != nil {
		return ent, err
	}
	err = ent.Unmarshal(data)
	if err != nil {
		return ent, err
	}

	return ent, nil
}

func (rs *RaftStorage) MarshalKey(idx uint64) ([]byte, error) {
	theBytes := make([]byte, pttdb.SizeDBKeyPrefix+types.SizePttID+8)
	copy(theBytes, DBRaftPrefix)
	copy(theBytes[pttdb.SizeDBKeyPrefix:], rs.myID[:])
	binary.BigEndian.PutUint64(theBytes[pttdb.SizeDBKeyPrefix+types.SizePttID:], idx)

	return theBytes, nil
}

func (rs *RaftStorage) MarshalKeyRaftHardState() ([]byte, error) {
	theBytes := make([]byte, pttdb.SizeDBKeyPrefix+types.SizePttID)
	copy(theBytes, DBKeyRaftHardState)
	copy(theBytes[pttdb.SizeDBKeyPrefix:], rs.myID[:])

	return theBytes, nil
}

func (rs *RaftStorage) MarshalKeyRaftSnapshot() ([]byte, error) {
	theBytes := make([]byte, pttdb.SizeDBKeyPrefix+types.SizePttID)
	copy(theBytes, DBKeyRaftSnapshot)
	copy(theBytes[pttdb.SizeDBKeyPrefix:], rs.myID[:])

	return theBytes, nil
}

func (rs *RaftStorage) GetIdxByKey(key []byte) (uint64, error) {
	idx := binary.BigEndian.Uint64(key[pttdb.SizeDBKeyPrefix+types.SizePttID:])
	return idx, nil
}

func (rs *RaftStorage) GetTermByIdx(idx uint64) (uint64, error) {
	entry, err := rs.GetEntry(idx)
	if err != nil {
		return 0, err
	}

	return entry.Term, nil
}

func (rs *RaftStorage) GetIter(idx uint64) (iterator.Iterator, error) {
	startKey, err := rs.MarshalKey(idx)
	endKey, err := rs.MarshalKey(math.MaxUint64)
	if err != nil {
		return nil, err
	}
	r := &util.Range{
		Start: startKey,
		Limit: endKey,
	}
	iter := dbRaft.NewIteratorWithRange(r, pttdb.ListOrderNext)
	return iter, nil
}

func (rs *RaftStorage) GetPrevIter(idx uint64) (iterator.Iterator, error) {
	startKey, err := rs.MarshalKey(0)
	endKey, err := rs.MarshalKey(idx)
	if err != nil {
		return nil, err
	}

	r := &util.Range{
		Start: startKey,
		Limit: endKey,
	}

	iter := dbRaft.NewIteratorWithRange(r, pttdb.ListOrderPrev)

	return iter, nil
}

func (rs *RaftStorage) Lock() {
	rs.lock.Lock()
}

func (rs *RaftStorage) Unlock() {
	rs.lock.Unlock()
}
