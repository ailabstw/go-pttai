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
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceCancelRequest(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	//var marshaled []byte
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}
	testCore(t0, bodyString, me0_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_1.Status)

	//nodeID0_1 := me0_1.NodeID
	//pubKey0_1, _ := nodeID0_1.Pubkey()
	//nodeAddr0_1 := crypto.PubkeyToAddress(*pubKey0_1)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)
	//nodeID1_1 := me1_1.NodeID
	//pubKey1_1, _ := nodeID1_1.Pubkey()
	//nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

	// 3. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_3 := &me.MyInfo{}
	testCore(t0, bodyString, me0_3, t, isDebug)
	assert.Equal(types.StatusAlive, me0_3.Status)
	assert.Equal(me0_1.ID, me0_3.ID)
	assert.Equal(1, len(me0_3.OwnerIDs))
	assert.Equal(me0_3.ID, me0_3.OwnerIDs[0])
	assert.Equal(true, me0_3.IsOwner(me0_3.ID))

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))

	// 4. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_4 string

	testCore(t0, bodyString, &myKey0_4, t, isDebug)
	if isDebug {
		t.Logf("myKey0_4: %v\n", myKey0_4)
	}

	// 5. show-me-url
	bodyString = `{"id": "testID", "method": "me_showMeURL", "params": []}`

	dataShowMeURL1_5 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowMeURL1_5, t, isDebug)
	meURL1_5 := dataShowMeURL1_5.URL

	// 6. me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_6 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_6, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes0_6.Result))

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes1_6 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_6, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes1_6.Result))

	// 6.1 getJoinKeys
	bodyString = `{"id": "testID", "method": "me_getJoinKeyInfos", "params": [""]}`
	dataGetJoinKeys0_6_1 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetJoinKeys0_6_1, t, isDebug)
	assert.Equal(1, len(dataGetJoinKeys0_6_1.Result))

	// 6. shutdown
	bodyString = `{"id": "testID", "method": "ptt_shutdown", "params": []}`

	resultString := `{"jsonrpc":"2.0","id":"testID","result":true}`
	testBodyEqualCore(t1, bodyString, resultString, t)

	time.Sleep(5 * time.Second)

	// 7. join-me
	log.Debug("7. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey0_4)

	dataJoinMe0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_7.NodeID)

	// 8. get-me-requests
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getMeRequests", "params": [""]}`)

	dataGetMeRequests0_8 := &struct {
		Result []*pkgservice.BackendJoinRequest `json:"result"`
	}{}

	testListCore(t0, bodyString, dataGetMeRequests0_8, t, isDebug)

	assert.Equal(1, len(dataGetMeRequests0_8.Result))
	meRequest0_8_0 := dataGetMeRequests0_8.Result[0]

	// 9. remove me-requests
	hashStr := base64.StdEncoding.EncodeToString(meRequest0_8_0.Hash)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_removeMeRequests", "params": ["", "%v"]}`, hashStr)

	t.Logf("9. remove me-requests: bodyString: %v", bodyString)

	dataRemoveMeRequests0_9 := false
	testCore(t0, bodyString, &dataRemoveMeRequests0_9, t, isDebug)

	assert.Equal(true, dataRemoveMeRequests0_9)

	// 10. get-me-requests
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getMeRequests", "params": [""]}`)

	dataGetMeRequests0_10 := &struct {
		Result []*pkgservice.BackendJoinRequest `json:"result"`
	}{}

	testListCore(t0, bodyString, dataGetMeRequests0_10, t, isDebug)

	assert.Equal(0, len(dataGetMeRequests0_10.Result))
}
