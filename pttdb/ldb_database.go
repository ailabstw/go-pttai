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

package pttdb

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LDBDatabase struct {
	name string      // filename not including data-dir
	fn   string      // filename for reporting
	db   *leveldb.DB // LevelDB instance

	compTimeMeter    metrics.Meter // Meter for measuring the total time spent in database compaction
	compReadMeter    metrics.Meter // Meter for measuring the data read during compaction
	compWriteMeter   metrics.Meter // Meter for measuring the data written during compaction
	writeDelayNMeter metrics.Meter // Meter for measuring the write delay number due to database compaction
	writeDelayMeter  metrics.Meter // Meter for measuring the write delay duration due to database compaction
	diskReadMeter    metrics.Meter // Meter for measuring the effective amount of data read
	diskWriteMeter   metrics.Meter // Meter for measuring the effective amount of data written

	quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

	log log.Logger // Contextual logger tracking the database path

	lockLockMap sync.Mutex
	lockMap     map[string]int
}

// NewLDBDatabase returns a LevelDB wrapped object.
func NewLDBDatabase(file string, dataDir string, cache int, handles int) (*LDBDatabase, error) {
	fullFilename := filepath.Join(dataDir, file)

	logger := log.New("database", fullFilename)

	// Ensure we have some minimal caching and file guarantees
	if cache < minCache {
		cache = minCache
	}
	if handles < minHandles {
		handles = minHandles
	}

	logger.Info("Allocated cache and file handles", "cache", cache, "handles", handles)

	// Open the db and recover any potential corruptions
	db, err := leveldb.OpenFile(fullFilename, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
		CompactionTableSize:    128 * opt.MiB,
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(fullFilename, nil)
	}
	// (Re)check for errors and abort if opening of the db failed
	if err != nil {
		return nil, err
	}
	return &LDBDatabase{
		name:    file,
		fn:      fullFilename,
		db:      db,
		log:     logger,
		lockMap: make(map[string]int),
	}, nil
}

// Path returns the path to the database directory.
func (db *LDBDatabase) Path() string {
	return db.fn
}

func (db *LDBDatabase) Name() string {
	return db.name
}

