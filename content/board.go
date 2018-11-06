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
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncTitleInfo struct {
	*pkgservice.BaseSyncInfo `json:"b"`

	Title []byte `json:"T"`
}

type Board struct {
	*pkgservice.BaseEntity `json:"e"`

	UpdateTS types.Timestamp `json:"UT"`

	Title         []byte         `json:"T,omitempty"`
	SyncTitleInfo *SyncTitleInfo `json:"st,omitempty"`

	// get from other dbs
	LastSeen        types.Timestamp `json:"-"`
	ArticleCreateTS types.Timestamp `json:"-"`

	BoardMerkle *pkgservice.Merkle `json:"-"`
}

func NewEmptyBoard() *Board {
	return &Board{BaseEntity: &pkgservice.BaseEntity{}}
}

func (b *Board) GetUpdateTS() types.Timestamp {
	return b.UpdateTS
}

func (b *Board) SetUpdateTS(ts types.Timestamp) {
	b.UpdateTS = ts
}

func (b *Board) Save(isLocked bool) error {
	return types.ErrNotImplemented
}

func (b *Board) Init(ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager) error {
	return types.ErrNotImplemented
}
