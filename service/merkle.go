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
	"encoding/binary"
	"encoding/json"
	"sync"
	"time"

	pttcommon "github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"

	"github.com/ethereum/go-ethereum/common"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/*
Merkle is the representation / op of the merkle-tree-over-time for the oplog.
*/
type Merkle struct {
	DBOplogPrefix                []byte
	DBMerklePrefix               []byte
	dbMerkleMetaPrefix           []byte
	dbMerkleToUpdatePrefixWithID []byte
	dbMerkleUpdatingPrefixWithID []byte
	PrefixID                     *types.PttID
	db                           *pttdb.LDBBatch
	LastGenerateTS               types.Timestamp
	BusyGenerateTS               types.Timestamp
	LastSyncTS                   types.Timestamp
	LastFailSyncTS               types.Timestamp
	GenerateSeconds              time.Duration
	ExpireGenerateSeconds        int64

	lockIsBusyForceSync sync.RWMutex
	isBusyForceSync     bool

	forceSync chan struct{}

	lockToUpdateTS sync.Mutex
	toUpdateTS     map[int64]bool

	Name string
}

func NewMerkle(dbOplogPrefix []byte, dbMerklePrefix []byte, prefixID *types.PttID, db *pttdb.LDBBatch, name string) (*Merkle, error) {

	prefixIDBytes := prefixID[:]

	dbMerkleMetaPrefix := common.CopyBytes(dbMerklePrefix)
	copy(dbMerkleMetaPrefix[pttdb.OffsetDBKeyPrefixPostfix:], DBMerkleMetaPostfix)

	dbMerkleToUpdatePrefix := common.CopyBytes(dbMerklePrefix)
	copy(dbMerkleToUpdatePrefix[pttdb.OffsetDBKeyPrefixPostfix:], DBMerkleToUpdatePostfix)

	dbMerkleToUpdatePrefixWithID := append(dbMerkleToUpdatePrefix, prefixIDBytes...)

	dbMerkleUpdatingPrefix := common.CopyBytes(dbMerklePrefix)
	copy(dbMerkleUpdatingPrefix[pttdb.OffsetDBKeyPrefixPostfix:], DBMerkleUpdatingPostfix)

	dbMerkleUpdatingPrefixWithID := append(dbMerkleUpdatingPrefix, prefixIDBytes...)

	m := &Merkle{
		DBOplogPrefix:                dbOplogPrefix,
		DBMerklePrefix:               dbMerklePrefix,
		dbMerkleMetaPrefix:           dbMerkleMetaPrefix,
		dbMerkleToUpdatePrefixWithID: dbMerkleToUpdatePrefixWithID,
		dbMerkleUpdatingPrefixWithID: dbMerkleUpdatingPrefixWithID,

		PrefixID:              prefixID,
		db:                    db,
		GenerateSeconds:       GenerateOplogMerkleTreeSeconds,
		ExpireGenerateSeconds: ExpireGenerateOplogMerkleTreeSeconds,
		toUpdateTS:            make(map[int64]bool),

		forceSync: make(chan struct{}),

		Name: name,
	}

	lastGenerateTS, err := m.GetGenerateTime()
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}
	m.LastGenerateTS = lastGenerateTS

	lastSyncTS, err := m.GetSyncTime()
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}
	m.LastSyncTS = lastSyncTS

	lastFailSyncTS, err := m.GetFailSyncTime()
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}
	m.LastFailSyncTS = lastFailSyncTS

	return m, nil
}

