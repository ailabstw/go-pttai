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
	"fmt"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMeRevokeOpKey(t *testing.T) {
	NNodes = 1
	isDebug := true

	var bodyString string
	var marshaled []byte
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}

	testCore(t0, bodyString, me0_1, t, isDebug)

	assert.Equal(types.StatusAlive, me0_1.Status)

	// 3. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": []}`

	me0_3 := &me.MyInfo{}

	testCore(t0, bodyString, me0_3, t, isDebug)

	assert.Equal(types.StatusAlive, me0_3.Status)
	assert.Equal(me0_1.ID, me0_3.ID)
	assert.Equal(1, len(me0_3.OwnerIDs))
	assert.Equal(me0_3.ID, me0_3.OwnerIDs[0])
	assert.Equal(true, me0_3.IsOwner(me0_3.ID))

	time.Sleep(5 * time.Second)

	// 4. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_4 string

	testCore(t0, bodyString, &myKey0_4, t, isDebug)
	if isDebug {
		t.Logf("myKey0_4: %v\n", myKey0_4)
	}

	// 8. getOpKeyInfo
	bodyString = `{"id": "testID", "method": "me_getOpKeyInfos", "params": []}`

	dataOpKeyInfos0_8 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfos0_8, t, isDebug)
	assert.Equal(1, len(dataOpKeyInfos0_8.Result))
	opKeyInfo0_8 := dataOpKeyInfos0_8.Result[0]
	assert.Equal(types.StatusAlive, opKeyInfo0_8.Status)

	// 8.1 ptt.GetOps
	bodyString = `{"id": "testID", "method": "ptt_getOps", "params": []}`

	dataOpKeyInfo0_8_1 := &struct {
		Result map[common.Address]*types.PttID `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfo0_8_1, t, isDebug)

	opKeyInfoMap0_8_1 := dataOpKeyInfo0_8_1.Result
	assert.Equal(2, len(opKeyInfoMap0_8_1))
	entityID, ok := opKeyInfoMap0_8_1[*opKeyInfo0_8.Hash]

	assert.Equal(true, ok)
	assert.Equal(me0_3.ID, entityID)

	// 8.2. getOpKeyInfoFromDB
	bodyString = `{"id": "testID", "method": "me_getOpKeyInfosFromDB", "params": []}`

	dataOpKeyInfos0_8_2 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfos0_8_2, t, isDebug)
	assert.Equal(1, len(dataOpKeyInfos0_8_2.Result))
	opKeyInfo0_8_2 := dataOpKeyInfos0_8_2.Result[0]
	assert.Equal(types.StatusAlive, opKeyInfo0_8_2.Status)
	assert.Equal(opKeyInfo0_8.ID, opKeyInfo0_8_2.ID)

	// 9.1. OpKeyOplog
	bodyString = `{"id": "testID", "method": "me_getOpKeyOplogList", "params": ["", 0, 2]}`

	dataOpKeyOplogs0_9_1 := &struct {
		Result []*pkgservice.OpKeyOplog `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyOplogs0_9_1, t, isDebug)
	assert.Equal(1, len(dataOpKeyOplogs0_9_1.Result))
	opKeyOplog0_9_1 := dataOpKeyOplogs0_9_1.Result[0]
	assert.Equal(me0_3.ID, opKeyOplog0_9_1.CreatorID)
	assert.Equal(opKeyInfo0_8.ID, opKeyOplog0_9_1.ObjID)
	assert.Equal(pkgservice.OpKeyOpTypeCreateOpKey, opKeyOplog0_9_1.Op)
	assert.Equal(nilPttID, opKeyOplog0_9_1.PreLogID)
	assert.Equal(types.Bool(true), opKeyOplog0_9_1.IsSync)

	// 10 revoke-key
	marshaled, _ = opKeyInfo0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_revokeOpKey", "params": ["%v", ""]}`, string(marshaled))

	isOk := false
	_, err := testCore(t0, bodyString, &isOk, t, isDebug)
	assert.Equal(false, isOk)
	assert.Equal("invalid me", err.Msg)

	// 10.1 revoke-key
	marshaled, _ = opKeyInfo0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_revokeOpKey", "params": ["%v", "%v"]}`, string(marshaled), string(myKey0_4))

	isOk = false
	_, err = testCore(t0, bodyString, &isOk, t, isDebug)
	assert.Equal(true, isOk)
	assert.Equal(0, err.Code)
	assert.Equal("", err.Msg)

	// 10.2. getOpKeyInfo
	bodyString = `{"id": "testID", "method": "me_getOpKeyInfos", "params": []}`

	dataOpKeyInfos0_10_2 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfos0_10_2, t, isDebug)
	assert.Equal(0, len(dataOpKeyInfos0_10_2.Result))

	// 10.3 ptt.GetOps
	bodyString = `{"id": "testID", "method": "ptt_getOps", "params": []}`

	dataOpKeyInfo0_10_3 := &struct {
		Result map[common.Address]*types.PttID `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfo0_10_3, t, isDebug)

	opKeyInfoMap0_10_3 := dataOpKeyInfo0_10_3.Result

	assert.Equal(1, len(opKeyInfoMap0_10_3))

	// 10.4. getOpKeyInfoFromDB
	bodyString = `{"id": "testID", "method": "me_getOpKeyInfosFromDB", "params": []}`

	dataOpKeyInfos0_10_4 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`

		Error *MyError `json:"error"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfos0_10_4, t, isDebug)
	assert.Equal(1, len(dataOpKeyInfos0_10_4.Result))
	opKeyInfo0_10_4 := dataOpKeyInfos0_10_4.Result[0]
	assert.Equal(types.StatusDeleted, opKeyInfo0_10_4.Status)
	assert.Equal(opKeyInfo0_8.ID, opKeyInfo0_10_4.ID)
}
