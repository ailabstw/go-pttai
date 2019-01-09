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

package e2e

import (
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestOffsetSecondBasic(t *testing.T) {
	NNodes = 1
	isDebug := true

	assert := assert.New(t)

	// setup test
	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")

	// 1. get offset second

	bodyString := `{"id": "testID", "method": "ptt_getOffsetSecond", "params": []}`
	resultString := `{"jsonrpc":"2.0","id":"testID","result":0}`

	testBodyEqualCore(t0, bodyString, resultString, t)

	// 2. get timestamp

	bodyString = `{"id": "testID", "method": "ptt_getTimestamp", "params": []}`

	ts0_2 := types.Timestamp{}

	testCore(t0, bodyString, &ts0_2, t, isDebug)

	// 3. set offset second

	bodyString = `{"id": "testID", "method": "ptt_setOffsetSecond", "params": [10000]}`
	resultString = `{"jsonrpc":"2.0","id":"testID","result":true}`

	testBodyEqualCore(t0, bodyString, resultString, t)

	// 4. get timestamp

	bodyString = `{"id": "testID", "method": "ptt_getTimestamp", "params": []}`

	ts0_4 := types.Timestamp{}

	testCore(t0, bodyString, &ts0_4, t, isDebug)

	assert.Equal(true, ts0_4.Ts-ts0_2.Ts >= 10000)

	// 4.1. get offset second

	bodyString = `{"id": "testID", "method": "ptt_getOffsetSecond", "params": []}`
	resultString = `{"jsonrpc":"2.0","id":"testID","result":10000}`

	testBodyEqualCore(t0, bodyString, resultString, t)

}