func (m *Merkle) SaveMerkleTree(ts types.Timestamp) error {
	dbPrefix := m.DBPrefix()
	err := m.db.DB().TryLockMap(dbPrefix)
	if err != nil {
		return err
	}
	defer m.db.DB().UnlockMap(dbPrefix)

	// level1 - hr
	offsetTS, nextTS := ts.ToHRTimestamp()
	newestTS, err := m.SaveMerkleTreeCore(MerkleTreeLevelHR, offsetTS, nextTS, ts)
	if err != nil {
		return err
	}

	// level2 - day
	offsetTS, nextTS = ts.ToDayTimestamp()
	_, err = m.SaveMerkleTreeCore(MerkleTreeLevelDay, offsetTS, nextTS, ts)
	if err != nil {
		return err
	}

	// level3 - month
	offsetTS, nextTS = ts.ToMonthTimestamp()
	_, err = m.SaveMerkleTreeCore(MerkleTreeLevelMonth, offsetTS, nextTS, ts)
	if err != nil {
		return err
	}

	// level4 - year
	offsetTS, nextTS = ts.ToYearTimestamp()
	_, err = m.SaveMerkleTreeCore(MerkleTreeLevelYear, offsetTS, nextTS, ts)
	if err != nil {
		return err
	}

	err = m.SaveGenerateTime(newestTS)
	if err != nil {
		return err
	}

	return nil
}

func (m *Merkle) SaveMerkleTreeCore(level MerkleTreeLevel, offsetTS types.Timestamp, nextTS types.Timestamp, updateTS types.Timestamp) (types.Timestamp, error) {
	// 1. get iter
	childLevel := level - 1
	iter, err := m.GetMerkleIter(childLevel, offsetTS, nextTS, pttdb.ListOrderNext)
	if err != nil {
		return types.ZeroTimestamp, err
	}
	defer iter.Release()

	// 2. get addr-bytes
	addrBytes := make([]byte, 0, common.AddressLength*NMerkleTreeMagicAlloc)
	eachAddr := make([]byte, common.AddressLength)
	nChildren := uint32(0)
	tsBytes := make([]byte, types.SizeTimestamp)
	for iter.Next() {
		val := iter.Value()
		copy(eachAddr, val[MerkleTreeOffsetAddr:])
		copy(tsBytes, val[MerkleTreeOffsetTS:])
		addrBytes = append(addrBytes, eachAddr...)
		nChildren++
	}
	theAddr := types.Addr(addrBytes)

	// 3. marshal-key
	theKey, err := m.MarshalKey(level, offsetTS)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	if nChildren == 0 { // no children
		m.db.DB().Delete(theKey)
		return types.ZeroTimestamp, nil
	}

	// 4. marshal-node
	merkleNode := &MerkleNode{
		Level:     level,
		Addr:      theAddr,
		UpdateTS:  offsetTS,
		NChildren: nChildren,
	}

	theVal, err := merkleNode.Marshal()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	err = m.db.DB().Put(theKey, theVal)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	newestTS, err := types.UnmarshalTimestamp(tsBytes)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return newestTS, nil
}

func (m *Merkle) GetNodeByLevelTS(level MerkleTreeLevel, ts types.Timestamp) (*MerkleNode, error) {

	if level == MerkleTreeLevelNow {
		return nil, ErrInvalidMerkle
	}

	key, err := m.MarshalKey(level, ts)
	if err != nil {
		return nil, err
	}

	val, err := m.db.DB().Get(key)
	if err != nil {
		return nil, err
	}
	node := &MerkleNode{}
	err = node.Unmarshal(val)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (m *Merkle) SaveGenerateTime(ts types.Timestamp) error {
	key, err := m.MarshalGenerateTimeKey()
	if err != nil {
		return err
	}

	val := &pttdb.DBable{UpdateTS: ts}
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = m.db.DB().TryPut(key, marshaled, ts)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return err
	}

	m.LastGenerateTS = ts

	return nil
}

func (m *Merkle) ToGenerateTime() types.Timestamp {
	ts := m.LastGenerateTS
	if ts.Ts < OffsetMerkleSyncTime*3 {
		return types.ZeroTimestamp
	}

	ts.Ts -= OffsetMerkleSyncTime * 2
	return ts
}

func (m *Merkle) GetGenerateTime() (types.Timestamp, error) {
	key, err := m.MarshalGenerateTimeKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	val, err := m.db.DBGet(key)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	data := &pttdb.DBable{}
	err = json.Unmarshal(val, data)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return data.UpdateTS, nil
}

