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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestLDBDatabase_NewPrevIteratorWithPrefix(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)
	tDefaultDB.Put([]byte("test123"), []byte("test"))

	// define test-structure
	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]int
	}
	type args struct {
		start  []byte
		prefix []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		db      *LDBDatabase
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			db: tDefaultDB,
			args: args{
				start:  []byte("test123"),
				prefix: []byte("test"),
			},
			want: []byte("test123"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.db
			iter, err := db.NewIteratorWithPrefix(tt.args.start, tt.args.prefix, ListOrderPrev)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got []byte
			for iter.Prev() {
				key := iter.Key()
				got = common.CloneBytes(key)
				break
			}
			t.Logf("LDBDatabase.NewPrevIteratorWithPrefix: got: %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestLDBDatabase_NewPrevIteratorWithPrefix2(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)
	tDefaultDB.Put([]byte("test123"), []byte("test"))
	tDefaultDB.Put([]byte("test125"), []byte("test"))

	// define test-structure
	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]int
	}
	type args struct {
		start  []byte
		prefix []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		db      *LDBDatabase
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			db: tDefaultDB,
			args: args{
				start:  []byte("test123"),
				prefix: []byte("test"),
			},
			want: []byte("test123"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.db
			iter, err := db.NewIteratorWithPrefix(tt.args.start, tt.args.prefix, ListOrderPrev)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got []byte
			for iter.Prev() {
				key := iter.Key()
				got = common.CloneBytes(key)
				break
			}
			t.Logf("LDBDatabase.NewPrevIteratorWithPrefix2: got: %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestLDBDatabase_NewPrevIteratorWithPrefix3(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)
	tDefaultDB.Put([]byte("test123"), []byte("test"))
	tDefaultDB.Put([]byte("test124"), []byte("test"))

	// define test-structure
	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]int
	}
	type args struct {
		start  []byte
		prefix []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		db      *LDBDatabase
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			db: tDefaultDB,
			args: args{
				start:  []byte("test123"),
				prefix: []byte("test"),
			},
			want: []byte("test123"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.db
			iter, err := db.NewIteratorWithPrefix(tt.args.start, tt.args.prefix, ListOrderPrev)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got []byte
			for iter.Prev() {
				key := iter.Key()
				got = common.CloneBytes(key)
				break
			}
			t.Logf("LDBDatabase.NewPrevIteratorWithPrefix3: got: %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestLDBDatabase_NewPrevIteratorWithPrefix4(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)
	tDefaultDB.Put([]byte("test122"), []byte("test"))
	tDefaultDB.Put([]byte("test123"), []byte("test"))
	tDefaultDB.Put([]byte("test124"), []byte("test"))

	// define test-structure
	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]int
	}
	type args struct {
		start  []byte
		prefix []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		db      *LDBDatabase
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			db: tDefaultDB,
			args: args{
				start:  []byte("test123"),
				prefix: []byte("test"),
			},
			want: []byte("test123"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.db
			iter, err := db.NewIteratorWithPrefix(tt.args.start, tt.args.prefix, ListOrderPrev)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got []byte
			for iter.Prev() {
				key := iter.Key()
				got = common.CloneBytes(key)
				break
			}
			t.Logf("LDBDatabase.NewPrevIteratorWithPrefix4: got: %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestLDBDatabase_NewPrevIteratorWithPrefix5(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)
	tDefaultDB.Put([]byte("test122"), []byte("test"))
	tDefaultDB.Put([]byte("test123"), []byte("test"))
	tDefaultDB.Put([]byte("test124"), []byte("test"))

	// define test-structure
	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]int
	}
	type args struct {
		start  []byte
		prefix []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		db      *LDBDatabase
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			db: tDefaultDB,
			args: args{
				start:  []byte("test123"),
				prefix: []byte("test"),
			},
			want: []byte("test122"),
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.db
			iter, err := db.NewIteratorWithPrefix(tt.args.start, tt.args.prefix, ListOrderPrev)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got []byte
			for iter.Prev() {
				key := iter.Key()
				got = common.CloneBytes(key)
			}
			t.Logf("LDBDatabase.NewPrevIteratorWithPrefix4: got: %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDBDatabase.NewPrevIteratorWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestLDBDatabase_TryPutGetDelete(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]bool
	}
	type args struct {
		key      []byte
		value    []byte
		updateTS types.Timestamp
	}
	type dbvalue struct {
		UpdateTS types.Timestamp `json:"UT"`
		Value    []byte
	}

	beginningTS, err := types.GetTimestamp()
	if err != nil {
		panic("failed to create beginning time stamp: " + err.Error())
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args:    args{key: []byte(""), value: []byte("")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("a"), value: []byte("a")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("1251"), value: []byte("1251")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("\x00123\x00"), value: []byte("\x00123\x00")},
			wantErr: false,
		},
	}

	//
	//  Test Get method from an empty DB
	//
	testName := "test Get method from empty db"
	for i, tt := range tests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			data, err := tDefaultDB.Get(tt.args.key)
			if err == nil {
				t.Errorf("LDBDatabase.Get() error is nil, wantErr: leveldb not found")
			}
			if data != nil {
				t.Fatalf("get returned wrong result, got %q expected nil", string(data))
			}
		})
	}

	tryPutUpdateTS, err := types.GetTimestamp()

	//
	//  Test TryPut method
	//
	testName = "test TryPut method"
	for i, tt := range tests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			dbValue, err := json.Marshal(dbvalue{UpdateTS: tryPutUpdateTS, Value: tt.args.value})
			if err != nil {
				panic("failed to marshal: " + err.Error())
			}
			if _, err := tDefaultDB.TryPut(tt.args.key, dbValue, tryPutUpdateTS); (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.TryPut() error = %v, wantErr %v", err, tt.wantErr)
			}
			data, err := tDefaultDB.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(data, dbValue) {
				t.Fatalf("get returned wrong result, got %q expected %q", string(data), string(dbValue))
			}
		})
	}

	updateTSTests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Existing keys
		{
			args:    args{key: []byte(""), value: []byte("")},
			wantErr: true,
		},
		{
			args:    args{key: []byte("a"), value: []byte("a")},
			wantErr: true,
		},
		{
			args:    args{key: []byte("1251"), value: []byte("1251")},
			wantErr: true,
		},
		{
			args:    args{key: []byte("\x00123\x00"), value: []byte("\x00123\x00")},
			wantErr: true,
		},
		// New keys
		{
			args:    args{key: []byte("empty"), value: []byte("empty")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("b"), value: []byte("a")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("9999"), value: []byte("1251")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("\x00456\x00"), value: []byte("\x00123\x00")},
			wantErr: false,
		},
	}

	//
	//  Test TryPut method and check updateTS
	//
	testName = "test TryPut and check updateTS"
	for i, tt := range updateTSTests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			dbValue, err := json.Marshal(dbvalue{UpdateTS: beginningTS, Value: tt.args.value})
			if err != nil {
				panic("failed to marshal: " + err.Error())
			}
			_, err = tDefaultDB.TryPut(tt.args.key, dbValue, beginningTS)
			if tt.wantErr {
				if err != ErrInvalidUpdateTS {
					t.Errorf("LDBDatabase.TryPut() should error, got %v, expect %v", err, ErrInvalidUpdateTS)
				}
			} else {
				if err != nil {
					t.Errorf("LDBDatabase.TryPut() error = %v", err)
				}
			}
		})
	}

	//
	//  Test Get method after key overrides
	//
	testName = "test Get method after key overrides"
	for i, tt := range updateTSTests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			var dbValue []byte
			if tt.wantErr {
				dbValue, err = json.Marshal(dbvalue{UpdateTS: tryPutUpdateTS, Value: tt.args.value})
				if err != nil {
					panic("failed to marshal: " + err.Error())
				}
			} else {
				dbValue, err = json.Marshal(dbvalue{UpdateTS: beginningTS, Value: tt.args.value})
				if err != nil {
					panic("failed to marshal: " + err.Error())
				}
			}

			data, err := tDefaultDB.Get(tt.args.key)
			if err != nil {
				t.Errorf("LDBDatabase.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(data, dbValue) {
				t.Fatalf("get returned wrong result, got %q expected %q", string(data), string(dbValue))
			}
		})
	}

	//
	//  Test Delete method
	//
	testName = "test Delete method"
	for i, tt := range updateTSTests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			err := tDefaultDB.Delete(tt.args.key)
			if err != nil {
				t.Errorf("LDBDatabase.Delete() error = %v", err)
			}
		})
	}

	//
	//  Test Get method after Deletion
	//
	testName = "test Get method after Deletion"
	for i, tt := range updateTSTests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			data, err := tDefaultDB.Get(tt.args.key)
			if err == nil {
				t.Errorf("LDBDatabase.Get() error is nil, wantErr: leveldb not found")
			}
			if data != nil {
				t.Fatalf("get returned wrong result, got %q expected nil", string(data))
			}
		})
	}

	duplicateTests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args:    args{key: []byte("a"), value: []byte("a")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("a"), value: []byte("a")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("b"), value: []byte("b")},
			wantErr: false,
		},
		{
			args:    args{key: []byte("b"), value: []byte("b")},
			wantErr: false,
		},
	}

	//
	//  Test TryPut method on duplicate keys
	//
	testName = "test TryPut method on duplicate keys"
	for i, tt := range duplicateTests {
		name := fmt.Sprintf("%s %v", testName, i)
		t.Run(name, func(t *testing.T) {
			updateTS, err := types.GetTimestamp()
			if err != nil {
				panic("failed to create update time stamp: " + err.Error())
			}
			dbValue, err := json.Marshal(dbvalue{UpdateTS: updateTS, Value: tt.args.value})
			if _, err := tDefaultDB.TryPut(tt.args.key, dbValue, updateTS); err != nil {
				t.Errorf("LDBDatabase.TryPut() error = %v", err)
			}
			data, err := tDefaultDB.Get(tt.args.key)
			if err != nil {
				t.Errorf("LDBDatabase.Get() error is nil, wantErr:")
			}
			if data == nil {
				t.Errorf("LDBDatabase.Get() error is nil, wantErr: data gone")
			}
		})
	}
}

