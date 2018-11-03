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

type SyncInfo interface {
	GetLogID() *types.PttID
	SetLogID(id *types.PttID)

	GetUpdateTS() types.Timestamp
	SetUpdateTS(ts types.Timestamp)

	GetStatus() types.Status
	SetStatus(status types.Status)
}

type BaseSyncInfo struct {
	LogID     *types.PttID    `json:"l"`
	UpdateTS  types.Timestamp `json:"UT"`
	UpdaterID *types.PttID    `json:"UID"`
	Status    types.Status    `json:"S"`
}

func (s *BaseSyncInfo) SetLogID(id *types.PttID) {
	s.LogID = id
}

func (s *BaseSyncInfo) GetLogID() *types.PttID {
	return s.LogID
}

func (s *BaseSyncInfo) GetUpdateTS() types.Timestamp {
	return s.UpdateTS
}

func (s *BaseSyncInfo) SetUpdateTS(ts types.Timestamp) {
	s.UpdateTS = ts
}

func (s *BaseSyncInfo) GetStatus() types.Status {
	return s.Status
}

func (s *BaseSyncInfo) SetStatus(status types.Status) {
	s.Status = status
}

func (s *BaseSyncInfo) InitWithOplog(oplog *BaseOplog) {
	s.LogID = oplog.ID
	s.UpdateTS = oplog.UpdateTS
	s.UpdaterID = oplog.CreatorID
	s.Status = oplog.ToStatus()
}
