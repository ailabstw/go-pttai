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

// Status
type Status int

const (
	StatusInvalid Status = iota

	StatusInit

	StatusInternalSync

	StatusInternalPending

	StatusPending

	StatusSync

	StatusAlive

	StatusFailed

	StatusPendingTransfer

	StatusTransferred

	StatusPendingMigrate

	StatusMigrated

	StatusPendingRevoke

	StatusRevoked

	StatusInternalDeleted

	StatusPendingDeleted

	StatusDeleted
)

func StatusToDeleteStatus(status Status) Status {
	switch status {
	case StatusInternalPending:
		return StatusInternalDeleted
	case StatusPending:
		return StatusPendingDeleted
	case StatusAlive:
		return StatusDeleted
	}

	return status
}

type StatusClass int

const (
	StatusClassInvalid StatusClass = iota
	StatusClassPendingAlive
	StatusClassAlive
	StatusClassFailed
	StatusClassPendingDelete
	StatusClassDeleted
)

var statusToStatusClass = map[Status]StatusClass{
	StatusInvalid: StatusClassInvalid,

	StatusInit:            StatusClassPendingAlive,
	StatusInternalSync:    StatusClassPendingAlive,
	StatusInternalPending: StatusClassPendingAlive,
	StatusPending:         StatusClassPendingAlive,
	StatusSync:            StatusClassPendingAlive,

	StatusAlive: StatusClassAlive,

	StatusFailed: StatusClassFailed,

	StatusPendingTransfer: StatusClassPendingDelete,
	StatusTransferred:     StatusClassDeleted,
	StatusPendingMigrate:  StatusClassPendingDelete,
	StatusMigrated:        StatusClassDeleted,
	StatusPendingRevoke:   StatusClassPendingDelete,
	StatusRevoked:         StatusClassDeleted,
	StatusInternalDeleted: StatusClassPendingDelete,
	StatusPendingDeleted:  StatusClassDeleted,
	StatusDeleted:         StatusClassDeleted,
}

func StatusToStatusClass(status Status) StatusClass {
	return statusToStatusClass[status]
}

// Sig
type Sig []byte

// image-type
type ImgType uint8

const (
	ImgTypeJPEG ImgType = iota
	ImgTypeGIF
	ImgTypePNG
)

// RaftStatus
type RaftState int

const (
	_ RaftState = iota
	RaftStateFollower
	RaftStateCandidate
	RaftStateLeader
)