func (m *Merkle) SaveSyncTime(ts types.Timestamp) error {
	log.Debug("SaveSyncTime: start", "ts", ts, "prefixID", m.PrefixID)

	if ts.IsLess(m.LastSyncTS) {
		if m.LastSyncTS.Ts-ts.Ts < OffsetMerkleSyncTime {
			return nil
		}

		log.Error("SaveSyncTime: ts < m.LastSyncTS", "ts", ts, "lastSyncTS", m.LastSyncTS)
		return pttdb.ErrInvalidUpdateTS
	}

	key, err := m.MarshalSyncTimeKey()
	if err != nil {
		return err
	}

	val := &pttdb.DBable{UpdateTS: ts}
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = m.db.DB().TryPut(key, marshaled, ts)
	if err != nil {
		return err
	}

	log.Debug("SaveSyncTime: to set ts", "ts", ts, "prefixID", m.PrefixID)

	m.LastSyncTS = ts

	return nil
}

func (m *Merkle) ForceSaveSyncTime(ts types.Timestamp) error {
	log.Debug("ForceSaveSyncTime: start", "ts", ts, "prefixID", m.PrefixID)
	key, err := m.MarshalSyncTimeKey()
	if err != nil {
		return err
	}

	val := &pttdb.DBable{UpdateTS: ts}
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = m.db.DB().Put(key, marshaled)
	if err != nil {
		return err
	}

	log.Debug("SaveSyncTime: to set ts", "ts", ts, "prefixID", m.PrefixID)

	m.LastSyncTS = ts

	return nil
}

func (m *Merkle) ToSyncTime() (types.Timestamp, error) {
	ts, err := m.GetSyncTime()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	if ts.Ts < OffsetMerkleSyncTime*3 {
		return types.ZeroTimestamp, nil
	}

	ts.Ts -= OffsetMerkleSyncTime
	return ts, nil
}

func (m *Merkle) GetSyncTime() (types.Timestamp, error) {
	key, err := m.MarshalSyncTimeKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	val, err := m.db.DBGet(key)
	if err == leveldb.ErrNotFound {
		return types.ZeroTimestamp, nil
	}
	if err != nil {
		return types.ZeroTimestamp, err
	}

	data := &pttdb.DBable{}
	err = json.Unmarshal(val, data)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return data.UpdateTS, nil
}

func (m *Merkle) SaveFailSyncTime(ts types.Timestamp) error {
	key, err := m.MarshalFailSyncTimeKey()
	if err != nil {
		return err
	}

	val := &pttdb.DBable{UpdateTS: ts}
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = m.db.DB().TryPut(key, marshaled, ts)
	if err != nil {
		return err
	}

	m.LastFailSyncTS = ts

	return nil
}

func (m *Merkle) GetFailSyncTime() (types.Timestamp, error) {
	key, err := m.MarshalFailSyncTimeKey()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	val, err := m.db.DBGet(key)
	if err == leveldb.ErrNotFound {
		return types.ZeroTimestamp, nil
	}
	if err != nil {
		return types.ZeroTimestamp, err
	}

	data := &pttdb.DBable{}
	err = json.Unmarshal(val, data)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	return data.UpdateTS, nil
}

func (m *Merkle) MarshalGenerateTimeKey() ([]byte, error) {
	log.Debug("MarshalGenerateTimeKey: start", "m", m)
	return pttcommon.Concat([][]byte{m.dbMerkleMetaPrefix, DBMerkleGenerateTimePrefix, m.PrefixID[:]})
}

func (m *Merkle) MarshalSyncTimeKey() ([]byte, error) {
	log.Debug("MarshalSyncTimeKey: to concat", "m", m)
	return pttcommon.Concat([][]byte{m.dbMerkleMetaPrefix, DBMerkleSyncTimePrefix, m.PrefixID[:]})
}

func (m *Merkle) MarshalFailSyncTimeKey() ([]byte, error) {
	return pttcommon.Concat([][]byte{m.dbMerkleMetaPrefix, DBMerkleFailSyncTimePrefix, m.PrefixID[:]})
}

func (m *Merkle) DBPrefix() []byte {
	return append(m.DBMerklePrefix, m.PrefixID[:]...)
}

func (m *Merkle) MarshalKey(level MerkleTreeLevel, ts types.Timestamp) ([]byte, error) {
	tsBytes, err := ts.Marshal()
	if err != nil {
		return nil, err
	}

	return pttcommon.Concat([][]byte{m.DBMerklePrefix, m.PrefixID[:], []byte{uint8(level)}, tsBytes})
}

