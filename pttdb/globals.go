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

package pttdb

import (
	"time"
)

const (
	// Code using batches should try to add this much data to the batch.
	// The value was determined empirically.
	IdealBatchSize = 100 * 1024

	writeDelayNThreshold       = 200
	writeDelayThreshold        = 350 * time.Millisecond
	writeDelayWarningThrottler = 1 * time.Minute

	InfiniteQuota = 0

	SizeDBKeyPrefix          = 5
	OffsetDBKeyPrefixPostfix = 3
)

var (
	dbLastKey = []byte{255}
	ValueTrue = []byte{1}
)

const (
	minCache   = 16
	minHandles = 16

	OpenFileLimit = 64
)
