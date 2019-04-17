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

// referring: https://gocn.vip/question/176

package types

import "time"

type Lock struct {
	isClosed bool
	lock     chan struct{}
}

func NewLock() *Lock {
	lock := &Lock{
		lock: make(chan struct{}, 1),
	}

	return lock
}

func (l *Lock) TryLock() error {
	if l.isClosed {
		return ErrLockClosed
	}
	select {
	case l.lock <- struct{}{}:
		return nil
	default:
		return ErrLock
	}
}

func (l *Lock) Unlock() error {
	select {
	case _, isOK := <-l.lock:
		if !isOK {
			return ErrLockClosed
		}
		return nil
	default:
		return ErrUnlock
	}
}

func (l *Lock) TryTimedLock(nMillisecond time.Duration) error {
	if l.isClosed {
		return ErrLockClosed
	}

	timer := time.NewTimer(nMillisecond * time.Millisecond)
	defer timer.Stop()

	select {
	case l.lock <- struct{}{}:
		return nil
	case <-timer.C:
		return ErrLock
	}
}

func (l *Lock) TimedUnlock(nMillisecond time.Duration) error {
	timer := time.NewTimer(nMillisecond * time.Millisecond)
	defer timer.Stop()

	select {
	case _, isOK := <-l.lock:
		if !isOK {
			return ErrLockClosed
		}
		return nil
	case <-timer.C:
		return ErrUnlock
	}
}

func (l *Lock) Close() error {
	if l.IsLocked() {
		return ErrClose
	}
	l.isClosed = true
	close(l.lock)
	return nil
}

// XXX possibly unable to get accurate information
func (l *Lock) IsLocked() bool {
	return len(l.lock) > 0
}
