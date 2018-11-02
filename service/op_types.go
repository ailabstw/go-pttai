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

package service

import "encoding/binary"

// CodeType
type CodeType uint64

const (
	CodeTypeInvalid CodeType = iota
	CodeTypeStatus
	CodeTypeJoin
	CodeTypeJoinAck
	CodeTypeOp
	CodeTypeOpAck
	CodeTypeOpFail
	CodeTypeRequestOp
	CodeTypeRequestOpAck

	CodeTypeIdentifyPeer
	CodeTypeIdentifyPeerAck
	CodeTypeIdentifyPeerFail

	CodeTypeIdentifyPeerWithMyID
	CodeTypeIdentifyPeerWithMyIDChallenge
	CodeTypeIdentifyPeerWithMyIDChallengeAck
	CodeTypeIdentifyPeerWithMyIDAck

	NCodeType
)

func MarshalCode(code CodeType) ([]byte, error) {
	codeBytes := make([]byte, SizeCodeType)
	binary.BigEndian.PutUint64(codeBytes, uint64(code))

	return codeBytes, nil

}

func UnmarshalCode(codeBytes []byte) (CodeType, error) {
	if len(codeBytes) != SizeCodeType {
		return 0, ErrInvalidOp
	}
	code := CodeType(binary.BigEndian.Uint64(codeBytes))
	return code, nil
}

// OpClass

type OpClass int

const (
	OpClassInvalid OpClass = iota
	OpClassCreate
	OpClassUpdate
	OpClassDelete
	OpClassOther
)

// OpType
type OpType uint32

const (
	ZeroOpType OpType = 0
	MaxOpType  OpType = 0xffffffff
)

func MarshalOp(op OpType) ([]byte, error) {
	opBytes := make([]byte, SizeOpType)
	binary.BigEndian.PutUint32(opBytes, uint32(op))

	return opBytes, nil

}

func UnmarshalOp(opBytes []byte) (OpType, error) {
	if len(opBytes) != SizeOpType {
		return 0, ErrInvalidOp
	}
	op := OpType(binary.BigEndian.Uint32(opBytes))
	return op, nil
}

// OpData
type OpData interface{}
