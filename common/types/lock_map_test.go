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

package types

import (
	"sync"
	"testing"
)

func TestLockMap_TryLock(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// setup test
	lock, _ := NewLockMap(10)
	invalidLock, _ := NewLockMap(10)
	id, _ := NewPttID()
	invalidLock.TryLock(id)

	// define test-structure
	type fields struct {
		lock      sync.Mutex
		lockMap   map[PttID]int
		sleepTime int
	}
	type args struct {
		id *PttID
	}

	// prepare test-cases
	tests := []struct {
		name    string
		lock    *LockMap
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			lock: lock,
			args: args{id},
		},
		{
			lock:    invalidLock,
			args:    args{id},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.lock
			if err := l.TryLock(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("LockMap.TryLock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestLockMap_Unlock(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	id, _ := NewPttID()

	lock, _ := NewLockMap(10)
	lock.TryLock(id)

	invalidLock, _ := NewLockMap(10)

	lock2, _ := NewLockMap(10)
	lock2.TryLock(id)
	lock2.TryLock(id)
	lock2.TryLock(id)
	lock2.TryLock(id)

	lock3, _ := NewLockMap(10)

	// define test-structure
	type fields struct {
		lock      sync.Mutex
		lockMap   map[PttID]int
		sleepTime int
	}
	type args struct {
		id *PttID
	}

	// prepare test-cases
	tests := []struct {
		name    string
		lock    *LockMap
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "unlock valid lock",
			lock: lock,
			args: args{id: id},
		},
		{
			name:    "unlock invalid lock",
			lock:    invalidLock,
			args:    args{id: id},
			wantErr: true,
		},
		{
			name:    "unlock closed lock",
			lock:    lock2,
			args:    args{id: id},
			wantErr: false,
		},
		{
			name:    "unlock closed lock",
			lock:    lock3,
			args:    args{id: id},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.lock
			if err := l.Unlock(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("LockMap.Unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestLockMap_Lock(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// setup test
	lock, _ := NewLockMap(5)
	invalidLock, _ := NewLockMap(5)
	id, _ := NewPttID()
	invalidLock.TryLock(id)

	// define test-structure
	type fields struct {
		lock      sync.Mutex
		lockMap   map[PttID]int
		sleepTime int
	}
	type args struct {
		id *PttID
	}

	// prepare test-cases
	tests := []struct {
		name    string
		lock    *LockMap
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			lock: lock,
			args: args{id},
		},
		{
			lock:    invalidLock,
			args:    args{id},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.lock
			if err := l.Lock(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("LockMap.Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}