/*
TryPut tries to put the key/val based on the updateTS of val.

Value is the jsonified bytes. the obj of the bytes needs to include UpdateTS.
*/
func (db *LDBDatabase) TryPut(key []byte, value []byte, updateTS types.Timestamp) ([]byte, error) {
	//log.Debug("TryPut: start", "key", key)

	err := db.TryLockMap(key)
	if err != nil {
		return nil, err
	}
	defer db.UnlockMap(key)

	isHasKey, err := db.Has(key)
	if err != nil {
		return nil, err
	}

	if !isHasKey { // new-one
		err := db.Put(key, value)
		return nil, err
	}

	v, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	d := &DBable{}

	err = json.Unmarshal(v, d)
	if err != nil {
		return nil, ErrInvalidDBable
	}

	if updateTS.IsLess(d.UpdateTS) {
		log.Warn("updateTS < d.UpdateTS", "updateTS", updateTS, "d.UpdateTS", d.UpdateTS, "key", key)
		return v, ErrInvalidUpdateTS
	}

	// put to db

	err = db.Put(key, value)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (db *LDBDatabase) TryLockMap(key []byte) error {
	mapKey := string(key)

	db.lockLockMap.Lock()
	defer db.lockLockMap.Unlock()

	//log.Debug("to TryLockMap", "key", key)

	// try to get lock
	val, ok := db.lockMap[mapKey]
	if ok { // someone-else is using the map-key
		log.Error("TryLockMap: busy", "mapKey", mapKey, "val", val)
		return ErrBusy
	}
	db.lockMap[mapKey] = -1

	return nil
}

func (db *LDBDatabase) UnlockMap(key []byte) error {
	mapKey := string(key)

	db.lockLockMap.Lock()
	defer db.lockLockMap.Unlock()

	_, ok := db.lockMap[mapKey]
	if !ok { // should not happen
		return ErrInvalidLock
	}

	delete(db.lockMap, mapKey)

	return nil
}

func (db *LDBDatabase) TryRLockMap(key []byte) error {
	mapKey := string(key)

	db.lockLockMap.Lock()
	defer db.lockLockMap.Unlock()

	// try to get lock
	i, ok := db.lockMap[mapKey]
	if ok && i < 0 { // write-lock
		log.Error("TryRLockMap: busy", "i", i, "ok", ok)
		return ErrBusy
	}

	log.Debug("after TryRLockMap (pass lock)", "i", i, "ok", ok)

	if !ok {
		db.lockMap[mapKey] = 0
	}
	db.lockMap[mapKey]++

	return nil
}

func (db *LDBDatabase) RUnlockMap(key []byte) error {
	mapKey := string(key)

	db.lockLockMap.Lock()
	defer db.lockLockMap.Unlock()

	i, ok := db.lockMap[mapKey]
	if !ok || i == 0 { // should not happen
		panic("db invalid unlock")
	}
	db.lockMap[mapKey]--

	if i == 1 {
		delete(db.lockMap, mapKey)
	}

	return nil
}

// Put puts the given key / value to the queue
func (db *LDBDatabase) Put(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

func (db *LDBDatabase) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

// Get returns the given key if it's present.
func (db *LDBDatabase) Get(key []byte) ([]byte, error) {
	dat, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Delete deletes the key from the queue and database
func (db *LDBDatabase) Delete(key []byte) error {
	err := db.db.Delete(key, nil)
	if err == leveldb.ErrNotFound {
		err = nil
	}

	return err
}

// Delete With Get
func (db *LDBDatabase) Pop(key []byte) ([]byte, error) {
	err := db.TryLockMap(key)
	if err != nil {
		return nil, err
	}
	defer db.UnlockMap(key)

	val, err := db.Get(key)
	if err != nil {
		// Unable to get key. possibly no key in the db. no need to do delete
		return nil, err
	}

	err = db.Delete(key)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (db *LDBDatabase) NewIterator(listOrder ListOrder) iterator.Iterator {
	iter := db.db.NewIterator(nil, nil)
	if listOrder == ListOrderPrev {
		iter.Seek(dbLastKey)
	}

	return iter
}

func (db *LDBDatabase) NewIteratorWithRange(r *util.Range, listOrder ListOrder) iterator.Iterator {
	iter := db.db.NewIterator(r, nil)
	if listOrder == ListOrderPrev {
		iter.Seek(r.Limit)
	}

	return iter
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
func (db *LDBDatabase) NewIteratorWithPrefix(start []byte, prefix []byte, listOrder ListOrder) (iterator.Iterator, error) {
	// both as nil
	if len(start) == 0 && len(prefix) == 0 {
		return db.NewIterator(listOrder), nil
	}

	// start as nil
	if len(start) == 0 {
		r := util.BytesPrefix(prefix)
		return db.NewIteratorWithRange(r, listOrder), nil
	}

	// prefix as nil
	if len(prefix) == 0 {
		startRange := util.BytesPrefix(start)
		var r *util.Range
		switch listOrder {
		case ListOrderPrev:
			r = &util.Range{Limit: startRange.Limit}
		case ListOrderNext:
			r = &util.Range{Start: startRange.Start}
		}

		return db.NewIteratorWithRange(r, listOrder), nil
	}

	// both non-nil
	if !strings.HasPrefix(string(start), string(prefix)) {
		return nil, ErrInvalidPrefix
	}

	startRange := util.BytesPrefix(start)
	prefixRange := util.BytesPrefix(prefix)
	var r *util.Range
	switch listOrder {
	case ListOrderPrev:
		r = &util.Range{Start: prefixRange.Start, Limit: startRange.Limit}
	case ListOrderNext:
		r = &util.Range{Start: startRange.Start, Limit: prefixRange.Limit}
	}

	return db.NewIteratorWithRange(r, listOrder), nil
}

func (db *LDBDatabase) Close() {
	// Stop the metrics collection to avoid internal database races
	db.quitLock.Lock()
	defer db.quitLock.Unlock()

	if db.quitChan != nil {
		errc := make(chan error)
		db.quitChan <- errc
		if err := <-errc; err != nil {
			db.log.Error("Metrics collection failed", "err", err)
		}
		db.quitChan = nil
	}
	err := db.db.Close()
	if err == nil {
		db.log.Info("Database closed")
	} else {
		db.log.Error("Failed to close database", "err", err)
	}
}

func (db *LDBDatabase) LDB() *leveldb.DB {
	return db.db
}

// Meter configures the database metrics collectors and
func (db *LDBDatabase) Meter(prefix string) {
	if metrics.Enabled {
		// Initialize all the metrics collector at the requested prefix
		db.compTimeMeter = metrics.NewRegisteredMeter(prefix+"compact/time", nil)
		db.compReadMeter = metrics.NewRegisteredMeter(prefix+"compact/input", nil)
		db.compWriteMeter = metrics.NewRegisteredMeter(prefix+"compact/output", nil)
		db.diskReadMeter = metrics.NewRegisteredMeter(prefix+"disk/read", nil)
		db.diskWriteMeter = metrics.NewRegisteredMeter(prefix+"disk/write", nil)
	}
	// Initialize write delay metrics no matter we are in metric mode or not.
	db.writeDelayMeter = metrics.NewRegisteredMeter(prefix+"compact/writedelay/duration", nil)
	db.writeDelayNMeter = metrics.NewRegisteredMeter(prefix+"compact/writedelay/counter", nil)

	// Create a quit channel for the periodic collector and run it
	db.quitLock.Lock()
	db.quitChan = make(chan chan error)
	db.quitLock.Unlock()

	go db.meter(3 * time.Second)
}

// meter periodically retrieves internal leveldb counters and reports them to
// the metrics subsystem.
//
// This is how a stats table look like (currently):
//   Compactions
//    Level |   Tables   |    Size(MB)   |    Time(sec)  |    Read(MB)   |   Write(MB)
//   -------+------------+---------------+---------------+---------------+---------------
//      0   |          0 |       0.00000 |       1.27969 |       0.00000 |      12.31098
//      1   |         85 |     109.27913 |      28.09293 |     213.92493 |     214.26294
//      2   |        523 |    1000.37159 |       7.26059 |      66.86342 |      66.77884
//      3   |        570 |    1113.18458 |       0.00000 |       0.00000 |       0.00000
//
// This is how the write delay look like (currently):
// DelayN:5 Delay:406.604657ms Paused: false
//
// This is how the iostats look like (currently):
// Read(MB):3895.04860 Write(MB):3654.64712
func (db *LDBDatabase) meter(refresh time.Duration) {
	// Create the counters to store current and previous compaction values
	compactions := make([][]float64, 2)
	for i := 0; i < 2; i++ {
		compactions[i] = make([]float64, 3)
	}
	// Create storage for iostats.
	var iostats [2]float64

	// Create storage and warning log tracer for write delay.
	var (
		delaystats      [2]int64
		lastWriteDelay  time.Time
		lastWriteDelayN time.Time
		lastWritePaused time.Time
	)

	var (
		errc chan error
		merr error
	)

	// Iterate ad infinitum and collect the stats
	for i := 1; errc == nil && merr == nil; i++ {
		// Retrieve the database stats
		stats, err := db.db.GetProperty("leveldb.stats")
		if err != nil {
			db.log.Error("Failed to read database stats", "err", err)
			merr = err
			continue
		}
		// Find the compaction table, skip the header
		lines := strings.Split(stats, "\n")
		for len(lines) > 0 && strings.TrimSpace(lines[0]) != "Compactions" {
			lines = lines[1:]
		}
		if len(lines) <= 3 {
			db.log.Error("Compaction table not found")
			merr = errors.New("compaction table not found")
			continue
		}
		lines = lines[3:]

		// Iterate over all the table rows, and accumulate the entries
		for j := 0; j < len(compactions[i%2]); j++ {
			compactions[i%2][j] = 0
		}
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) != 6 {
				break
			}
			for idx, counter := range parts[3:] {
				value, err := strconv.ParseFloat(strings.TrimSpace(counter), 64)
				if err != nil {
					db.log.Error("Compaction entry parsing failed", "err", err)
					merr = err
					continue
				}
				compactions[i%2][idx] += value
			}
		}
		// Update all the requested meters
		if db.compTimeMeter != nil {
			db.compTimeMeter.Mark(int64((compactions[i%2][0] - compactions[(i-1)%2][0]) * 1000 * 1000 * 1000))
		}
		if db.compReadMeter != nil {
			db.compReadMeter.Mark(int64((compactions[i%2][1] - compactions[(i-1)%2][1]) * 1024 * 1024))
		}
		if db.compWriteMeter != nil {
			db.compWriteMeter.Mark(int64((compactions[i%2][2] - compactions[(i-1)%2][2]) * 1024 * 1024))
		}

		// Retrieve the write delay statistic
		writedelay, err := db.db.GetProperty("leveldb.writedelay")
		if err != nil {
			db.log.Error("Failed to read database write delay statistic", "err", err)
			merr = err
			continue
		}
		var (
			delayN        int64
			delayDuration string
			duration      time.Duration
			paused        bool
		)
		if n, err := fmt.Sscanf(writedelay, "DelayN:%d Delay:%s Paused:%t", &delayN, &delayDuration, &paused); n != 3 || err != nil {
			db.log.Error("Write delay statistic not found")
			merr = err
			continue
		}
		duration, err = time.ParseDuration(delayDuration)
		if err != nil {
			db.log.Error("Failed to parse delay duration", "err", err)
			merr = err
			continue
		}
		if db.writeDelayNMeter != nil {
			db.writeDelayNMeter.Mark(delayN - delaystats[0])
			// If the write delay number been collected in the last minute exceeds the predefined threshold,
			// print a warning log here.
			// If a warning that db performance is laggy has been displayed,
			// any subsequent warnings will be withhold for 1 minute to don't overwhelm the user.
			if int(db.writeDelayNMeter.Rate1()) > writeDelayNThreshold &&
				time.Now().After(lastWriteDelayN.Add(writeDelayWarningThrottler)) {
				db.log.Warn("Write delay number exceeds the threshold (200 per second) in the last minute")
				lastWriteDelayN = time.Now()
			}
		}
		if db.writeDelayMeter != nil {
			db.writeDelayMeter.Mark(duration.Nanoseconds() - delaystats[1])
			// If the write delay duration been collected in the last minute exceeds the predefined threshold,
			// print a warning log here.
			// If a warning that db performance is laggy has been displayed,
			// any subsequent warnings will be withhold for 1 minute to don't overwhelm the user.
			if int64(db.writeDelayMeter.Rate1()) > writeDelayThreshold.Nanoseconds() &&
				time.Now().After(lastWriteDelay.Add(writeDelayWarningThrottler)) {
				db.log.Warn("Write delay duration exceeds the threshold (35% of the time) in the last minute")
				lastWriteDelay = time.Now()
			}
		}
		// If a warning that db is performing compaction has been displayed, any subsequent
		// warnings will be withheld for one minute not to overwhelm the user.
		if paused && delayN-delaystats[0] == 0 && duration.Nanoseconds()-delaystats[1] == 0 &&
			time.Now().After(lastWritePaused.Add(writeDelayWarningThrottler)) {
			db.log.Warn("Database compacting, degraded performance")
			lastWritePaused = time.Now()
		}

		delaystats[0], delaystats[1] = delayN, duration.Nanoseconds()

		// Retrieve the database iostats.
		ioStats, err := db.db.GetProperty("leveldb.iostats")
		if err != nil {
			db.log.Error("Failed to read database iostats", "err", err)
			merr = err
			continue
		}
		var nRead, nWrite float64
		parts := strings.Split(ioStats, " ")
		if len(parts) < 2 {
			db.log.Error("Bad syntax of ioStats", "ioStats", ioStats)
			merr = fmt.Errorf("bad syntax of ioStats %s", ioStats)
			continue
		}
		if n, err := fmt.Sscanf(parts[0], "Read(MB):%f", &nRead); n != 1 || err != nil {
			db.log.Error("Bad syntax of read entry", "entry", parts[0])
			merr = err
			continue
		}
		if n, err := fmt.Sscanf(parts[1], "Write(MB):%f", &nWrite); n != 1 || err != nil {
			db.log.Error("Bad syntax of write entry", "entry", parts[1])
			merr = err
			continue
		}
		if db.diskReadMeter != nil {
			db.diskReadMeter.Mark(int64((nRead - iostats[0]) * 1024 * 1024))
		}
		if db.diskWriteMeter != nil {
			db.diskWriteMeter.Mark(int64((nWrite - iostats[1]) * 1024 * 1024))
		}
		iostats[0], iostats[1] = nRead, nWrite

		// Sleep a bit, then repeat the stats collection
		select {
		case errc = <-db.quitChan:
			// Quit requesting, stop hammering the database
		case <-time.After(refresh):
			// Timeout, gather a new set of stats
		}
	}

	if errc == nil {
		errc = <-db.quitChan
	}
	errc <- merr
}

func (db *LDBDatabase) NewBatch() Batch {
	return &ldbBatch{db: db, b: new(leveldb.Batch)}
}

type ldbBatch struct {
	db   *LDBDatabase
	b    *leveldb.Batch
	size int
}

func (b *ldbBatch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

func (b *ldbBatch) DBGet(key []byte) ([]byte, error) {
	return b.db.Get(key)
}

func (b *ldbBatch) DBDelete(key []byte) error {
	return b.db.Delete(key)
}

func (b *ldbBatch) Delete(key []byte) error {
	b.b.Delete(key)

	return nil
}

func (b *ldbBatch) Write() error {
	db := b.db.LDB()
	return db.Write(b.b, nil)
}

func (b *ldbBatch) ValueSize() int {
	return b.size
}

func (b *ldbBatch) Reset() {
	b.b.Reset()
	b.size = 0
}

func (b *ldbBatch) DB() *LDBDatabase {
	return b.db
}

type LDBBatch struct {
	*ldbBatch
}

func NewLDBBatch(db *LDBDatabase) (*LDBBatch, error) {
	batch := &ldbBatch{
		db: db,
		b:  new(leveldb.Batch),
	}

	return &LDBBatch{
		ldbBatch: batch,
	}, nil
}

func (b *LDBBatch) PutAllWithKeyIndex(key []byte, idx *Index, kvs []*KeyVal) error {
	b.Reset()
	marshaledIdx, err := idx.Marshal()
	if err != nil {
		return err
	}

	err = b.Put(key, marshaledIdx)
	if err != nil {
		return err
	}

	return b.PutAll(kvs, false)
}

func (b *LDBBatch) PutAll(kvs []*KeyVal, isInit bool) error {
	if isInit {
		b.Reset()
	}
	for _, kv := range kvs {
		err := b.Put(kv.K, kv.V)
		if err != nil {
			return err
		}
	}

	err := b.Write()
	if err != nil {
		return err
	}

	b.Reset()
	return nil
}

/*
TryPutAll tries to put all the key-vals with comparing the updateTS of the 1st item.

This is used when the real content is stored in ts-based key, but we want to refer the content directly from id-key.

This assumes 1-record per key. The old data will be deleted if key conflict.

For oplog: the id of each oplog is unique.
For content:
*/
func (b *LDBBatch) TryPutAll(idxKey []byte, idx *Index, kvs []*KeyVal, isDeleteOrig bool, isGetOrig bool) ([]*KeyVal, error) {
	log.Debug("TryPutAll: start", "idxKey", idxKey)

	db := b.ldbBatch.DB()

	err := db.TryLockMap(idxKey)
	if err != nil {
		log.Error("TryPutAll: unable to lock", "idxKey", idxKey, "e", err)
		return nil, err
	}
	defer db.UnlockMap(idxKey)

	isHasKey, err := db.Has(idxKey)
	if err != nil {
		log.Error("TryPutAll: unable to has", "idxKey", idxKey, "e", err)
		return nil, err
	}

	if !isHasKey { // new-one
		err := b.PutAllWithKeyIndex(idxKey, idx, kvs)
		log.Debug("TryPutAll: after PutAllWithKeyIndex (new-one)", "idxKey", idxKey, "e", err)
		return nil, err
	}

	v, err := db.Get(idxKey)
	if err != nil {
		log.Error("TryPutAll: unable to Get", "idxKey", idxKey, "e", err)
		return nil, err
	}

	d := &Index{}
	err = d.Unmarshal(v)
	if err != nil { // unable to get original data.
		log.Error("TryPutAll: unable to unmarshal index", "idxKey", idxKey, "v", v, "e", err)
		return nil, ErrInvalidDBable
	}
	var origKVs []*KeyVal
	i := 0
	var key []byte
	if isGetOrig {
		origKVs = make([]*KeyVal, len(d.Keys))
		for i, key = range d.Keys {
			v, err = db.Get(key)
			if err != nil {
				log.Error("TryPutAll: (GetOrig) unable to get key", "idxKey", idxKey, "k", key, "e", err)
				return nil, err
			}
			origKVs[i] = &KeyVal{
				K: key,
				V: v,
			}
		}
	}

	if idx.UpdateTS.IsLess(d.UpdateTS) {
		log.Warn("updateTS < d.UpdateTS", "idxKey", idxKey, "updateTS", idx.UpdateTS, "d.UpdateTS", d.UpdateTS, "d.Keys", d.Keys)
		return origKVs, ErrInvalidUpdateTS
	}

	// delete original data
	if isDeleteOrig {
		for _, eachKey := range d.Keys {
			db.Delete(eachKey)
		}
	}

	// put to db

	err = b.PutAllWithKeyIndex(idxKey, idx, kvs)
	if err != nil {
		log.Error("TryPutAll: unable to PutAllWithKeyIndex", "idxKey", idxKey, "e", err)
		return nil, err
	}

	return origKVs, nil
}

/*
TryPutAll tries to put all the key-vals with comparing the updateTS of the 1st item.

This is used when the real content is stored in ts-based key, but we want to refer the content directly from id-key.

This assumes 1-record per key. The old data will be deleted if key conflict.

For oplog: the id of each oplog is unique.
For content:
*/
func (b *LDBBatch) TryPutAllSameUT(idxKey []byte, idx *Index, kvs []*KeyVal, isDeleteOrig bool) ([][]byte, error) {
	log.Debug("TryPutAllSameUT: start", "idxKey", idxKey)

	db := b.ldbBatch.DB()

	err := db.TryLockMap(idxKey)
	if err != nil {
		log.Error("TryPutAllSameUT: unable to lock", "idxKey", idxKey, "e", err)
		return nil, err
	}
	defer db.UnlockMap(idxKey)

	isHasKey, err := db.Has(idxKey)
	if err != nil {
		log.Error("TryPutAllSameUT: unable to has", "idxKey", idxKey, "e", err)
		return nil, err
	}

	if !isHasKey { // new-one
		err := b.PutAllWithKeyIndex(idxKey, idx, kvs)
		log.Debug("TryPutAllSameUT: after PutAllWithKeyIndex (new-one)", "idxKey", idxKey, "e", err)
		return nil, err
	}

	v, err := db.Get(idxKey)
	if err != nil {
		log.Error("TryPutAllSameUT: unable to Get", "idxKey", idxKey, "e", err)
		return nil, err
	}

	d := &Index{}
	err = d.Unmarshal(v)
	if err != nil { // unable to get original data.
		return nil, ErrInvalidDBable
	}

	if idx.UpdateTS.IsLess(d.UpdateTS) {
		log.Warn("TryPutAllSameUT: updateTS < d.UpdateTS", "idxKey", idxKey, "updateTS", idx.UpdateTS, "d.UpdateTS", d.UpdateTS)
		return d.Keys, ErrInvalidUpdateTS
	}

	if idx.UpdateTS == d.UpdateTS && !reflect.DeepEqual(d, idx) {
		log.Warn("updateTS == d.UpdateTS but idx diff", "idxKey", idxKey, "updateTS", idx.UpdateTS, "d.UpdateTS", d.UpdateTS, "idx", idx, "d", d)
		return d.Keys, ErrInvalidUpdateTS
	}

	// delete original data
	if isDeleteOrig {
		for _, eachKey := range d.Keys {
			db.Delete(eachKey)
		}
	}

	// put to db

	err = b.PutAllWithKeyIndex(idxKey, idx, kvs)
	if err != nil {
		log.Error("TryPutAllSameUT: unable to PutAllWithKeyIndex", "idxKey", idxKey, "e", err)
		return nil, err
	}

	return d.Keys, nil
}

/*
ForcePutAll tries to put all the key-vals with comparing the updateTS of the 1st item.

This is used when the real content is stored in ts-based key, but we want to refer the content directly from id-key.

This assumes 1-record per key. The old data will be deleted if key conflict.

For oplog: the id of each oplog is unique.
For content:
*/
func (b *LDBBatch) ForcePutAll(idxKey []byte, idx *Index, kvs []*KeyVal) ([][]byte, error) {
	log.Debug("ForcePutAll: start", "idxKey", idxKey)

	db := b.ldbBatch.DB()

	err := db.TryLockMap(idxKey)
	if err != nil {
		log.Error("ForcePutAll: unable to lock", "idxKey", idxKey, "e", err)
		return nil, err
	}
	defer db.UnlockMap(idxKey)

	isHasKey, err := db.Has(idxKey)
	if err != nil {
		log.Error("ForcePutAll: unable to has", "idxKey", idxKey, "e", err)
		return nil, err
	}

	if !isHasKey { // new-one
		err := b.PutAllWithKeyIndex(idxKey, idx, kvs)
		log.Debug("ForcePutAll: after PutAllWithKeyIndex (new-one)", "idxKey", idxKey, "e", err)
		return nil, err
	}

	v, err := db.Get(idxKey)
	if err != nil {
		log.Error("ForcePutAll: unable to Get", "idxKey", idxKey, "e", err)
		return nil, err
	}

	d := &Index{}
	err = d.Unmarshal(v)
	if err != nil { // unable to get original data.
		return nil, ErrInvalidDBable
	}

	// delete original data
	for _, eachKey := range d.Keys {
		db.Delete(eachKey)
	}

	// put to db

	err = b.PutAllWithKeyIndex(idxKey, idx, kvs)
	if err != nil {
		log.Error("TryPutAll: unable to PutAllWithKeyIndex", "idxKey", idxKey, "e", err)
		return nil, err
	}

	return d.Keys, nil
}

func (b *LDBBatch) DeleteAllKeys(keys [][]byte) error {
	if keys == nil {
		return nil
	}

	db := b.ldbBatch.DB()
	for _, key := range keys {
		db.Delete(key)
	}

	return nil
}

func (b *LDBBatch) DeleteAll(idxKey []byte) error {
	db := b.ldbBatch.DB()

	err := db.TryLockMap(idxKey)
	if err != nil {
		return err
	}
	defer db.UnlockMap(idxKey)

	v, err := db.Get(idxKey)
	if err == leveldb.ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	d := &Index{}
	err = d.Unmarshal(v)
	if err != nil { // unable to get original data.
		return ErrInvalidDBable
	}

	for _, eachKey := range d.Keys {
		db.Delete(eachKey)
	}

	return db.Delete(idxKey)
}

func (b *LDBBatch) GetByIdxKey(idxKey []byte, idx int) ([]byte, error) {
	db := b.ldbBatch.DB()
	err := db.TryRLockMap(idxKey)
	if err != nil {
		return nil, err
	}
	defer db.RUnlockMap(idxKey)

	v, err := db.Get(idxKey)
	if err != nil {
		return nil, err
	}

	d := &Index{}
	err = d.Unmarshal(v)
	if err != nil { // unable to get original data.
		return nil, ErrInvalidDBable
	}

	if d.Keys == nil || len(d.Keys) <= idx || d.Keys[idx] == nil {
		return nil, ErrInvalidKeys
	}

	v, err = db.Get(d.Keys[idx])
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (b *LDBBatch) GetKeyByIdxKey(idxKey []byte, idx int) ([]byte, error) {
	db := b.ldbBatch.DB()
	err := db.TryRLockMap(idxKey)
	if err != nil {
		return nil, err
	}
	defer db.RUnlockMap(idxKey)

	v, err := db.Get(idxKey)
	if err != nil {
		return nil, err
	}

	d := &Index{}
	err = d.Unmarshal(v)
	if err != nil { // unable to get original data.
		return nil, ErrInvalidDBable
	}

	if d.Keys == nil || len(d.Keys) <= idx || d.Keys[idx] == nil {
		return nil, ErrInvalidKeys
	}

	return d.Keys[idx], nil
}

func (b *LDBBatch) GetBy2ndIdxKey(idxKey []byte) ([]byte, error) {
	db := b.ldbBatch.DB()
	err := db.TryRLockMap(idxKey)
	if err != nil {
		return nil, err
	}
	defer db.RUnlockMap(idxKey)

	key, err := db.Get(idxKey)
	if err != nil {
		return nil, err
	}

	v, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (b *LDBBatch) GetKeyBy2ndIdxKey(idxKey []byte) ([]byte, error) {
	db := b.ldbBatch.DB()
	err := db.TryRLockMap(idxKey)
	if err != nil {
		return nil, err
	}
	defer db.RUnlockMap(idxKey)

	key, err := db.Get(idxKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}
