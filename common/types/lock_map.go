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

package types

import (
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/ailabstw/go-pttai/log"
)

/*
LockMap implements per-ptt-id lock
*/
type LockMap struct {
	lock      sync.Mutex
	lockMap   map[PttID]int
	sleepTime int
	isSet     bool
}

func NewLockMap(sleepTime int) (*LockMap, error) {
	return &LockMap{
		lockMap:   make(map[PttID]int),
		sleepTime: sleepTime,
		isSet:     true,
	}, nil
}

func (l *LockMap) TryLock(id *PttID) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, ok := l.lockMap[*id]
	if ok {
		return ErrBusy
	}
	l.lockMap[*id] = -1

	return nil
}

func (l *LockMap) Unlock(id *PttID) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	val, ok := l.lockMap[*id]
	if !ok || val >= 0 {
		return ErrInvalidLock
	}
	delete(l.lockMap, *id)

	return nil
}

func (l *LockMap) TryRLock(id *PttID) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	val, ok := l.lockMap[*id]
	if ok && val < 0 {
		return ErrBusy
	}

	if !ok {
		l.lockMap[*id] = 0
	}
	l.lockMap[*id]++

	return nil
}

func (l *LockMap) RUnlock(id *PttID) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	val, ok := l.lockMap[*id]
	if !ok || val <= 0 {
		return ErrInvalidLock
	}

	l.lockMap[*id]--

	if val == 1 {
		delete(l.lockMap, *id)
	}

	return nil
}

func (l *LockMap) Lock(id *PttID) error {
	sleepTime := time.Duration((rand.Intn(l.sleepTime) + 1)) * time.Millisecond
	var err error
	for i := 0; i < NIterLock; i++ {
		err = l.TryLock(id)
		if err == nil {
			return nil
		}
		log.Warn("Lock: to sleep", "sleepTime", sleepTime)
		debug.PrintStack()
		time.Sleep(sleepTime)
	}

	return err
}

func (l *LockMap) MustLock(id *PttID) (err error) {
	for {
		err = l.Lock(id)
		if err == nil {
			return
		}
	}
}

func (l *LockMap) RLock(id *PttID) error {
	sleepTime := time.Duration((rand.Intn(l.sleepTime) + 1)) * time.Millisecond
	var err error
	for i := 0; i < NIterLock; i++ {
		err = l.TryRLock(id)
		if err == nil {
			return nil
		}
		time.Sleep(sleepTime)
	}

	return err
}
