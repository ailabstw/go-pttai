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

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type LRUCount struct {
	*Count   `json:"C"`
	MaxCount uint64     `json:"M"`
	Full     types.Bool `json:"f"`
}

func NewLRUCount(maxCount uint64, db *pttdb.LDBBatch, dbPrefixID *types.PttID, dbID *types.PttID, dbPrefix []byte, p uint, isNewBits bool) (*LRUCount, error) {
	count, err := NewCount(db, dbPrefixID, dbID, dbPrefix, p, isNewBits)
	if err != nil {
		return nil, err
	}
	return &LRUCount{
		Count:    count,
		MaxCount: maxCount,
	}, nil
}

func (l *LRUCount) IsFull() bool {
	if l.Full {
		return true
	}

	count := l.Count.Count()
	if count >= l.MaxCount {
		l.Full = true
		l.Save()

		return true
	}

	return false
}