/*
Given the ts, retrieve the merkle until ts.
*/
func (m *Merkle) GetMerkleTreeList(ts types.Timestamp, isNow bool) ([]*MerkleNode, []*MerkleNode, error) {
	// year
	offsetYearTS, _ := ts.ToYearTimestamp()

	yearMerkleTreeList, err := m.GetMerkleTreeListCore(MerkleTreeLevelYear, types.ZeroTimestamp, offsetYearTS)
	//log.Debug("GetMerkleTreeList: after year", "offsetYearTS", offsetYearTS, "year", len(yearMerkleTreeList), "e", err)
	if err != nil {
		return nil, nil, err
	}

	// month
	offsetMonthTS, _ := ts.ToMonthTimestamp()

	monthMerkleTreeList, err := m.GetMerkleTreeListCore(MerkleTreeLevelMonth, offsetYearTS, offsetMonthTS)
	//log.Debug("GetMerkleTreeList: after month", "offsetYearTS", offsetYearTS, "offsetMonthTS", offsetMonthTS, "month", len(monthMerkleTreeList), "e", err)
	if err != nil {
		return nil, nil, err
	}

	// day
	offsetDayTS, _ := ts.ToDayTimestamp()

	dayMerkleTreeList, err := m.GetMerkleTreeListCore(MerkleTreeLevelDay, offsetMonthTS, offsetDayTS)
	//log.Debug("GetMerkleTreeList: after day", "offsetMonthTS", offsetMonthTS, "offsetDayTS", offsetDayTS, "day", len(dayMerkleTreeList), "e", err)
	if err != nil {
		return nil, nil, err
	}

	// hour
	offsetHourTS, _ := ts.ToHRTimestamp()

	hrMerkleTreeList, err := m.GetMerkleTreeListCore(MerkleTreeLevelHR, offsetDayTS, offsetHourTS)
	//log.Debug("GetMerkleTreeList: after hour", "offsetDayTS", offsetDayTS, "offsetHourTS", offsetHourTS, "hr", len(hrMerkleTreeList), "e", err)
	if err != nil {
		return nil, nil, err
	}

	lenList := len(yearMerkleTreeList) + len(monthMerkleTreeList) + len(dayMerkleTreeList) + len(hrMerkleTreeList)

	merkleTreeList := make([]*MerkleNode, 0, lenList)
	merkleTreeList = append(merkleTreeList, yearMerkleTreeList...)
	merkleTreeList = append(merkleTreeList, monthMerkleTreeList...)
	merkleTreeList = append(merkleTreeList, dayMerkleTreeList...)
	merkleTreeList = append(merkleTreeList, hrMerkleTreeList...)

	if !isNow {
		return merkleTreeList, nil, nil
	}

	//now
	nowMerkleTreeList, err := m.GetMerkleTreeListCore(MerkleTreeLevelNow, offsetHourTS, ts)
	if err != nil {
		return nil, nil, err
	}

	return merkleTreeList, nowMerkleTreeList, nil
}

func (m *Merkle) GetMerkleTreeListByLevel(level MerkleTreeLevel, ts types.Timestamp, nextTS types.Timestamp) ([]*MerkleNode, error) {
	return m.GetMerkleTreeListCore(level, ts, nextTS)
}

func (m *Merkle) GetMerkleTreeListCore(level MerkleTreeLevel, ts types.Timestamp, nextTS types.Timestamp) ([]*MerkleNode, error) {
	iter, err := m.GetMerkleIter(level, ts, nextTS, pttdb.ListOrderNext)
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	//log.Debug("GetMerkleTreeListCore: after GetMerkleIter", "level", level, "ts", ts, "nextTS", nextTS)

	results := make([]*MerkleNode, 0)
	for iter.Next() {
		val := iter.Value()
		eachMerkleNode := &MerkleNode{}
		err := eachMerkleNode.Unmarshal(val)
		//log.Debug("GetMerkleTreeListCore (in-loop): after Unmarshal", "val", val, "eachMerkleNode", eachMerkleNode, "e", err)
		if err != nil {
			continue
		}

		results = append(results, eachMerkleNode)
	}

	/*
		for i, result := range results {
			log.Debug("GetMerkleTreeListCore (after-loop)", "idx", fmt.Sprintf("(%d/%d)", i, len(results)), "result", result)
		}
	*/

	//log.Debug("GetMerkleTreeListCore: end", "level", level, "ts", ts, "nextTS", nextTS, "results", len(results))
	return results, nil
}

