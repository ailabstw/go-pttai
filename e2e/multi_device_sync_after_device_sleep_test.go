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
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceSyncAfterDeviceSleep(t *testing.T) {
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
	nodeID1_1 := me1_1.NodeID
	pubKey1_1, _ := nodeID1_1.Pubkey()
	nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

	// 2. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_2 := &me.MyInfo{}
	testCore(t0, bodyString, me0_2, t, isDebug)
	assert.Equal(types.StatusAlive, me0_2.Status)
	assert.Equal(me0_1.ID, me0_2.ID)
	assert.Equal(1, len(me0_2.OwnerIDs))
	assert.Equal(me0_2.ID, me0_2.OwnerIDs[0])
	assert.Equal(true, me0_2.IsOwner(me0_2.ID))

	me1_2 := &me.MyInfo{}
	testCore(t1, bodyString, me1_2, t, isDebug)
	assert.Equal(types.StatusAlive, me1_2.Status)
	assert.Equal(me1_1.ID, me1_2.ID)
	assert.Equal(1, len(me1_2.OwnerIDs))
	assert.Equal(me1_2.ID, me1_2.OwnerIDs[0])
	assert.Equal(true, me1_2.IsOwner(me1_2.ID))

	// 3. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_3 string

	testCore(t0, bodyString, &myKey0_3, t, isDebug)
	if isDebug {
		t.Logf("myKey0_3: %v\n", myKey0_3)
	}

	// 4. show-me-url
	bodyString = `{"id": "testID", "method": "me_showMeURL", "params": []}`

	dataShowMeURL1_4 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowMeURL1_4, t, isDebug)
	meURL1_4 := dataShowMeURL1_4.URL

	// 5. me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_5 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_5, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes0_5.Result))

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes1_5 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_5, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes1_5.Result))

	// 5.1 ptt-shutdown
	bodyString = `{"id": "testID", "method": "ptt_shutdown", "params": []}`

	resultString := `{"jsonrpc":"2.0","id":"testID","result":true}`
	testBodyEqualCore(t0, bodyString, resultString, t)

	time.Sleep(5 * time.Second)

	// 5.2 test-error
	err0_5_2 := testError("http://127.0.0.1:9450")
	assert.NotEqual(nil, err0_5_2)

	// 8.4 start-node
	startNode(t, 0)

	// wait 15 seconds
	time.Sleep(15 * time.Second)

	// 8.5 test-error
	err0_8_7 := testError("http://127.0.0.1:9450")
	assert.Equal(nil, err0_8_7)

	// 8.5.1 get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_8_5_1 := &me.BackendMyInfo{}
	testCore(t0, bodyString, me0_8_5_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_8_5_1.Status)
	assert.Equal(me0_1.ID, me0_8_5_1.ID)
	assert.Equal(me0_1.NodeID, me0_8_5_1.NodeID)

	// 8.5.2 getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_8_5_2 := &me.MyInfo{}
	testCore(t0, bodyString, me0_8_5_2, t, isDebug)
	assert.Equal(types.StatusAlive, me0_8_5_2.Status)
	assert.Equal(me0_2.ID, me0_8_5_2.ID)
	assert.Equal(1, len(me0_8_5_2.OwnerIDs))
	assert.Equal(me0_8_5_2.ID, me0_8_5_2.OwnerIDs[0])
	assert.Equal(true, me0_8_5_2.IsOwner(me0_8_5_2.ID))

	me1_8_5_2 := &me.MyInfo{}
	testCore(t1, bodyString, me1_8_5_2, t, isDebug)
	assert.Equal(types.StatusAlive, me1_8_5_2.Status)
	assert.Equal(me1_2.ID, me1_8_5_2.ID)
	assert.Equal(1, len(me1_8_5_2.OwnerIDs))
	assert.Equal(me1_8_5_2.ID, me1_8_5_2.OwnerIDs[0])
	assert.Equal(true, me1_8_5_2.IsOwner(me1_8_5_2.ID))

	// 8.5.3 show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_8_5_3 string

	testCore(t0, bodyString, &myKey0_8_5_3, t, isDebug)
	if isDebug {
		t.Logf("myKey0_3: %v\n", myKey0_8_5_3)
	}

	// 8.6 join-me
	log.Debug("7.5 join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_4, myKey0_8_5_3)

	dataJoinMe0_8_6 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_8_6, t, true)

	assert.Equal(me1_2.ID, dataJoinMe0_8_6.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_8_6.NodeID)

	// wait 10
	t.Logf("wait 15 seconds for hand-shaking")
	time.Sleep(TimeSleepRestart)

	// 8.7 me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_8_7 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_8_7, t, isDebug)
	assert.Equal(2, len(dataGetMyNodes0_8_7.Result))
	myNode0_8_7_0 := dataGetMyNodes0_8_7.Result[0]
	myNode0_8_7_1 := dataGetMyNodes0_8_7.Result[1]

	assert.Equal(types.StatusAlive, myNode0_8_7_0.Status)
	assert.Equal(types.StatusAlive, myNode0_8_7_1.Status)

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes1_8_7 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_8_7, t, isDebug)
	assert.Equal(2, len(dataGetMyNodes1_8_7.Result))
	myNode1_8_0 := dataGetMyNodes1_8_7.Result[0]
	myNode1_8_1 := dataGetMyNodes1_8_7.Result[1]

	assert.Equal(types.StatusAlive, myNode1_8_0.Status)
	assert.Equal(types.StatusAlive, myNode1_8_1.Status)

	// 8.7 getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_8_7 := &me.MyInfo{}
	testCore(t0, bodyString, me0_8_7, t, isDebug)
	assert.Equal(types.StatusAlive, me0_8_7.Status)
	assert.Equal(1, len(me0_8_7.OwnerIDs))
	assert.Equal(me1_2.ID, me0_8_7.OwnerIDs[0])
	assert.Equal(true, me0_8_7.IsOwner(me1_2.ID))

	me1_8_7 := &me.MyInfo{}
	testCore(t1, bodyString, me1_8_7, t, isDebug)
	assert.Equal(types.StatusAlive, me1_8_7.Status)
	assert.Equal(me1_2.ID, me1_8_7.ID)
	assert.Equal(1, len(me1_8_7.OwnerIDs))
	assert.Equal(me1_2.ID, me1_8_7.OwnerIDs[0])
	assert.Equal(true, me1_8_7.IsOwner(me1_2.ID))

	// 9. MasterOplog
	bodyString = `{"id": "testID", "method": "me_getMyMasterOplogList", "params": ["", "", 0, 2]}`

	dataMasterOplogs0_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_9, t, isDebug)
	assert.Equal(3, len(dataMasterOplogs0_9.Result))
	masterOplog0_9 := dataMasterOplogs0_9.Result[0]
	assert.Equal(me1_2.ID[:common.AddressLength], masterOplog0_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_2.ID, masterOplog0_9.ObjID)
	assert.Equal(me.MasterOpTypeAddMaster, masterOplog0_9.Op)
	assert.Equal(nilPttID, masterOplog0_9.PreLogID)
	assert.Equal(types.Bool(true), masterOplog0_9.IsSync)
	assert.Equal(masterOplog0_9.ID, masterOplog0_9.MasterLogID)

	dataMasterOplogs1_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogs1_9, t, isDebug)
	assert.Equal(3, len(dataMasterOplogs1_9.Result))
	masterOplog1_9 := dataMasterOplogs1_9.Result[0]
	assert.Equal(me1_2.ID[:common.AddressLength], masterOplog1_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_2.ID, masterOplog1_9.ObjID)
	assert.Equal(me.MasterOpTypeAddMaster, masterOplog1_9.Op)
	assert.Equal(nilPttID, masterOplog1_9.PreLogID)
	assert.Equal(types.Bool(true), masterOplog1_9.IsSync)
	assert.Equal(masterOplog1_9.ID, masterOplog1_9.MasterLogID)

	for i, oplog := range dataMasterOplogs0_9.Result {
		oplog1 := dataMasterOplogs1_9.Result[i]
		oplog.CreateTS = oplog1.CreateTS
		oplog.CreatorID = oplog1.CreatorID
		oplog.CreatorHash = oplog1.CreatorHash
		oplog.Salt = oplog1.Salt
		oplog.Sig = oplog1.Sig
		oplog.Pubkey = oplog1.Pubkey
		oplog.KeyExtra = oplog1.KeyExtra
		oplog.UpdateTS = oplog1.UpdateTS
		oplog.Hash = oplog1.Hash
		oplog.IsNewer = oplog1.IsNewer
		oplog.Extra = oplog1.Extra
	}
	assert.Equal(dataMasterOplogs0_9, dataMasterOplogs1_9)

	// 9.1. getRawMe
	marshaled, _ = me0_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaled))

	me0_9_1 := &me.MyInfo{}
	testCore(t0, bodyString, me0_9_1, t, isDebug)
	assert.Equal(types.StatusMigrated, me0_9_1.Status)
	assert.Equal(2, len(me0_9_1.OwnerIDs))
	assert.Equal(true, me0_9_1.IsOwner(me1_2.ID))
	assert.Equal(true, me0_9_1.IsOwner(me0_2.ID))

	// 9.2. MeOplog
	bodyString = `{"id": "testID", "method": "me_getMeOplogList", "params": ["", 0, 2]}`

	dataMeOplogs0_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMeOplogs0_9_2, t, isDebug)
	assert.Equal(1, len(dataMeOplogs0_9_2.Result))
	meOplog0_9_2 := dataMeOplogs0_9_2.Result[0]
	assert.Equal(me1_2.ID, meOplog0_9_2.CreatorID)
	assert.Equal(me1_2.ID, meOplog0_9_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog0_9_2.Op)
	assert.Equal(nilPttID, meOplog0_9_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog0_9_2.IsSync)
	assert.Equal(masterOplog1_9.ID, meOplog0_9_2.MasterLogID)
	assert.Equal(me1_2.LogID, meOplog0_9_2.ID)
	masterSign0_9_2 := meOplog0_9_2.MasterSigns[0]
	assert.Equal(nodeAddr1_1[:], masterSign0_9_2.ID[:common.AddressLength])
	assert.Equal(me1_2.ID[:common.AddressLength], masterSign0_9_2.ID[common.AddressLength:])
	assert.Equal(me0_8_7.LogID, meOplog0_9_2.ID)

	dataMeOplogs1_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMeOplogs1_9_2, t, isDebug)
	assert.Equal(1, len(dataMeOplogs1_9_2.Result))
	meOplog1_9_2 := dataMeOplogs1_9_2.Result[0]
	assert.Equal(me1_2.ID, meOplog1_9_2.CreatorID)
	assert.Equal(me1_2.ID, meOplog1_9_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog1_9_2.Op)
	assert.Equal(nilPttID, meOplog1_9_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog1_9_2.IsSync)
	assert.Equal(masterOplog1_9.ID, meOplog1_9_2.MasterLogID)
	assert.Equal(me1_2.LogID, meOplog1_9_2.ID)
	masterSign1_9_2 := meOplog1_9_2.MasterSigns[0]
	assert.Equal(nodeAddr1_1[:], masterSign1_9_2.ID[:common.AddressLength])
	assert.Equal(me1_2.ID[:common.AddressLength], masterSign1_9_2.ID[common.AddressLength:])
	assert.Equal(meOplog0_9_2, meOplog1_9_2)
	assert.Equal(me1_8_7.LogID, meOplog1_9_2.ID)
}