func TestLDBDatabase_NewIteratorWithPrefix(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)
	tDefaultDB.Put([]byte("test1"), []byte("test1_value"))
	tDefaultDB.Put([]byte("test2"), []byte("test2_value"))
	tDefaultDB.Put([]byte("test3"), []byte("test3_value"))
	tDefaultDB.Put([]byte("1109"), []byte("1109_value"))
	tDefaultDB.Put([]byte("1110"), []byte("1110_value"))
	tDefaultDB.Put([]byte("1111"), []byte("1111_value"))
	tDefaultDB.Put([]byte(""), []byte("_value"))

	type fields struct {
		name             string
		fn               string
		db               *leveldb.DB
		compTimeMeter    metrics.Meter
		compReadMeter    metrics.Meter
		compWriteMeter   metrics.Meter
		writeDelayNMeter metrics.Meter
		writeDelayMeter  metrics.Meter
		diskReadMeter    metrics.Meter
		diskWriteMeter   metrics.Meter
		quitLock         sync.Mutex
		quitChan         chan chan error
		log              log.Logger
		lockLockMap      sync.Mutex
		lockMap          map[string]bool
	}
	type args struct {
		prefix []byte
		start  []byte
	}

	tests := []struct {
		name    string
		db      *LDBDatabase
		args    args
		want    [][]byte
		wantErr bool
	}{
		{
			name: "test empty prefix and start",
			db:   tDefaultDB,
			args: args{
				prefix: []byte(""),
				start:  []byte(""),
			},
			want:    [][]byte{[]byte("_value"), []byte("1109_value"), []byte("1110_value"), []byte("1111_value"), []byte("test1_value"), []byte("test2_value"), []byte("test3_value")},
			wantErr: false,
		},
		{
			name: "test prefix with empty start",
			db:   tDefaultDB,
			args: args{
				prefix: []byte("111"),
				start:  []byte(""),
			},
			want:    [][]byte{[]byte("1110_value"), []byte("1111_value")},
			wantErr: false,
		},
		{
			name: "test start with empty prefix",
			db:   tDefaultDB,
			args: args{
				prefix: []byte(""),
				start:  []byte("1109"),
			},
			want:    [][]byte{[]byte("1109_value"), []byte("1110_value"), []byte("1111_value"), []byte("test1_value"), []byte("test2_value"), []byte("test3_value")},
			wantErr: false,
		},
		{
			name: "test non-existing prefix",
			db:   tDefaultDB,
			args: args{
				prefix: []byte("foo"),
				start:  []byte("test1"),
			},
			want:    [][]byte{},
			wantErr: true,
		},
		{
			name: "test non-existing start with empty prefix",
			db:   tDefaultDB,
			args: args{
				prefix: []byte(""),
				start:  []byte("foo"),
			},
			want:    [][]byte{[]byte("test1_value"), []byte("test2_value"), []byte("test3_value")},
			wantErr: false,
		},
		{
			name: "test non-existing start with prefix",
			db:   tDefaultDB,
			args: args{
				prefix: []byte("11"),
				start:  []byte("foo"),
			},
			want:    [][]byte{},
			wantErr: true,
		},
		{
			name: "test start field with prefix",
			db:   tDefaultDB,
			args: args{
				prefix: []byte("11"),
				start:  []byte("1110"),
			},
			want:    [][]byte{[]byte("1110_value"), []byte("1111_value")},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.NewIteratorWithPrefix(tt.args.start, tt.args.prefix, ListOrderNext)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBDatabase.NewIteratorWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if idx := 0; got != nil {
				for got.Next() {
					gotVal := got.Value()
					if idx >= len(tt.want) {
						t.Errorf("Iterator got too many values")
						return
					}

					if !bytes.Equal(gotVal, tt.want[idx]) {
						t.Errorf("Wrong result, got %q expected %q", string(gotVal), tt.want[idx])
					}

					idx = idx + 1
				}

				if idx < len(tt.want)-1 {
					t.Errorf("Iterator missed some values")
				}
			}
		})
	}
}

