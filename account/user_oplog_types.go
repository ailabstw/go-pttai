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

package account

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

const (
	UserOpTypeInvalid pkgservice.OpType = iota
	UserOpTypeCreateProfile
	UserOpTypeDeleteProfile
	UserOpTypeTransferProfile

	UserOpTypeAddUserNode
	UserOpTypeRemoveUserNode

	UserOpTypeCreateUserName
	UserOpTypeUpdateUserName

	UserOpTypeCreateUserImg
	UserOpTypeUpdateUserImg

	UserOpTypeCreateNameCard
	UserOpTypeUpdateNameCard

	NUserOpType
)

type UserOpCreateProfile struct {
}

type UserOpDeleteProfile struct {
}

type UserOpTransferProfile struct {
	ToID *types.PttID `json:"t"`
}

type UserOpAddUserNode struct {
	NodeID *discover.NodeID `json:"n"`
}

type UserOpRemoveUserNode struct {
	NodeID *discover.NodeID `json:"n"`
}

type UserOpCreateUserName struct {
}

type UserOpUpdateUserName struct {
	Hash []byte `json:"H"`
}

type UserOpCreateUserImg struct {
}

type UserOpUpdateUserImg struct {
	Hash []byte `json:"H"`
}

type UserOpCreateNameCard struct {
}

type UserOpUpdateNameCard struct {
	Hash []byte `json:"H"`
}
