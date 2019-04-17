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

package me

import "errors"

var (
	ErrInvalidMe = errors.New("invalid me")

	ErrInvalidNode = errors.New("invalid node")

	ErrInvalidPrivateKeyPostfix = errors.New("private-key and postfix must be specified at the same time")
	ErrInvalidPrivateKey        = errors.New("invalid private key")
	ErrInvalidPrivateKeyFileHex = errors.New("cannot set private-key file / hex at the same time")
	ErrInvalidPrivateKeyFile    = errors.New("invalid private key file")
	ErrInvalidPrivateKeyHex     = errors.New("invalid private key hex")

	ErrAlreadyMyNode = errors.New("already my node")

	ErrInvalidEntry     = errors.New("invalid raft entry")
	ErrInvalidRaftIndex = errors.New("invalid raft index")

	ErrUnableToBeLead = errors.New("unable to be lead")

	ErrWithLead = errors.New("with lead")
)