func (m *Merkle) GetNodeByKey(key []byte) (*MerkleNode, error) {
	val, err := m.db.DB().Get(key)
	if err != nil {
		return nil, err
	}
	node := &MerkleNode{}
	err = node.Unmarshal(val)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (m *Merkle) GetMerkleIter(level MerkleTreeLevel, ts types.Timestamp, nextTS types.Timestamp, listOrder pttdb.ListOrder) (iterator.Iterator, error) {
	startKey, err := m.MarshalKey(level, ts)
	if err != nil {
		return nil, err
	}

	endKey, err := m.MarshalKey(level, nextTS)
	if err != nil {
		return nil, err
	}

	r := &util.Range{Start: startKey, Limit: endKey}

	return m.db.DB().NewIteratorWithRange(r, listOrder), nil
}

func (m *Merkle) GetMerkleIterByKey(startKey []byte, level MerkleTreeLevel, listOrder pttdb.ListOrder) (iterator.Iterator, error) {
	prefix := append(m.DBPrefix(), byte(level))

	return m.db.DB().NewIteratorWithPrefix(startKey, prefix, listOrder)
}

func (m *Merkle) Clean() {
	var key []byte

	log.Debug("Clean: clean now", "prefixID", m.PrefixID)
	iter, err := m.GetMerkleIter(MerkleTreeLevelNow, types.ZeroTimestamp, types.MaxTimestamp, pttdb.ListOrderNext)
	if err != nil {
		return
	}
	defer iter.Release()

	db := m.db.DB()

	for iter.Next() {
		key = iter.Key()
		log.Debug("Clean: (now)", "key", key)
		db.Delete(key)
	}

	log.Debug("Clean: clean hr", "prefixID", m.PrefixID)
	iter, err = m.GetMerkleIter(MerkleTreeLevelHR, types.ZeroTimestamp, types.MaxTimestamp, pttdb.ListOrderNext)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		log.Debug("Clean: (hr)", "key", key)
		db.Delete(key)
	}

	log.Debug("Clean: clean day", "prefixID", m.PrefixID)

	iter, err = m.GetMerkleIter(MerkleTreeLevelDay, types.ZeroTimestamp, types.MaxTimestamp, pttdb.ListOrderNext)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		log.Debug("Clean: (day)", "key", key)
		db.Delete(key)
	}

	log.Debug("Clean: clean month", "prefixID", m.PrefixID)
	iter, err = m.GetMerkleIter(MerkleTreeLevelMonth, types.ZeroTimestamp, types.MaxTimestamp, pttdb.ListOrderNext)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		log.Debug("Clean: (month)", "key", key)
		db.Delete(key)
	}

	log.Debug("Clean: clean year", "prefixID", m.PrefixID)
	iter, err = m.GetMerkleIter(MerkleTreeLevelYear, types.ZeroTimestamp, types.MaxTimestamp, pttdb.ListOrderNext)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		log.Debug("Clean: (year)", "key", key)
		db.Delete(key)
	}

	log.Debug("Clean: clean generate-time key", "prefixID", m.PrefixID)
	key, err = m.MarshalGenerateTimeKey()
	err = db.Delete(key)
	log.Debug("Clean: (generate-time)", "key", key, "db", m.db, "e", err)

	log.Debug("Clean: clean sync-time key", "prefixID", m.PrefixID)
	key, err = m.MarshalSyncTimeKey()
	log.Debug("Clean: (sync-time)", "key", key)
	db.Delete(key)

	log.Debug("Clean: clean failed-sync-time key", "prefixID", m.PrefixID)
	key, err = m.MarshalFailSyncTimeKey()
	log.Debug("Clean: (fail-sync-time)", "key", key)
	db.Delete(key)
}

