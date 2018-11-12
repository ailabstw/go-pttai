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

type ProcessInfo interface{}

func ProcessInfoToSyncIDList(info map[types.PttID]*BaseOplog, op OpType) []*SyncID {
	theList := make([]*SyncID, 0, len(info))
	for _, eachLog := range info {
		if eachLog.Op == op {
			theList = append(theList, &SyncID{ID: eachLog.ObjID, LogID: eachLog.ID})
		}
	}
	return theList
}

func ProcessInfoToLogs(info map[types.PttID]*BaseOplog, op OpType) []*BaseOplog {
	theList := make([]*BaseOplog, 0, len(info))
	for _, eachLog := range info {
		if eachLog.Op == op {
			theList = append(theList, eachLog)
		}
	}
	return theList
}

func ProcessInfoToBroadcastLogs(info map[types.PttID]*BaseOplog, broadcastLogs []*BaseOplog) []*BaseOplog {
	for _, eachLog := range info {
		broadcastLogs = append(broadcastLogs, eachLog)
	}
	return broadcastLogs
}
