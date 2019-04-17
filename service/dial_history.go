// Copyright 2019 The go-pttai Authors
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
	"container/heap"
	"sync"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ethereum/go-ethereum/common"
)

type DialHistory struct {
	hist   *DialHistoryHeap
	theMap map[discover.NodeID]*DialInfo
	lock   sync.RWMutex
}

type DialHistoryHeap []*DialInfo

type DialInfo struct {
	NodeID   *discover.NodeID
	UpdateTS types.Timestamp
	OpKey    *common.Address
}

func NewDialHistory() *DialHistory {
	return &DialHistory{
		hist:   &DialHistoryHeap{},
		theMap: make(map[discover.NodeID]*DialInfo),
	}
}

func (h *DialHistory) Add(id *discover.NodeID, opKey *common.Address) error {
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	dialInfo := &DialInfo{NodeID: id, UpdateTS: ts, OpKey: opKey}
	heap.Push(h.hist, dialInfo)
	h.theMap[*id] = dialInfo

	return nil
}

func (h *DialHistory) Get(id *discover.NodeID) *DialInfo {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.theMap[*id]
}

func (h *DialHistory) Expire() {
	expireTS, err := types.GetTimestamp()
	if err != nil {
		return
	}

	expireTS.Ts -= ExpireDialHistorySeconds
	for h.hist.Len() > 0 && h.hist.min().UpdateTS.IsLess(expireTS) {
		h.Pop()
	}
}

func (h *DialHistory) Pop() {
	h.lock.Lock()
	defer h.lock.Unlock()

	dialInfo := heap.Pop(h.hist).(*DialInfo)

	delete(h.theMap, *dialInfo.NodeID)
}

// heap.Interface boilerplate
// Use only these methods to access or modify DialHistory.
func (h DialHistoryHeap) min() *DialInfo {
	return h[0]
}

func (h DialHistoryHeap) Len() int           { return len(h) }
func (h DialHistoryHeap) Less(i, j int) bool { return h[i].UpdateTS.IsLess(h[j].UpdateTS) }
func (h DialHistoryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *DialHistoryHeap) Push(x interface{}) {
	*h = append(*h, x.(*DialInfo))
}
func (h *DialHistoryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
