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

type BackendMyInfo struct {
	V        types.Version
	ID       *types.PttID
	CreateTS types.Timestamp `json:"CT"`
	UpdateTS types.Timestamp `json:"UT"`

	Status types.Status `json:"S"`

	RaftID uint64
	NodeID *discover.NodeID
}

func MarshalBackendMyInfo(m *MyInfo, ptt pkgservice.MyPtt) *BackendMyInfo {
	if m == nil {
		return nil
	}

	myRaftID := ptt.MyRaftID()
	myNodeID := ptt.MyNodeID()

	return &BackendMyInfo{
		V:        m.V,
		ID:       m.ID,
		CreateTS: m.CreateTS,
		UpdateTS: m.UpdateTS,
		Status:   m.Status,

		RaftID: myRaftID,
		NodeID: myNodeID,
	}
}
