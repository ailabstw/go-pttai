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
	"testing"
	"time"
)

func TestLock_Unlock(t *testing.T) {
	// setup test
	lock := NewLock()
	lock.TryLock()

	invalidLock := NewLock()

	lock2 := NewLock()
	lock2.TryLock()
	lock2.TryLock()
	lock2.TryLock()
	lock2.TryLock()
	lock2.Close()

	lock3 := NewLock()
	lock3.Close()

	// define test-structure
	type fields struct {
		lock chan struct{}
	}

	// prepare test-cases
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "unlock valid lock",
			fields: fields{lock: lock.lock},
		},
		{
			name:    "unlock invalid lock",
			fields:  fields{lock: invalidLock.lock},
			wantErr: true,
		},
		{
			name:    "unlock closed lock",
			fields:  fields{lock: lock2.lock},
			wantErr: false,
		},
		{
			name:    "unlock closed lock",
			fields:  fields{lock: lock3.lock},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				lock: tt.fields.lock,
			}
			if err := l.Unlock(); (err != nil) != tt.wantErr {
				t.Errorf("Lock.Unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestLock_TryLock(t *testing.T) {
	// setup test
	lock := NewLock()
	invalidLock := NewLock()
	invalidLock.TryLock()

	// define test-structure
	type fields struct {
		lock chan struct{}
	}

	// prepare test-cases
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{lock: lock.lock},
		},
		{
			fields:  fields{lock: invalidLock.lock},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				lock: tt.fields.lock,
			}
			if err := l.TryLock(); (err != nil) != tt.wantErr {
				t.Errorf("Lock.TryLock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestLock_TryTimedLock(t *testing.T) {
	// setup test
	lock := NewLock()
	lock.TryLock()

	lock2 := NewLock()
	lock2.TryLock()

	go func() {
		time.Sleep(10 * time.Millisecond)
		lock.Unlock()
	}()

	go func() {
		time.Sleep(17 * time.Millisecond) // wait 10 secs for the 1st lock
		t.Log("to unlock lock2")
		lock2.Unlock()
	}()

	// define test-structure
	type fields struct {
		lock chan struct{}
	}
	type args struct {
		nMillisecond time.Duration
	}

	// prepare test-cases
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{lock: lock.lock},
			args:   args{20},
		},
		{
			fields:  fields{lock: lock2.lock},
			args:    args{5},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				lock: tt.fields.lock,
			}
			t.Log("tt", tt)
			if err := l.TryTimedLock(tt.args.nMillisecond); (err != nil) != tt.wantErr {
				t.Errorf("Lock.TryTimedLock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestLock_TimedUnlock(t *testing.T) {
	// setup test
	lock := NewLock()

	go func() {
		time.Sleep(10 * time.Millisecond)
		lock.TryLock()
	}()

	lock2 := NewLock()

	go func() {
		time.Sleep(15 * time.Millisecond)
		lock.TryLock()
	}()

	// define test-structure
	type fields struct {
		lock chan struct{}
	}
	type args struct {
		nMillisecond time.Duration
	}

	// prepare test-cases
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{lock: lock.lock},
			args:   args{nMillisecond: 15},
		},
		{
			fields:  fields{lock: lock2.lock},
			args:    args{nMillisecond: 3},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				lock: tt.fields.lock,
			}
			if err := l.TimedUnlock(tt.args.nMillisecond); (err != nil) != tt.wantErr {
				t.Errorf("Lock.TimedUnlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// teardown test
}

func TestLock_IsLocked(t *testing.T) {
	// setup test
	lock := NewLock()

	lock2 := NewLock()
	lock2.TryLock()

	lock3 := NewLock()
	lock3.TryLock()
	lock3.Close()

	lock4 := NewLock()
	lock4.Close()

	t.Log("lock3:", len(lock3.lock))

	// define test-structure
	type fields struct {
		lock chan struct{}
	}

	// prepare test-cases
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{lock: lock.lock},
			want:   false,
		},
		{
			fields: fields{lock: lock2.lock},
			want:   true,
		},
		{
			fields: fields{lock: lock3.lock},
			want:   true,
		},
		{
			fields: fields{lock: lock4.lock},
			want:   false,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				lock: tt.fields.lock,
			}
			if got := l.IsLocked(); got != tt.want {
				t.Errorf("Lock.IsLocked() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