func (m *Merkle) ResetUpdateTS() error {
	m.lockToUpdateTS.Lock()
	defer m.lockToUpdateTS.Unlock()

	m.toUpdateTS = make(map[int64]bool)

	return nil
}

func (m *Merkle) SetUpdateTS(ts types.Timestamp) error {
	hrTS, _ := ts.ToHRTimestamp()

	m.lockToUpdateTS.Lock()
	defer m.lockToUpdateTS.Unlock()

	log.Debug("SetUpdateTS: to key", "hrTS", hrTS, "merkle", m.Name)

	key := m.MarshalToUpdateTSKey(hrTS)

	m.db.DB().Put(key, pttdb.ValueTrue)

	m.toUpdateTS[hrTS.Ts] = true

	return nil
}

func (m *Merkle) SetUpdateTS2(ts types.Timestamp, ts2 types.Timestamp) error {
	hrTS, _ := ts.ToHRTimestamp()
	hrTS2, _ := ts2.ToHRTimestamp()

	m.lockToUpdateTS.Lock()
	defer m.lockToUpdateTS.Unlock()

	m.toUpdateTS[hrTS.Ts] = true
	m.toUpdateTS[hrTS2.Ts] = true

	log.Debug("SetUpdateTS2: to key", "hrTS", hrTS, "hrTS2", hrTS2, "merkle", m.Name)

	key := m.MarshalToUpdateTSKey(hrTS)

	m.db.DB().Put(key, pttdb.ValueTrue)

	if hrTS.IsEqual(hrTS2) {
		return nil
	}

	key = m.MarshalToUpdateTSKey(hrTS2)

	m.db.DB().Put(key, pttdb.ValueTrue)

	return nil
}

func (m *Merkle) MarshalToUpdateTSKey(ts types.Timestamp) []byte {

	tsBytes := make([]byte, 8) // int64
	binary.BigEndian.PutUint64(tsBytes, uint64(ts.Ts))

	theBytes, _ := pttcommon.Concat([][]byte{m.dbMerkleToUpdatePrefixWithID, tsBytes})

	return theBytes
}

func (m *Merkle) GetAndResetToUpdateTSList() ([]int64, error) {
	m.lockToUpdateTS.Lock()
	defer m.lockToUpdateTS.Unlock()

	toUpdateTSs, err := m.getAndResetToUpdateTSs(true)
	if err != nil {
		return nil, err
	}

	toUpdateTSList := make([]int64, len(toUpdateTSs))
	pToUpdateTSList := toUpdateTSList
	for ts, _ := range toUpdateTSs {
		pToUpdateTSList[0] = ts
		pToUpdateTSList = pToUpdateTSList[1:]
	}

	toUpdateTSListBytes, err := json.Marshal(toUpdateTSList)
	if err != nil {
		return nil, err
	}

	key := m.MarshalUpdatingKey()
	m.db.DB().Put(key, toUpdateTSListBytes)

	return toUpdateTSList, nil
}

func (m *Merkle) MarshalUpdatingKey() []byte {
	return m.dbMerkleUpdatingPrefixWithID
}

func (m *Merkle) getAndResetToUpdateTSs(isLocked bool) (map[int64]bool, error) {
	if !isLocked {
		m.lockToUpdateTS.Lock()
		defer m.lockToUpdateTS.Unlock()
	}

	origToUpdateTS := m.toUpdateTS
	m.toUpdateTS = make(map[int64]bool)

	m.resetToUpdateTSDB()

	return origToUpdateTS, nil
}

func (m *Merkle) resetToUpdateTSDB() error {
	iter, err := m.db.DB().NewIteratorWithPrefix(nil, m.dbMerkleToUpdatePrefixWithID, pttdb.ListOrderNext)
	if err != nil {
		return err
	}
	defer iter.Release()

	var key []byte
	for iter.Next() {
		key = iter.Key()
		m.db.DB().Delete(key)
	}

	return nil
}

