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
	"encoding/json"
	"reflect"
	"sort"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/*
Merkle is the representation / op of the merkle-tree-over-time for the oplog.
*/
type Merkle struct {
	DBOplogPrefix         []byte
	DBMerklePrefix        []byte
	dbMerkleMetaPrefix    []byte
	PrefixID              *types.PttID
	db                    *pttdb.LDBBatch
	LastGenerateTS        types.Timestamp
	BusyGenerateTS        types.Timestamp
	LastSyncTS            types.Timestamp
	LastFailSyncTS        types.Timestamp
	GenerateSeconds       time.Duration
	ExpireGenerateSeconds int64
}

func NewMerkle(dbOplogPrefix []byte, dbMerklePrefix []byte, prefixID *types.PttID, db *pttdb.LDBBatch) (*Merkle, error) {
	dbMerkleMetaPrefix := common.CloneBytes(dbMerklePrefix)
	copy(dbMerkleMetaPrefix[pttdb.OffsetDBKeyPrefixPostfix:], pttdb.DBMetaPostfix)

	m := &Merkle{
		DBOplogPrefix:         dbOplogPrefix,
		DBMerklePrefix:        dbMerklePrefix,
		dbMerkleMetaPrefix:    dbMerkleMetaPrefix,
		PrefixID:              prefixID,
		db:                    db,
		GenerateSeconds:       GenerateOplogMerkleTreeSeconds,
		ExpireGenerateSeconds: ExpireGenerateOplogMerkleTreeSeconds,
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

	if newestTS.IsEqual(types.ZeroTimestamp) {
		return nil
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

func (m *Merkle) SaveMerkleTreeCore(level MerkleTreeLevel, ts types.Timestamp, nextTS types.Timestamp, updateTS types.Timestamp) (types.Timestamp, error) {
	// 1. get iter
	childLevel := level - 1
	iter, err := m.GetMerkleIter(childLevel, ts, nextTS, pttdb.ListOrderNext)
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

	if nChildren == 0 { // no children
		return types.ZeroTimestamp, nil
	}

	// 3. marshal-key
	theKey, err := m.MarshalKey(level, ts)
	if err != nil {
		return types.ZeroTimestamp, err
	}

	// 4. marshal-node
	merkleNode := &MerkleNode{
		Level:     level,
		Addr:      theAddr,
		UpdateTS:  updateTS,
		NChildren: nChildren,
	}

	theVal, err := merkleNode.Marshal()
	if err != nil {
		return types.ZeroTimestamp, err
	}

	//log.Debug("SaveMerkleTreeCore: to save", "K", theKey, "V", theVal, "merkleNode", merkleNode)
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

	m.LastSyncTS = ts

	return nil
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
	return common.Concat([][]byte{m.dbMerkleMetaPrefix, DBMerkleGenerateTimePrefix, m.PrefixID[:]})
}

func (m *Merkle) MarshalSyncTimeKey() ([]byte, error) {
	log.Debug("MarshalSyncTimeKey: to concat", "m", m)
	return common.Concat([][]byte{m.dbMerkleMetaPrefix, DBMerkleSyncTimePrefix, m.PrefixID[:]})
}

func (m *Merkle) MarshalFailSyncTimeKey() ([]byte, error) {
	return common.Concat([][]byte{m.dbMerkleMetaPrefix, DBMerkleFailSyncTimePrefix, m.PrefixID[:]})
}

func (m *Merkle) DBPrefix() []byte {
	return append(m.DBMerklePrefix, m.PrefixID[:]...)
}

func (m *Merkle) MarshalKey(level MerkleTreeLevel, ts types.Timestamp) ([]byte, error) {
	tsBytes, err := ts.Marshal()
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{m.DBMerklePrefix, m.PrefixID[:], []byte{uint8(level)}, tsBytes})
}

/*
Given the ts, retrieve the merkle until ts.
*/
func (m *Merkle) GetMerkleTreeList(ts types.Timestamp) ([]*MerkleNode, []*MerkleNode, error) {
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

	//now
	nowMerkleTreeList, err := m.GetMerkleTreeListCore(MerkleTreeLevelNow, offsetHourTS, ts)
	//log.Debug("GetMerkleTreeList: after now", "offsetHourTS", offsetHourTS, "ts", ts, "now", len(nowMerkleTreeList), "e", err)
	if err != nil {
		return nil, nil, err
	}

	lenList := len(yearMerkleTreeList) + len(monthMerkleTreeList) + len(dayMerkleTreeList) + len(hrMerkleTreeList)

	merkleTreeList := make([]*MerkleNode, 0, lenList)
	merkleTreeList = append(merkleTreeList, yearMerkleTreeList...)
	merkleTreeList = append(merkleTreeList, monthMerkleTreeList...)
	merkleTreeList = append(merkleTreeList, dayMerkleTreeList...)
	merkleTreeList = append(merkleTreeList, hrMerkleTreeList...)

	return merkleTreeList, nowMerkleTreeList, nil
}

func (m *Merkle) GetMerkleTreeListByLevel(level MerkleTreeLevel, ts types.Timestamp, nextTS types.Timestamp) ([]*MerkleNode, error) {
	return m.GetMerkleTreeListCore(level, ts, nextTS)
}

func (m *Merkle) GetMerkleTreeListCore(level MerkleTreeLevel, ts types.Timestamp, nextTS types.Timestamp) ([]*MerkleNode, error) {
	//log.Debug("GetMerkleTreeListCore: start", "level", level, "ts", ts, "nextTS", nextTS)
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
			//log.Debug("GetMerkleTreeListCore (after-loop)", "idx", fmt.Sprintf("(%d/%d)", i, len(results)), "result", result)
		}
	*/

	//log.Debug("GetMerkleTreeListCore: end", "level", level, "ts", ts, "nextTS", nextTS, "results", len(results))
	return results, nil
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

func ValidateMerkleTree(myNodes []*MerkleNode, theirNodes []*MerkleNode, ts types.Timestamp) bool {
	myNodes = validateMerkleTreeTrimNodes(myNodes, ts)
	theirNodes = validateMerkleTreeTrimNodes(theirNodes, ts)

	lenMyNodes := len(myNodes)
	lenTheirNodes := len(theirNodes)
	if lenMyNodes != lenTheirNodes {
		log.Error("ValidateMerkleTree: len", "ts", ts, "lenMyNodes", lenMyNodes, "lenTheirNodes", lenTheirNodes)
		return false
	}

	for i, pMyNode, pTheirNode := 0, myNodes, theirNodes; i < lenMyNodes; i, pMyNode, pTheirNode = i+1, pMyNode[1:], pTheirNode[1:] {
		if !reflect.DeepEqual(pMyNode[0].Addr, pTheirNode[0].Addr) {
			log.Error("ValidateMerkleTree: invalid", "i", i, "len", lenMyNodes, "myNode", pMyNode[0], "theirNode", pTheirNode[0])
			return false
		}
	}

	return true
}

func validateMerkleTreeTrimNodes(nodes []*MerkleNode, ts types.Timestamp) []*MerkleNode {
	nNodes := len(nodes)
	idx := sort.Search(nNodes, func(i int) bool {
		return ts.IsLessEqual(nodes[i].UpdateTS)
	})

	return nodes[:idx]
}
