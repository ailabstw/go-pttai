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
	"github.com/ailabstw/go-pttai/pttdb"
)

type SyncMediaInfo struct {
	LogID *types.PttID `json:"pl"`

	UpdateTS  types.Timestamp `json:"UT"`
	UpdaterID *types.PttID    `json:"UID"`
	Status    types.Status    `json:"S"`
}

type Media struct {
	V         types.Version
	ID        *types.PttID
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`
	CreateTS  types.Timestamp `json:"CT"`
	UpdateTS  types.Timestamp `json:"UT"`

	Status types.Status `json:"S"`

	BoardID *types.PttID `json:"BID"`

	OrigContentBlockID *types.PttID `json:"obID"`
	OrigNBlock         int          `json:"oNB"`
	OrigMediaType      MediaType    `json:"oT"`
	OrigData           interface{}  `json:"oD"`

	ContentBlockID *types.PttID `json:"bID"`
	NBlock         int          `json:"NB"`
	MediaType      MediaType    `json:"T"`
	Data           interface{}  `json:"D"`

	Buf []byte `json:"-"` // from content-blocks

	LastSeen types.Timestamp `json:"L"`

	LogID  *types.PttID `json:"l"`
	IsSync types.Bool   `json:"y"`

	SyncMediaInfo *SyncMediaInfo `json:"s,omitempty"`

	db *pttdb.LDBBatch

	dbLock *types.LockMap
}
