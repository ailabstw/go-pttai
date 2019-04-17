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
	"errors"
	"fmt"
)

var (
	ErrAlreadyPrestarted = errors.New("already prestarted")
	ErrAlreadyStarted    = errors.New("already started")
	ErrToClose           = errors.New("peer is to close")
	ErrClosed            = errors.New("peer set is closed")
	ErrAlreadyRegistered = errors.New("peer is already registered")
	ErrNotRegistered     = errors.New("peer is not registered")
	ErrPeerUserID        = errors.New("peer user id")

	ErrEntityAlreadyRegistered = errors.New("entity is already registered")
	ErrEntityNotRegistered     = errors.New("entity is not registered")

	ErrInit         = errors.New("failed init")
	ErrQuota        = errors.New("size exceeds quota")
	ErrNegativeSize = errors.New("size < 0")
	ErrInvalidKey   = errors.New("invalid key")

	ErrMsgTooLarge = errors.New("msg too large")

	ErrFileTooLarge = errors.New("file too large")

	ErrInvalidMaster0 = errors.New("invalid master0")

	ErrServiceUnknown = errors.New("service unknown")

	ErrInvalidMsgCode = errors.New("invalid msg code")
	ErrInvalidMsg     = errors.New("invalid msg")

	ErrNotSent = errors.New("not sent")

	ErrInvalidData = errors.New("invalid data")

	ErrInvalidObject = errors.New("invalid object")

	ErrTimeout = errors.New("timeout")

	ErrInvalidEntity = errors.New("invalid entity")

	ErrInvalidOp          = errors.New("invalid op")
	ErrInvalidOplog       = errors.New("invalid oplog")
	ErrOplogAlreadyExists = errors.New("oplog already exists")
	ErrSkipOplog          = errors.New("skip oplog")
	ErrNewerOplog         = errors.New("newer oplog")
	ErrInvalidPreLog      = errors.New("invalid pre-log")

	ErrNoValidOplogs = errors.New("no valid oplogs")

	ErrInvalidKeyInfo = errors.New("invalid key info")

	ErrNotFound = errors.New("not found")

	ErrNoPeer = errors.New("no peer")

	ErrInvalidStatus = errors.New("invalid status")

	ErrBusy = errors.New("busy")

	ErrPeerRecentAdded = errors.New("peer recent added")

	ErrAlreadyMyNode = errors.New("already my node")

	ErrInvalidSyncInfo = errors.New("invalid sync info")

	ErrTooManyMasters = errors.New("too many masters")

	ErrInvalidBlock = errors.New("invalid block")

	ErrAlreadyPending = errors.New("already pending")

	ErrNotAlive = errors.New("not alive")

	ErrInvalidFunc = errors.New("invalid function")

	ErrInvalidMerkle = errors.New("invalid merkle")
)

func ErrResp(code error, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func errMapToErr(errMap map[string]error) error {
	if len(errMap) == 0 {
		return nil
	}

	var str string
	i := 0
	for kind, err := range errMap {
		if i != 0 {
			str += ","
		}
		str += fmt.Sprintf("%v: %v", kind, err)
	}

	return errors.New(str)
}