func TestLDBBatch_GetDelete(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	type fields struct {
		ldbBatch *ldbBatch
	}
	type args struct {
		name   string
		idxKey []byte
		idx    int
	}

	dbBatch, _ := NewLDBBatch(tDefaultDB)

	updateTS1, _ := types.GetTimestamp()
	updateTS2, _ := types.GetTimestamp()
	updateTS3, _ := types.GetTimestamp()

	dbBatch.TryPutAll(
		[]byte("test-idx-key"),
		&Index{
			Keys:     [][]byte{[]byte("test-idx-1"), []byte("test-idx-2"), []byte("test-idx-3")},
			UpdateTS: updateTS1,
		},
		[]*KeyVal{
			&KeyVal{K: []byte("test-idx-1"), V: []byte("test-value-1")},
			&KeyVal{K: []byte("test-idx-2"), V: []byte("test-value-2")},
			&KeyVal{K: []byte("test-idx-3"), V: []byte("test-value-3")},
		},
		false, false)

	dbBatch.TryPutAll(
		[]byte("test-idx-key2"),
		nil,
		nil,
		false, false)

	dbBatch.TryPutAll(
		[]byte("test-idx-key3"),
		&Index{
			Keys:     nil,
			UpdateTS: updateTS2,
		},
		nil,
		false, false)

	dbBatch.TryPutAll(
		[]byte("test-idx-key4"),
		&Index{
			Keys:     [][]byte{[]byte("test-idx2-1"), nil, []byte("test-idx2-3")},
			UpdateTS: updateTS3,
		},
		[]*KeyVal{
			&KeyVal{K: []byte("test-idx2-1"), V: []byte("test-value2-1")},
			&KeyVal{K: nil, V: nil},
			&KeyVal{K: []byte("test-idx2-3"), V: []byte("test-value2-3")},
		},
		false, false)

	testsGetByIdxKey := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test GetByIdxKey 0",
			args: args{
				idxKey: []byte("test-idx-key"),
				idx:    1,
			},
			want:    []byte("test-value-2"),
			wantErr: false,
		},
		{
			name: "test GetByIdxKey 1",
			args: args{
				idxKey: []byte("test-idx-key2"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetByIdxKey 2",
			args: args{
				idxKey: []byte("test-idx-key3"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetByIdxKey 3",
			args: args{
				idxKey: []byte("test-idx-key4"),
				idx:    0,
			},
			want:    []byte("test-value2-1"),
			wantErr: false,
		},
		{
			name: "test GetByIdxKey 4",
			args: args{
				idxKey: []byte("test-idx-key4"),
				idx:    1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetByIdxKey with non-existing idxKey",
			args: args{
				idxKey: []byte("test-idx-key99"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetByIdxKey with wrong idx",
			args: args{
				idxKey: []byte("test-idx-key"),
				idx:    10,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range testsGetByIdxKey {
		t.Run(tt.name, func(t *testing.T) {
			data, err := dbBatch.GetByIdxKey(tt.args.idxKey, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBBatch.GetByIdxKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("LDBBatch.GetByIdxKey(), got %q expected %q", string(data), string(tt.want))
			}
		})
	}

	testsGetKeyByIdxKey := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test GetKeyByIdxKey 0",
			args: args{
				idxKey: []byte("test-idx-key"),
				idx:    1,
			},
			want:    []byte("test-idx-2"),
			wantErr: false,
		},
		{
			name: "test GetKeyByIdxKey 1",
			args: args{
				idxKey: []byte("test-idx-key2"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetKeyByIdxKey 2",
			args: args{
				idxKey: []byte("test-idx-key3"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetKeyByIdxKey 3",
			args: args{
				idxKey: []byte("test-idx-key4"),
				idx:    0,
			},
			want:    []byte("test-idx2-1"),
			wantErr: false,
		},
		{
			name: "test GetKeyByIdxKey 4",
			args: args{
				idxKey: []byte("test-idx-key4"),
				idx:    1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetKeyByIdxKey with non-existing idxKey",
			args: args{
				idxKey: []byte("test-idx-key99"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test GetKeyByIdxKey with wrong idx",
			args: args{
				idxKey: []byte("test-idx-key"),
				idx:    10,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range testsGetKeyByIdxKey {
		t.Run(tt.name, func(t *testing.T) {
			data, err := dbBatch.GetKeyByIdxKey(tt.args.idxKey, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBBatch.GetKeyByIdxKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("LDBBatch.GetKeyByIdxKey(), got %q expected %q", string(data), string(tt.want))
			}
		})
	}

	testsDelete := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test Delete 0",
			args: args{
				idxKey: []byte("test-idx-key"),
				idx:    1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test Delete 1",
			args: args{
				idxKey: []byte("test-idx-key2"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test Delete 2",
			args: args{
				idxKey: []byte("test-idx-key3"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test Delete 3",
			args: args{
				idxKey: []byte("test-idx-key4"),
				idx:    0,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range testsDelete {
		t.Run(tt.name+" after delete", func(t *testing.T) {
			err1 := dbBatch.DeleteAll(tt.args.idxKey)
			if err1 != nil {
				t.Errorf("LDBBatch.DeleteAll(), err %q", err1)
			}

			_, err := dbBatch.GetByIdxKey(tt.args.idxKey, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBBatch.GetByIdxKey() after delete error = %v, wantErr %v", err, tt.wantErr)
			}

			_, err = dbBatch.GetKeyByIdxKey(tt.args.idxKey, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBBatch.GetKeyByIdxKey() after delete error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestLDBBatch_TryPutAll(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	dbBatch, _ := NewLDBBatch(tDefaultDB)

	updateTS1, _ := types.GetTimestamp()

	time.Sleep(1)

	updateTS2, _ := types.GetTimestamp()

	// define test-structure
	type fields struct {
		ldbBatch *ldbBatch
	}
	type args struct {
		idxKey       []byte
		idx          *Index
		kvs          []*KeyVal
		isDeleteOrig bool
		isGetOrig    bool
	}

	// prepare test-cases
	tests := []struct {
		name    string
		b       *LDBBatch
		args    args
		want    []*KeyVal
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test TryPutAll method with new key",
			b:    dbBatch,
			args: args{
				idxKey: []byte("test-idx-key"),
				idx: &Index{
					Keys:     [][]byte{[]byte("test-idx-1"), []byte("test-idx-2"), []byte("test-idx-3")},
					UpdateTS: updateTS1,
				},
				kvs: []*KeyVal{
					&KeyVal{K: []byte("test-idx-1"), V: []byte("test-value-1")},
					&KeyVal{K: []byte("test-idx-2"), V: []byte("test-value-2")},
					&KeyVal{K: []byte("test-idx-3"), V: []byte("test-value-3")},
				},
				isDeleteOrig: false,
				isGetOrig:    true,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test TryPutAll method with old key",
			b:    dbBatch,
			args: args{
				idxKey: []byte("test-idx-key"),
				idx: &Index{
					Keys:     [][]byte{[]byte("test-newidx-1"), []byte("test-newidx-2"), []byte("test-newidx-3")},
					UpdateTS: updateTS2,
				},
				kvs: []*KeyVal{
					&KeyVal{K: []byte("test-newidx-1"), V: []byte("test-value-1")},
					&KeyVal{K: []byte("test-newidx-2"), V: []byte("test-value-2")},
					&KeyVal{K: []byte("test-newidx-3"), V: []byte("test-value-3")},
				},
				isDeleteOrig: true,
				isGetOrig:    true,
			},
			want: []*KeyVal{
				&KeyVal{K: []byte("test-idx-1"), V: []byte("test-value-1")},
				&KeyVal{K: []byte("test-idx-2"), V: []byte("test-value-2")},
				&KeyVal{K: []byte("test-idx-3"), V: []byte("test-value-3")},
			},
			wantErr: false,
		},
		{
			name: "test TryPutAll method update old key with smaller TS",
			b:    dbBatch,
			args: args{
				idxKey: []byte("test-idx-key"),
				idx: &Index{
					Keys:     [][]byte{[]byte("test-newidx-5"), []byte("test-newidx-6"), []byte("test-newidx-7")},
					UpdateTS: updateTS1,
				},
				kvs: []*KeyVal{
					&KeyVal{K: []byte("test-newidx-5"), V: []byte("test-value-5")},
					&KeyVal{K: []byte("test-newidx-6"), V: []byte("test-value-6")},
					&KeyVal{K: []byte("test-newidx-7"), V: []byte("test-value-7")},
				},
				isDeleteOrig: true,
				isGetOrig:    true,
			},
			want: []*KeyVal{
				&KeyVal{K: []byte("test-newidx-1"), V: []byte("test-value-1")},
				&KeyVal{K: []byte("test-newidx-2"), V: []byte("test-value-2")},
				&KeyVal{K: []byte("test-newidx-3"), V: []byte("test-value-3")},
			},
			wantErr: true,
		},
		{
			name: "test TryPutAll method update old key with same TS but different idx",
			b:    dbBatch,
			args: args{
				idxKey: []byte("test-idx-key"),
				idx: &Index{
					Keys:     [][]byte{[]byte("test-newidx-8"), []byte("test-newidx-9")},
					UpdateTS: updateTS2,
				},
				kvs: []*KeyVal{
					&KeyVal{K: []byte("test-newidx-8"), V: []byte("test-value-8")},
					&KeyVal{K: []byte("test-newidx-9"), V: []byte("test-value-9")},
				},
				isDeleteOrig: true,
				isGetOrig:    true,
			},
			want: []*KeyVal{
				&KeyVal{K: []byte("test-newidx-1"), V: []byte("test-value-1")},
				&KeyVal{K: []byte("test-newidx-2"), V: []byte("test-value-2")},
				&KeyVal{K: []byte("test-newidx-3"), V: []byte("test-value-3")},
			},
			wantErr: false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.b
			got, err := b.TryPutAll(tt.args.idxKey, tt.args.idx, tt.args.kvs, tt.args.isDeleteOrig, tt.args.isGetOrig)
			if (err != nil) != tt.wantErr {
				t.Errorf("LDBBatch.TryPutAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDBBatch.TryPutAll() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestLDBBatch_TryPutAll2(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	dbBatch, _ := NewLDBBatch(tDefaultDB)

	updateTS1, _ := types.GetTimestamp()

	time.Sleep(1)

	updateTS2, _ := types.GetTimestamp()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		idxKey := []byte("test-idx-2-1")
		idx := &Index{
			Keys:     [][]byte{[]byte("test-newidx-2-8"), []byte("test-newidx-2-9")},
			UpdateTS: updateTS1,
		}

		kvs := []*KeyVal{
			&KeyVal{K: []byte("test-newidx-2-8"), V: []byte("test-value-2-8")},
			&KeyVal{K: []byte("test-newidx-2-9"), V: []byte("test-value-2-9")},
		}

		dbBatch.TryPutAll(idxKey, idx, kvs, true, false)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		idxKey := []byte("test-idx-2-2")
		idx := &Index{
			Keys:     [][]byte{[]byte("test-newidx-2-10"), []byte("test-newidx-2-11")},
			UpdateTS: updateTS2,
		}

		kvs := []*KeyVal{
			&KeyVal{K: []byte("test-newidx-2-10"), V: []byte("test-value-2-10")},
			&KeyVal{K: []byte("test-newidx-2-11"), V: []byte("test-value-2-11")},
		}

		dbBatch.TryPutAll(idxKey, idx, kvs, true, false)
	}()

	wg.Wait()

	val, err := dbBatch.DBGet([]byte("test-idx-2-1"))
	t.Logf("after DBGet: val: %v e: %v", val, err)
	assert.NoError(t, err)
	idx := &Index{}
	err = idx.Unmarshal(val)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(idx.Keys))
	assert.Equal(t, []byte("test-newidx-2-8"), idx.Keys[0])
	assert.Equal(t, []byte("test-newidx-2-9"), idx.Keys[1])

	val, err = dbBatch.DBGet([]byte("test-newidx-2-8"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-2-8"), val)

	val, err = dbBatch.DBGet([]byte("test-newidx-2-9"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-2-9"), val)

	val, err = dbBatch.DBGet([]byte("test-idx-2-2"))
	t.Logf("after DBGet: val: %v e: %v", val, err)
	assert.NoError(t, err)
	idx = &Index{}
	err = idx.Unmarshal(val)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(idx.Keys))
	assert.Equal(t, []byte("test-newidx-2-10"), idx.Keys[0])
	assert.Equal(t, []byte("test-newidx-2-11"), idx.Keys[1])

	val, err = dbBatch.DBGet([]byte("test-newidx-2-10"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-2-10"), val)

	val, err = dbBatch.DBGet([]byte("test-newidx-2-11"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-2-11"), val)
}

func TestLDBBatch_ForcePutAll(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	dbBatch, _ := NewLDBBatch(tDefaultDB)

	updateTS1, _ := types.GetTimestamp()

	time.Sleep(1)

	updateTS2, _ := types.GetTimestamp()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		idxKey := []byte("test-idx-3-1")
		idx := &Index{
			Keys:     [][]byte{[]byte("test-newidx-3-8"), []byte("test-newidx-3-9")},
			UpdateTS: updateTS1,
		}

		kvs := []*KeyVal{
			&KeyVal{K: []byte("test-newidx-3-8"), V: []byte("test-value-3-8")},
			&KeyVal{K: []byte("test-newidx-3-9"), V: []byte("test-value-3-9")},
		}

		dbBatch.ForcePutAll(idxKey, idx, kvs)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		idxKey := []byte("test-idx-3-2")
		idx := &Index{
			Keys:     [][]byte{[]byte("test-newidx-3-10"), []byte("test-newidx-3-11")},
			UpdateTS: updateTS2,
		}

		kvs := []*KeyVal{
			&KeyVal{K: []byte("test-newidx-3-10"), V: []byte("test-value-3-10")},
			&KeyVal{K: []byte("test-newidx-3-11"), V: []byte("test-value-3-11")},
		}

		dbBatch.ForcePutAll(idxKey, idx, kvs)
	}()

	wg.Wait()

	val, err := dbBatch.DBGet([]byte("test-idx-3-1"))
	t.Logf("after DBGet: val: %v e: %v", val, err)
	assert.NoError(t, err)
	idx := &Index{}
	err = idx.Unmarshal(val)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(idx.Keys))
	assert.Equal(t, []byte("test-newidx-3-8"), idx.Keys[0])
	assert.Equal(t, []byte("test-newidx-3-9"), idx.Keys[1])

	val, err = dbBatch.DBGet([]byte("test-newidx-3-8"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-3-8"), val)

	val, err = dbBatch.DBGet([]byte("test-newidx-3-9"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-3-9"), val)

	val, err = dbBatch.DBGet([]byte("test-idx-3-2"))
	t.Logf("after DBGet: val: %v e: %v", val, err)
	assert.NoError(t, err)
	idx = &Index{}
	err = idx.Unmarshal(val)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(idx.Keys))
	assert.Equal(t, []byte("test-newidx-3-10"), idx.Keys[0])
	assert.Equal(t, []byte("test-newidx-3-11"), idx.Keys[1])

	val, err = dbBatch.DBGet([]byte("test-newidx-3-10"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-3-10"), val)

	val, err = dbBatch.DBGet([]byte("test-newidx-3-11"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value-3-11"), val)
}
