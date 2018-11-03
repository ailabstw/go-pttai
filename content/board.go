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

package content

import (
	"sync"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncTitleInfo struct {
	LogID    *types.PttID    `json:"pl"`
	Title    []byte          `json:"T"`
	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`
}

type Board struct {
	*pkgservice.BaseEntity `json:"b"`

	V         types.Version
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	UpdateTS  types.Timestamp `json:"UT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`

	Status types.Status `json:"S"`

	Title         []byte         `json:"T"`
	TitleLogID    *types.PttID   `json:"TID,omitempty"`
	SyncTitleInfo *SyncTitleInfo `json:"st,omitempty"`

	BoardType BoardType `json:"BT"`

	// get from other dbs
	LastSeen        types.Timestamp `json:"-"`
	ArticleCreateTS types.Timestamp `json:"-"`

	lockMaster sync.RWMutex
	lockMember sync.RWMutex

	Masters map[types.PttID]*Member `json:"-"`

	MasterMerkle  *pkgservice.Merkle `json:"-"`
	MemberMerkle  *pkgservice.Merkle `json:"-"`
	BoardMerkle   *pkgservice.Merkle `json:"-"`
	CommentMerkle *pkgservice.Merkle `json:"-"`
	NodeMerkle    *pkgservice.Merkle `json:"_"`

	LogID *types.PttID `json:"l"`

	dbLock *types.LockMap
}
