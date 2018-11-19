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
)

type SyncInfo interface {
	GetLogID() *types.PttID
	SetLogID(id *types.PttID)

	GetUpdateTS() types.Timestamp
	SetUpdateTS(ts types.Timestamp)

	GetUpdaterID() *types.PttID
	SetUpdaterID(id *types.PttID)

	GetStatus() types.Status
	SetStatus(status types.Status)

	SetBlockInfo(blockInfo *BlockInfo) error
	GetBlockInfo() *BlockInfo

	SetIsGood(isGood types.Bool)
	GetIsGood() types.Bool

	SetIsAllGood(isAllGood types.Bool)
	GetIsAllGood() types.Bool
	CheckIsAllGood() types.Bool

	FromOplog(status types.Status, oplog *BaseOplog, opData OpData) error
	ToObject(obj Object) error
}

type BaseSyncInfo struct {
	LogID     *types.PttID    `json:"l"`
	UpdateTS  types.Timestamp `json:"UT"`
	UpdaterID *types.PttID    `json:"UID"`
	Status    types.Status    `json:"S"`

	BlockInfo *BlockInfo `json:"b,omitempty"`

	IsGood    types.Bool `json:"g"`
	IsAllGood types.Bool `json:"a"`
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

func (s *BaseSyncInfo) SetUpdaterID(id *types.PttID) {
	s.UpdaterID = id
}

func (s *BaseSyncInfo) GetUpdaterID() *types.PttID {
	return s.UpdaterID
}

func (s *BaseSyncInfo) GetStatus() types.Status {
	return s.Status
}

func (s *BaseSyncInfo) SetStatus(status types.Status) {
	s.Status = status
}

func (s *BaseSyncInfo) InitWithOplog(status types.Status, oplog *BaseOplog) {
	s.LogID = oplog.ID
	s.UpdateTS = oplog.UpdateTS
	s.UpdaterID = oplog.CreatorID
	s.Status = status
}

func (s *BaseSyncInfo) SetBlockInfo(blockInfo *BlockInfo) error {
	s.BlockInfo = blockInfo
	return nil
}

func (s *BaseSyncInfo) GetBlockInfo() *BlockInfo {
	return s.BlockInfo
}

func (s *BaseSyncInfo) SetIsGood(isGood types.Bool) {
	s.IsGood = isGood
}

func (s *BaseSyncInfo) GetIsGood() types.Bool {
	return s.IsGood
}

func (s *BaseSyncInfo) SetIsAllGood(isAllGood types.Bool) {
	s.IsAllGood = isAllGood
}

func (s *BaseSyncInfo) GetIsAllGood() types.Bool {
	return s.IsAllGood
}

func (s *BaseSyncInfo) CheckIsAllGood() types.Bool {
	if s.IsAllGood {
		return true
	}

	if !s.IsGood {
		return false
	}

	if s.BlockInfo != nil && !s.BlockInfo.GetIsAllGood() {
		return false
	}

	s.IsAllGood = true
	return true
}

func (s *BaseSyncInfo) ToObject(obj Object) error {
	obj.SetStatus(s.Status)
	obj.SetUpdateTS(s.UpdateTS)
	obj.SetUpdaterID(s.UpdaterID)
	obj.SetBlockInfo(s.BlockInfo)
	obj.SetIsGood(true)
	obj.SetIsAllGood(true)
	obj.SetSyncInfo(nil)

	obj.SetUpdateLogID(s.LogID)

	return nil
}

func (s *BaseSyncInfo) FromOplog(status types.Status, oplog *BaseOplog, opData OpData) error {
	s.LogID = oplog.ID
	s.UpdateTS = oplog.UpdateTS
	s.UpdaterID = oplog.CreatorID
	s.Status = status

	return nil
}
