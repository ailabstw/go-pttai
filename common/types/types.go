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

// Status
type Status int

const (
	StatusInvalid Status = iota

	StatusInit

	StatusInternalSync

	StatusInternalPending

	StatusPending

	StatusSync

	StatusToBeSynced

	StatusAlive

	StatusFailed

	// Putting intenal-deleted after alive.
	// Because it's the competition between update-object and pending-delete, which does not affect the new-object, and no dead-lock for pending-delete (referring to new-object).
	StatusInternalDeleted

	StatusInternalRevoke

	StatusInternalTransfer

	StatusInternalMigrate

	StatusInternalTerminal

	StatusPendingDeleted

	StatusPendingRevoke

	StatusPendingTransfer

	StatusPendingMigrate

	StatusPendingTerminal

	StatusDeleted

	StatusRevoked

	StatusTransferred

	StatusMigrated

	StatusTerminal
)

var (
	statusStr = map[Status]string{
		StatusInvalid:          "invalid",
		StatusInit:             "init",
		StatusInternalSync:     "internal-sync",
		StatusInternalPending:  "internal-pending",
		StatusPending:          "pending",
		StatusSync:             "sync",
		StatusToBeSynced:       "be-synced",
		StatusAlive:            "alive",
		StatusFailed:           "failed",
		StatusInternalDeleted:  "internal-delete",
		StatusInternalRevoke:   "internal-revoke",
		StatusInternalTransfer: "internal-transfer",
		StatusInternalMigrate:  "internal-migrate",
		StatusInternalTerminal: "internal-terminal",
		StatusPendingDeleted:   "pending-delete",
		StatusPendingRevoke:    "pending-revoke",
		StatusPendingTransfer:  "pending-transfer",
		StatusPendingMigrate:   "pending-migrate",
		StatusPendingTerminal:  "pending-terminal",
		StatusDeleted:          "deleted",
		StatusRevoked:          "revoked",
		StatusTransferred:      "transferred",
		StatusMigrated:         "migrated",
		StatusTerminal:         "terminal",
	}
)

func (s Status) String() string {
	return statusStr[s]
}

func StatusToDeleteStatus(status Status, internalPendingStatus Status, pendingStatus Status, aliveStatus Status) Status {
	switch status {
	case StatusInternalPending:
		return internalPendingStatus
	case StatusPending:
		return pendingStatus
	case StatusAlive:
		return aliveStatus
	}

	return status
}

type StatusClass int

const (
	StatusClassInvalid StatusClass = iota
	StatusClassInternalPendingAlive
	StatusClassPendingAlive
	StatusClassAlive
	StatusClassFailed
	StatusClassInternalDelete
	StatusClassPendingDelete
	StatusClassInternalTerminal
	StatusClassPendingTerminal
	StatusClassDeleted
	StatusClassTerminal
)

var (
	statusClassStr = map[StatusClass]string{
		StatusClassInvalid: "class-invalid",

		StatusClassInternalPendingAlive: "class-internal-pending",

		StatusClassPendingAlive:     "class-pending",
		StatusClassAlive:            "class-alive",
		StatusClassFailed:           "class-failed",
		StatusClassInternalDelete:   "class-internal-delete",
		StatusClassPendingDelete:    "class-pending-delete",
		StatusClassDeleted:          "class-deleted",
		StatusClassInternalTerminal: "class-internal-migrate",
		StatusClassPendingTerminal:  "class-pending-migrate",
		StatusClassTerminal:         "class-migrated",
	}
)

func (s StatusClass) String() string {
	return statusClassStr[s]
}

var statusToStatusClass = map[Status]StatusClass{
	StatusInvalid: StatusClassInvalid,

	StatusInit:            StatusClassInternalPendingAlive,
	StatusInternalSync:    StatusClassInternalPendingAlive,
	StatusInternalPending: StatusClassInternalPendingAlive,
	StatusPending:         StatusClassPendingAlive,
	StatusSync:            StatusClassPendingAlive,

	StatusToBeSynced: StatusClassAlive,

	StatusAlive: StatusClassAlive,

	StatusFailed: StatusClassFailed,

	StatusInternalDeleted:  StatusClassInternalDelete,
	StatusInternalRevoke:   StatusClassInternalDelete,
	StatusInternalTransfer: StatusClassInternalDelete, // transfer and migrate is treated the same as delete in pending mode.
	StatusInternalMigrate:  StatusClassInternalTerminal,
	StatusInternalTerminal: StatusClassInternalTerminal,

	StatusPendingDeleted:  StatusClassPendingDelete,
	StatusPendingRevoke:   StatusClassPendingDelete,
	StatusPendingTransfer: StatusClassPendingDelete,
	StatusPendingMigrate:  StatusClassPendingTerminal,
	StatusPendingTerminal: StatusClassPendingTerminal,

	StatusDeleted:     StatusClassDeleted,
	StatusRevoked:     StatusClassDeleted,
	StatusTransferred: StatusClassDeleted,
	StatusMigrated:    StatusClassTerminal,
	StatusTerminal:    StatusClassTerminal,
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
