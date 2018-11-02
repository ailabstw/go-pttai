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

type SyncArticleInfo struct {
	LogID          *types.PttID    `json:"pl"`
	ContentBlockID *types.PttID    `json:"bID,omitempty"`
	NBlock         int             `json:"NB"`
	NLine          int             `json:"NL"`
	MediaIDs       []*types.PttID  `json:"ms,omitempty"`
	UpdateTS       types.Timestamp `json:"UT"`
	UpdaterID      *types.PttID    `json:"UID"`
	Status         types.Status    `json:"S"`
}

type Article struct {
	V           types.Version
	ID          *types.PttID
	CreateTS    types.Timestamp `json:"CT"`
	UpdateTS    types.Timestamp `json:"UT"`
	CreatorID   *types.PttID    `json:"CID"`
	UpdaterID   *types.PttID    `json:"UID"`
	UpdateLogID *types.PttID    `json:"ul,omitempty"`

	Status types.Status `json:"S"`

	BoardID *types.PttID `json:"BID"`

	Title []byte `json:"T,omitempty"`

	ContentBlockID *types.PttID `json:"bID"`
	NBlock         int          `json:"NB"`
	NLine          int          `json:"NL"`

	NPush *pkgservice.Count `json:"-"` // from other db-records
	NBoo  *pkgservice.Count `json:"-"` // from other db-records

	CommentCreateTS types.Timestamp `json:"-"` // from other db-records
	LastSeen        types.Timestamp `json:"-"` // from other db-records

	MediaIDs []*types.PttID `json:"ms"`

	LogID           *types.PttID     `json:"l"`
	SyncArticleInfo *SyncArticleInfo `json:"s,omitempty"`

	dbLock *types.LockMap

	//	IsSync          types.Bool       `json:"y"`
}
