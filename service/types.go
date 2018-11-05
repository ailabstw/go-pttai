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

import "github.com/ailabstw/go-pttai/common/types"

// ServiceConstructor is the function signature of the constructors needed to be
// registered for service instantiation.
type ServiceConstructor func(ctx *ServiceContext) (PttService, error)

// merkletree

type MerkleTreeLevel uint8

const (
	_ MerkleTreeLevel = iota
	MerkleTreeLevelNow
	MerkleTreeLevelHR
	MerkleTreeLevelDay
	MerkleTreeLevelMonth
	MerkleTreeLevelYear
)

// PeerType

type PeerType int

const (
	PeerTypeErr PeerType = iota
	PeerTypeRemoved
	PeerTypeRandom
	PeerTypePending
	PeerTypeMember
	PeerTypeImportant
	PeerTypeHub
	PeerTypeMe
	NPeerType
)

// NodeType
type NodeType int

const (
	NodeTypeUnknown NodeType = iota
	NodeTypeMobile
	NodeTypeDesktop
	NodeTypeServer
)

// SignInfo

type SignInfo struct {
	ID       *types.PttID    `json:"ID"`
	CreateTS types.Timestamp `json:"CT"`

	Hash   []byte        `json:"H"`
	Salt   types.Salt    `json:"s"`
	Sig    []byte        `json:"S"`
	Pubkey []byte        `json:"K"`
	Extra  *KeyExtraInfo `json:"e,omitempty"`
}

type SyncID struct {
	ID    *types.PttID
	LogID *types.PttID `json:"l"`
}
