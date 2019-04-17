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

package e2e

import (
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

func TestMultiDeviceRevokeMe(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
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

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)

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

	// 7. join-me
	log.Debug("7. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey0_4)

	dataJoinMe0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_7.NodeID)

	// wait 15
	t.Logf("wait 15 seconds for hand-shaking")
	time.Sleep(15 * time.Second)

	// 8. me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_8, t, isDebug)
	assert.Equal(2, len(dataGetMyNodes0_8.Result))
	myNode0_8_0 := dataGetMyNodes0_8.Result[0]
	myNode0_8_1 := dataGetMyNodes0_8.Result[1]

	assert.Equal(types.StatusAlive, myNode0_8_0.Status)
	assert.Equal(types.StatusAlive, myNode0_8_1.Status)

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes1_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_8, t, isDebug)
	assert.Equal(2, len(dataGetMyNodes1_8.Result))
	myNode1_8_0 := dataGetMyNodes1_8.Result[0]
	myNode1_8_1 := dataGetMyNodes1_8.Result[1]

	assert.Equal(types.StatusAlive, myNode1_8_0.Status)
	assert.Equal(types.StatusAlive, myNode1_8_1.Status)

	// 8.1. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_8_1 := &me.MyInfo{}
	testCore(t0, bodyString, me0_8_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_8_1.Status)
	assert.Equal(1, len(me0_8_1.OwnerIDs))
	assert.Equal(me1_3.ID, me0_8_1.OwnerIDs[0])
	assert.Equal(true, me0_8_1.IsOwner(me1_3.ID))

	me1_8_1 := &me.MyInfo{}
	testCore(t1, bodyString, me1_8_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_8_1.Status)
	assert.Equal(me1_3.ID, me1_8_1.ID)
	assert.Equal(1, len(me1_8_1.OwnerIDs))
	assert.Equal(me1_3.ID, me1_8_1.OwnerIDs[0])
	assert.Equal(true, me1_8_1.IsOwner(me1_3.ID))

	// 8.2. getPeers
	bodyString = `{"id": "testID", "method": "me_getPeers", "params": [""]}`

	dataPeers0_8_2 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_8_2, t, isDebug)
	assert.Equal(1, len(dataPeers0_8_2.Result))

	dataPeers1_8_2 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPeers1_8_2, t, isDebug)
	assert.Equal(1, len(dataPeers1_8_2.Result))

	// 9. getRawMe
	marshaled, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaled))

	me0_9 := &me.MyInfo{}
	testCore(t0, bodyString, me0_9, t, isDebug)
	assert.Equal(types.StatusMigrated, me0_9.Status)
	assert.Equal(2, len(me0_9.OwnerIDs))
	assert.Equal(true, me0_9.IsOwner(me1_3.ID))
	assert.Equal(true, me0_9.IsOwner(me0_3.ID))

	// 10 revoke-me
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_revoke", "params": [""]}`)

	isOk := false
	_, err := testCore(t0, bodyString, &isOk, t, isDebug)
	assert.Equal(false, isOk)
	assert.Equal("invalid me", err.Msg)

	// 10.1 revoke
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_revoke", "params": ["%v"]}`, string(myKey0_4))

	isOk = false
	_, err = testCore(t0, bodyString, &isOk, t, isDebug)
	assert.Equal(false, isOk)
	assert.Equal("invalid me", err.Msg)

	// 10.2. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_10_2 string

	testCore(t0, bodyString, &myKey0_10_2, t, isDebug)
	if isDebug {
		t.Logf("myKey0_10_2: %v\n", myKey0_10_2)
	}

	// 10.3 revoke
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_revoke", "params": ["%v"]}`, string(myKey0_10_2))

	isOk = false
	_, err = testCore(t0, bodyString, &isOk, t, isDebug)
	assert.Equal(true, isOk)
	assert.Equal(0, err.Code)
	assert.Equal("", err.Msg)

	// wait 5 seconds
	time.Sleep(10 * time.Second)

	// 10.4. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_10_4 := &me.MyInfo{}
	testCore(t0, bodyString, me0_10_4, t, isDebug)
	assert.Equal(types.StatusDeleted, me0_10_4.Status)
	assert.Equal(1, len(me0_10_4.OwnerIDs))
	assert.Equal(me1_3.ID, me0_10_4.OwnerIDs[0])
	assert.Equal(true, me0_10_4.IsOwner(me1_3.ID))

	me1_10_4 := &me.MyInfo{}
	testCore(t1, bodyString, me1_10_4, t, isDebug)
	assert.Equal(types.StatusDeleted, me1_10_4.Status)
	assert.Equal(me1_3.ID, me1_10_4.ID)
	assert.Equal(1, len(me1_10_4.OwnerIDs))
	assert.Equal(me1_3.ID, me1_10_4.OwnerIDs[0])
	assert.Equal(true, me1_10_4.IsOwner(me1_3.ID))
}
