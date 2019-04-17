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

import "errors"

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
	ErrInvalidBitVector = errors.New("invalid bit vector")

	ErrLock        = errors.New("unable to lock")
	ErrUnlock      = errors.New("unable to unlock")
	ErrClose       = errors.New("unable to close")
	ErrLockClosed  = errors.New("lock closed")
	ErrBusy        = errors.New("busy")
	ErrInvalidLock = errors.New("invalid lock")

	ErrInvalidURL = errors.New("invalid url")

	ErrAlreadyExists = errors.New("already exists")

	ErrAlreadyDeleted = errors.New("already deleted")

	ErrAlreadyPending = errors.New("already pending")

	ErrAlreadyPendingDelete = errors.New("already pending delete")

	ErrInvalidNode = errors.New("invalid node")

	ErrNotImplemented = errors.New("not implemented")

	ErrNotBool = errors.New("not bool")

	ErrNotBoolDAry = errors.New("not bool double array")

	ErrInvalidStatus = errors.New("invalid status")
)