func (m *Merkle) LoadToUpdateTSs() error {
	m.lockToUpdateTS.Lock()
	defer m.lockToUpdateTS.Unlock()

	iter, err := m.db.DB().NewIteratorWithPrefix(nil, m.dbMerkleToUpdatePrefixWithID, pttdb.ListOrderNext)
	if err != nil {
		return err
	}
	defer iter.Release()

	var key []byte
	var ts int64
	for iter.Next() {
		key = iter.Key()
		ts = m.tsFromToUpdateTSKey(key)
		m.toUpdateTS[ts] = true
	}

	return nil
}

func (m *Merkle) tsFromToUpdateTSKey(key []byte) int64 {
	offset := pttdb.SizeDBKeyPrefix + types.SizePttID
	ts := binary.BigEndian.Uint64(key[offset:])

	return int64(ts)
}

func (m *Merkle) LoadUpdatingTSList() ([]int64, error) {
	key := m.MarshalUpdatingKey()

	val, err := m.db.DB().Get(key)
	if err != nil {
		return nil, err
	}

	var toUpdateTSList []int64
	err = json.Unmarshal(val, &toUpdateTSList)
	if err != nil {
		return nil, err
	}

	return toUpdateTSList, nil
}

func (m *Merkle) ResetUpdatingTSList() error {
	key := m.MarshalUpdatingKey()
	m.db.DB().Delete(key)

	return nil
}

func (m *Merkle) ForceSync() chan struct{} {
	return m.forceSync
}

func (m *Merkle) TryForceSync(pm ProtocolManager) error {
	err := m.trySetBusyForceSync()
	log.Debug("TryForceSync: after TrySetBusyForceSync", "e", err, "merkle", m.Name, "entity", pm.Entity().IDString())
	if err != nil {
		return err
	}

	m.forceSync <- struct{}{}

	m.resetBusyForceSync()

	log.Debug("TryForceSync: done", "merkle", m.Name, "entity", pm.Entity().GetID(), "service", pm.Entity().IDString())

	return nil
}

func (m *Merkle) trySetBusyForceSync() error {
	m.lockIsBusyForceSync.Lock()
	defer m.lockIsBusyForceSync.Unlock()

	if m.isBusyForceSync {
		return ErrBusy
	}

	m.isBusyForceSync = true

	return nil
}

func (m *Merkle) resetBusyForceSync() {
	m.lockIsBusyForceSync.Lock()
	defer m.lockIsBusyForceSync.Unlock()

	m.isBusyForceSync = false
}

func (m *Merkle) GetChildKeys(level MerkleTreeLevel, ts types.Timestamp) ([][]byte, error) {

	var startTS types.Timestamp
	var endTS types.Timestamp
	switch level {
	case MerkleTreeLevelHR:
		startTS, endTS = ts.ToHRTimestamp()
	case MerkleTreeLevelDay:
		startTS, endTS = ts.ToDayTimestamp()
	case MerkleTreeLevelMonth:
		startTS, endTS = ts.ToMonthTimestamp()
	case MerkleTreeLevelYear:
		startTS, endTS = ts.ToYearTimestamp()
	}

	iter, err := m.GetMerkleIter(level-1, startTS, endTS, pttdb.ListOrderNext)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	keys := make([][]byte, 0)
	var key []byte
	for iter.Next() {
		key = iter.Key()
		keys = append(keys, common.CopyBytes(key))
	}

	return keys, nil
}

func (m *Merkle) GetChildNodes(level MerkleTreeLevel, ts types.Timestamp) ([]*MerkleNode, error) {

	var startTS types.Timestamp
	var endTS types.Timestamp
	switch level {
	case MerkleTreeLevelHR:
		startTS, endTS = ts.ToHRTimestamp()
	case MerkleTreeLevelDay:
		startTS, endTS = ts.ToDayTimestamp()
	case MerkleTreeLevelMonth:
		startTS, endTS = ts.ToMonthTimestamp()
	case MerkleTreeLevelYear:
		startTS, endTS = ts.ToYearTimestamp()
	}

	iter, err := m.GetMerkleIter(level-1, startTS, endTS, pttdb.ListOrderNext)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	nodes := make([]*MerkleNode, 0, NMerkleTreeMagicAlloc)
	var val []byte
	for iter.Next() {
		val = iter.Value()
		node := &MerkleNode{}
		err = node.Unmarshal(val)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}
