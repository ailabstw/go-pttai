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

	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestPttPeers(t *testing.T) {
	var bodyString string
	NNodes = 3

	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")
	t2 := baloo.New("http://127.0.0.1:9452")

	// 1. ptt_countPeers. ensure connecting to each other.
	bodyString = `{"id": "testID", "method": "ptt_countPeers", "params": []}`
	resultString := `{"jsonrpc":"2.0","id":"testID","result":{"M":0,"I":0,"E":0,"R":2}}`
	testBodyEqualCore(t0, bodyString, resultString, t)
	testBodyEqualCore(t1, bodyString, resultString, t)
	testBodyEqualCore(t2, bodyString, resultString, t)

	// 2. ptt_getPeers.
	bodyString = `{"id": "testID", "method": "ptt_getPeers", "params": []}`

	dataGetPeers0_2 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}

	testListCore(t0, bodyString, dataGetPeers0_2, t, true)
	assert.Equal(2, len(dataGetPeers0_2.Result))
	peer0 := dataGetPeers0_2.Result[0]
	assert.Equal(pkgservice.PeerTypeRandom, peer0.PeerType)
}
