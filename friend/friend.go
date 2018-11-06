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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Friend struct {
	*pkgservice.BaseEntity `json:"e"`

	UpdateTS types.Timestamp `json:"UT"`

	SyncInfo *pkgservice.SyncInfo `json:"st,omitempty"`

	// get from other dbs
	LastSeen        types.Timestamp `json:"-"`
	ArticleCreateTS types.Timestamp `json:"-"`

	FriendMerkle *pkgservice.Merkle `json:"-"`
}

func NewEmptyFriend() *Friend {
	return &Friend{BaseEntity: &pkgservice.BaseEntity{}}
}

func NewFriend() (*Friend, error) {
	return &Friend{}, nil
}

func (f *Friend) GetUpdateTS() types.Timestamp {
	return f.UpdateTS
}

func (f *Friend) SetUpdateTS(ts types.Timestamp) {
	f.UpdateTS = ts
}

func (f *Friend) Save(isLocked bool) error {
	return types.ErrNotImplemented
}

func (f *Friend) Init(ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager) error {
	return types.ErrNotImplemented
}
