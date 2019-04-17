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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	startNode(t, 0, 0, false)

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
	assert.Equal(me0_8_7.ID, me1_8_7.ID)
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

	// 9.3 get user name
	marshaled, _ = me1_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawUserName", "params": ["%v"]}`, string(marshaled))

	dataGetUserName0_9_3 := &account.UserName{}
	testCore(t0, bodyString, dataGetUserName0_9_3, t, isDebug)
	assert.Equal(me1_2.ID, dataGetUserName0_9_3.ID)

	dataGetUserName1_9_3 := &account.UserName{}
	testCore(t1, bodyString, dataGetUserName1_9_3, t, isDebug)
	assert.Equal(me1_2.ID, dataGetUserName1_9_3.ID)
	assert.Equal(dataGetUserName0_9_3.Name, dataGetUserName1_9_3.Name)

	// 9.4 t0 my_setMyImg
	normalizedImgStr := "data:image/png;base64,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"

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_setMyImage", "params": ["%v"]}`, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==")
	t.Logf("setMyImage: bodyString: %v", bodyString)
	dataSetMyImg0_9_4 := &account.UserImg{}
	testCore(t0, bodyString, dataSetMyImg0_9_4, t, false)
	assert.Equal(me1_2.ID, dataSetMyImg0_9_4.ID)

	time.Sleep(20 * time.Second)

	// 9.5 get user img
	marshaled, _ = me1_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawUserImg", "params": ["%v"]}`, string(marshaled))

	dataGetUserImg0_9_5 := &account.UserImg{}
	testCore(t0, bodyString, dataGetUserImg0_9_5, t, isDebug)
	assert.Equal(me1_2.ID, dataGetUserImg0_9_5.ID)
	assert.Equal(normalizedImgStr, dataGetUserImg0_9_5.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg0_9_5.Status)

	dataGetUserImg1_9_5 := &account.UserImg{}
	testCore(t1, bodyString, dataGetUserImg1_9_5, t, isDebug)
	assert.Equal(me1_2.ID, dataGetUserImg1_9_5.ID)
	assert.Equal(normalizedImgStr, dataGetUserImg1_9_5.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg1_9_5.Status)
	assert.Equal(dataGetUserImg0_9_5.Str, dataGetUserImg1_9_5.Str)

	// 9.6 t1 my_setMyImg
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_setMyImage", "params": ["%v"]}`, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==")
	t.Logf("setMyImage: bodyString: %v", bodyString)
	dataSetMyImg1_9_4 := &account.UserImg{}
	testCore(t1, bodyString, dataSetMyImg1_9_4, t, false)
	assert.Equal(me1_2.ID, dataSetMyImg1_9_4.ID)

	time.Sleep(20 * time.Second)

	// 9.7 get user img
	marshaled, _ = me1_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawUserImg", "params": ["%v"]}`, string(marshaled))

	dataGetUserImg0_9_7 := &account.UserImg{}
	testCore(t0, bodyString, dataGetUserImg0_9_7, t, isDebug)
	assert.Equal(me1_2.ID, dataGetUserImg0_9_7.ID)
	assert.Equal(normalizedImgStr, dataGetUserImg0_9_7.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg0_9_7.Status)

	dataGetUserImg1_9_7 := &account.UserImg{}
	testCore(t1, bodyString, dataGetUserImg1_9_7, t, isDebug)
	assert.Equal(me1_2.ID, dataGetUserImg1_9_7.ID)
	assert.Equal(normalizedImgStr, dataGetUserImg1_9_7.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg1_9_7.Status)
	assert.Equal(dataGetUserImg0_9_7.Str, dataGetUserImg1_9_7.Str)
}
