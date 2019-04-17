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

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

const (
	_ pkgservice.OpType = iota
	MasterOpTypeAddMaster
	MasterOpTypeRevokeMaster
)

type MasterOpAddMaster struct {
	ID      *discover.NodeID
	From    *types.PttID               `json:"f"`
	Masters map[discover.NodeID]uint32 `json:"M"`
	Weight  uint32                     `json:"W"`
}

type MasterOpRevokeMaster struct {
	ID      *discover.NodeID
	From    *types.PttID               `json:"f"`
	Masters map[discover.NodeID]uint32 `json:"M"`
}
