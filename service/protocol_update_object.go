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
	"bytes"

	"github.com/ailabstw/go-pttai/common/types"
)

func ProtocolUpdateObject() error {
	return nil
}

func isReplaceOrigSyncInfo(syncInfo SyncInfo, status types.Status, ts types.Timestamp, newLogID *types.PttID) bool {

	if syncInfo == nil {
		return true
	}

	statusClass := types.StatusToStatusClass(status)
	syncStatusClass := types.StatusToStatusClass(syncInfo.GetStatus())

	switch syncStatusClass {
	case types.StatusClassInternalMigrate:
		syncStatusClass = types.StatusClassInternalDelete
	case types.StatusClassPendingMigrate:
		syncStatusClass = types.StatusClassPendingDelete
	case types.StatusClassMigrated:
		syncStatusClass = types.StatusClassAlive
	}

	if statusClass < syncStatusClass {
		return false
	}
	if statusClass > syncStatusClass {
		return true
	}

	syncTS := syncInfo.GetUpdateTS()
	if syncTS.IsLess(ts) {
		return false
	}
	if ts.IsLess(syncTS) {
		return true
	}

	origLogID := syncInfo.GetLogID()
	return bytes.Compare(origLogID[:], newLogID[:]) > 0
}
