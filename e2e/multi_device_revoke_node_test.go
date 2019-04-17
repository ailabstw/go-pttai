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
	"reflect"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceRevokeNode(t *testing.T) {
	NNodes = 2
	isDebug := true

	var err error
	var bodyString string
	var marshaled []byte
	var dummyBool bool
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")

	// 0 test-error
	err = testError("http://127.0.0.1:9450")
	assert.Equal(nil, err)

	err = testError("http://127.0.0.1:9451")
	assert.Equal(nil, err)

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
	profileID0_3 := me0_3.ProfileID

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))
	profileID1_3 := me1_3.ProfileID

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

	// 9. getRawMe
	marshaled, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaled))

	me0_9 := &me.MyInfo{}
	testCore(t0, bodyString, me0_9, t, isDebug)
	assert.Equal(types.StatusMigrated, me0_9.Status)
	assert.Equal(2, len(me0_9.OwnerIDs))
	assert.Equal(true, me0_9.IsOwner(me1_3.ID))
	assert.Equal(true, me0_9.IsOwner(me0_3.ID))

	// 9.1. getPeers
	t.Logf("9.1 get Peers")
	bodyString = `{"id": "testID", "method": "me_getPeers", "params": [""]}`

	dataPeers0_9_1 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_9_1, t, isDebug)
	assert.Equal(1, len(dataPeers0_9_1.Result))
	peer0_9_1_0 := dataPeers0_9_1.Result[0]
	assert.Equal(me1_1.NodeID, peer0_9_1_0.NodeID)
	assert.Equal(pkgservice.PeerTypeMe, peer0_9_1_0.PeerType)

	dataPeers1_9_1 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPeers1_9_1, t, isDebug)
	assert.Equal(1, len(dataPeers1_9_1.Result))
	peer1_9_1_0 := dataPeers1_9_1.Result[0]
	assert.Equal(me0_1.NodeID, peer1_9_1_0.NodeID)
	assert.Equal(pkgservice.PeerTypeMe, peer1_9_1_0.PeerType)

	// 9.1 getRawProfile
	t.Logf("9.1 getRawProfile")

	marshaled, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_9_1 := &account.Profile{}
	testCore(t0, bodyString, profile0_9_1, t, isDebug)
	assert.Equal(types.StatusTerminal, profile0_9_1.Status)
	assert.Equal(me0_3.ID, profile0_9_1.MyID)
	assert.Equal(me0_3.ProfileID, profile0_9_1.ID)
	assert.Equal(false, profile0_9_1.IsOwner(me0_1.ID))
	assert.Equal(true, profile0_9_1.IsOwner(me1_1.ID))

	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile1_9_1 := &account.Profile{}
	testCore(t1, bodyString, profile1_9_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile1_9_1.Status)
	assert.Equal(me1_3.ID, profile1_9_1.MyID)
	assert.Equal(me1_3.ProfileID, profile1_9_1.ID)

	// 9.2. getMasterOplogList
	t.Logf("9.2 getMasterOplogList")
	marshaled, _ = profile0_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList0_9_2 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMasterOplogList0_9_2, t, isDebug)
	assert.Equal(2, len(dataGetMasterOplogList0_9_2.Result))
	masterOplog0_9_2 := dataGetMasterOplogList0_9_2.Result[0]
	masterSigns0_9_2 := masterOplog0_9_2.MasterSigns
	assert.Equal(1, len(masterSigns0_9_2))
	assert.Equal(me0_3.ID, masterSigns0_9_2[0].ID)
	assert.Equal(masterOplog0_9_2.ID, masterOplog0_9_2.MasterLogID)

	marshaled, _ = profile1_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList1_9_2 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMasterOplogList1_9_2, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogList1_9_2.Result))
	masterOplog1_9_2 := dataGetMasterOplogList1_9_2.Result[0]
	masterSigns1_9_2 := masterOplog1_9_2.MasterSigns
	assert.Equal(1, len(masterSigns0_9_2))
	assert.Equal(me1_3.ID, masterSigns1_9_2[0].ID)
	assert.Equal(masterOplog1_9_2.ID, masterOplog1_9_2.MasterLogID)
	assert.Equal(types.StatusAlive, masterOplog1_9_2.ToStatus())
	assert.Equal(types.Bool(true), masterOplog1_9_2.IsSync)

	marshaled, _ = profile1_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList0_9_2_1 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMasterOplogList0_9_2_1, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogList0_9_2_1.Result))
	masterOplog0_9_2_1 := dataGetMasterOplogList0_9_2_1.Result[0]
	masterSigns0_9_2_1 := masterOplog0_9_2_1.MasterSigns
	assert.Equal(1, len(masterSigns0_9_2_1))
	assert.Equal(me1_3.ID, masterSigns0_9_2_1[0].ID)
	assert.Equal(masterOplog1_9_2.ID, masterOplog0_9_2_1.MasterLogID)
	assert.Equal(types.StatusAlive, masterOplog0_9_2_1.ToStatus())
	assert.Equal(types.Bool(true), masterOplog0_9_2.IsSync)

	// 9.4. getUserOplogList
	t.Logf("9.4 GetUserOplogList: t0")
	marshaled, _ = profile0_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_9_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_9_4, t, isDebug)
	assert.Equal(6, len(dataGetUserOplogList0_9_4.Result))
	userOplog0_9_4_0 := dataGetUserOplogList0_9_4.Result[0]
	masterSigns0_9_4_0 := userOplog0_9_4_0.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_0))
	assert.Equal(me0_3.ID, masterSigns0_9_4_0[0].ID)
	assert.Equal(masterOplog0_9_2.ID, userOplog0_9_4_0.MasterLogID)
	assert.Equal(account.UserOpTypeCreateProfile, userOplog0_9_4_0.Op)

	userOplog0_9_4_1 := dataGetUserOplogList0_9_4.Result[1]
	masterSigns0_9_4_1 := userOplog0_9_4_1.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_1))
	assert.Equal(me0_3.ID, masterSigns0_9_4_1[0].ID)
	assert.Equal(masterOplog0_9_2.ID, userOplog0_9_4_1.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserName, userOplog0_9_4_1.Op)

	userOplog0_9_4_2 := dataGetUserOplogList0_9_4.Result[2]
	masterSigns0_9_4_2 := userOplog0_9_4_2.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_2))
	assert.Equal(me0_3.ID, masterSigns0_9_4_2[0].ID)
	assert.Equal(masterOplog0_9_2.ID, userOplog0_9_4_2.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserImg, userOplog0_9_4_2.Op)

	userOplog0_9_4_4 := dataGetUserOplogList0_9_4.Result[4]
	masterSigns0_9_4_4 := userOplog0_9_4_4.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_4))
	assert.Equal(me0_3.ID, masterSigns0_9_4_4[0].ID)
	assert.Equal(masterOplog0_9_2.ID, userOplog0_9_4_4.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog0_9_4_4.Op)

	userOplog0_9_4_5 := dataGetUserOplogList0_9_4.Result[5]
	masterSigns0_9_4_5 := userOplog0_9_4_5.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_5))
	assert.Equal(me0_3.ID, masterSigns0_9_4_5[0].ID)
	assert.Equal(masterOplog0_9_2.ID, userOplog0_9_4_5.MasterLogID)
	assert.Equal(account.UserOpTypeDeleteProfile, userOplog0_9_4_5.Op)

	// t1
	t.Logf("9.4 GetUserOplogList: t1")
	marshaled, _ = profile1_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList1_9_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetUserOplogList1_9_4, t, isDebug)
	assert.Equal(6, len(dataGetUserOplogList1_9_4.Result))
	userOplog1_9_4_0 := dataGetUserOplogList1_9_4.Result[0]
	masterSigns1_9_4_0 := userOplog1_9_4_0.MasterSigns
	assert.Equal(1, len(masterSigns1_9_4_0))
	assert.Equal(me1_3.ID, masterSigns1_9_4_0[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog1_9_4_0.MasterLogID)
	assert.Equal(account.UserOpTypeCreateProfile, userOplog1_9_4_0.Op)

	userOplog1_9_4_1 := dataGetUserOplogList1_9_4.Result[1]
	masterSigns1_9_4_1 := userOplog1_9_4_1.MasterSigns
	assert.Equal(1, len(masterSigns1_9_4_1))
	assert.Equal(me1_3.ID, masterSigns1_9_4_1[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog1_9_4_1.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserName, userOplog1_9_4_1.Op)

	userOplog1_9_4_2 := dataGetUserOplogList1_9_4.Result[2]
	masterSigns1_9_4_2 := userOplog1_9_4_2.MasterSigns
	assert.Equal(1, len(masterSigns1_9_4_2))
	assert.Equal(me1_3.ID, masterSigns1_9_4_2[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog1_9_4_2.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserImg, userOplog1_9_4_2.Op)

	userOplog1_9_4_4 := dataGetUserOplogList1_9_4.Result[4]
	masterSigns1_9_4_4 := userOplog1_9_4_4.MasterSigns
	assert.Equal(1, len(masterSigns1_9_4_4))
	assert.Equal(me1_3.ID, masterSigns1_9_4_4[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog1_9_4_4.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog1_9_4_4.Op)

	userOplog1_9_4_5 := dataGetUserOplogList1_9_4.Result[5]
	masterSigns1_9_4_5 := userOplog1_9_4_5.MasterSigns
	assert.Equal(1, len(masterSigns1_9_4_5))
	assert.Equal(me1_3.ID, masterSigns1_9_4_5[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog1_9_4_5.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog1_9_4_5.Op)

	// new t0 (should be the same as t1)
	t.Logf("9.4 GetUserOplogList: new t0")
	marshaled, _ = profile1_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_9_4_1 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_9_4_1, t, isDebug)
	assert.Equal(6, len(dataGetUserOplogList0_9_4_1.Result))
	userOplog0_9_4_1_0 := dataGetUserOplogList0_9_4_1.Result[0]
	mastersigns0_9_4_1_0 := userOplog0_9_4_1_0.MasterSigns
	assert.Equal(1, len(mastersigns0_9_4_1_0))
	assert.Equal(me1_3.ID, mastersigns0_9_4_1_0[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog0_9_4_1_0.MasterLogID)
	assert.Equal(account.UserOpTypeCreateProfile, userOplog0_9_4_1_0.Op)

	userOplog0_9_4_1_1 := dataGetUserOplogList0_9_4_1.Result[1]
	masterSigns0_9_4_1_1 := userOplog0_9_4_1_1.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_1_1))
	assert.Equal(me1_3.ID, masterSigns0_9_4_1_1[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog0_9_4_1_1.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserName, userOplog0_9_4_1_1.Op)

	userOplog0_9_4_1_2 := dataGetUserOplogList0_9_4_1.Result[2]
	masterSigns0_9_4_1_2 := userOplog0_9_4_1_2.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_1_2))
	assert.Equal(me1_3.ID, masterSigns0_9_4_1_2[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog0_9_4_1_2.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserImg, userOplog0_9_4_1_2.Op)

	userOplog0_9_4_1_4 := dataGetUserOplogList0_9_4_1.Result[4]
	masterSigns0_9_4_1_4 := userOplog0_9_4_1_4.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_1_4))
	assert.Equal(me1_3.ID, masterSigns0_9_4_1_4[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog0_9_4_1_4.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog0_9_4_1_4.Op)
	opData0_9_4_1_3 := &account.UserOpAddUserNode{}
	userOplog0_9_4_1_4.GetData(opData0_9_4_1_3)
	assert.Equal(me1_1.NodeID, opData0_9_4_1_3.NodeID)

	userOplog0_9_4_1_5 := dataGetUserOplogList0_9_4_1.Result[5]
	masterSigns0_9_4_1_5 := userOplog0_9_4_1_5.MasterSigns
	assert.Equal(1, len(masterSigns0_9_4_1_5))
	assert.Equal(me1_3.ID, masterSigns0_9_4_1_5[0].ID)
	assert.Equal(masterOplog1_9_2.ID, userOplog0_9_4_1_5.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog0_9_4_1_5.Op)
	opData0_9_4_1_4 := &account.UserOpAddUserNode{}
	userOplog0_9_4_1_5.GetData(opData0_9_4_1_4)
	assert.Equal(me0_1.NodeID, opData0_9_4_1_4.NodeID)

	assert.Equal(dataGetUserOplogList1_9_4, dataGetUserOplogList0_9_4_1)

	// 10. revoke-node
	marshaled, _ = me1_1.NodeID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_removeNode", "params": ["%v"]}`, string(marshaled))

	testCore(t0, bodyString, &dummyBool, t, isDebug)
	assert.Equal(true, dummyBool)

	// wait 10 seconds

	time.Sleep(10 * time.Second)

	// 11.0 test-error
	err = testError("http://127.0.0.1:9450")
	assert.Equal(nil, err)

	err = testError("http://127.0.0.1:9451")
	assert.NotEqual(nil, err)

	// 11. get my nodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_11 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_11, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes0_11.Result))
	myNode0_11_0 := dataGetMyNodes0_11.Result[0]

	assert.Equal(types.StatusAlive, myNode0_11_0.Status)
	assert.Equal(me0_1.NodeID, myNode0_11_0.NodeID)

	// 12. getPeers
	bodyString = `{"id": "testID", "method": "me_getPeers", "params": [""]}`

	dataPeers0_12 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_12, t, isDebug)
	assert.Equal(0, len(dataPeers0_12.Result))

	// 13. getPeers
	bodyString = `{"id": "testID", "method": "ptt_getPeers", "params": []}`

	dataPeers0_13 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_13, t, isDebug)
	assert.Equal(0, len(dataPeers0_13.Result))

	// 14. getUserNode
	t.Logf("10.6 GetUserNodeList")
	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserNodeList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserNodeList0_14 := &struct {
		Result []*account.UserNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserNodeList0_14, t, isDebug)
	assert.Equal(2, len(dataGetUserNodeList0_14.Result))
	userNode0_14_0 := dataGetUserNodeList0_14.Result[0]
	userNode0_14_1 := dataGetUserNodeList0_14.Result[1]

	var myUserNode0_14 *account.UserNode
	var theirUserNode0_14 *account.UserNode
	if reflect.DeepEqual(userNode0_14_0.NodeID, me0_1.NodeID) {
		myUserNode0_14 = userNode0_14_0
		theirUserNode0_14 = userNode0_14_1
	} else {
		myUserNode0_14 = userNode0_14_1
		theirUserNode0_14 = userNode0_14_0
	}
	assert.Equal(types.StatusAlive, myUserNode0_14.Status)
	assert.Equal(types.StatusDeleted, theirUserNode0_14.Status)

	// 15. new t0 (should be the same as t1)
	t.Logf("15 GetUserOplogList: new t0")
	marshaled, _ = profile1_9_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_15 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_15, t, isDebug)
	assert.Equal(7, len(dataGetUserOplogList0_15.Result))
	assert.Equal(dataGetUserOplogList0_9_4_1.Result, dataGetUserOplogList0_15.Result[:6])
	userOplog0_15 := dataGetUserOplogList0_15.Result[6]
	assert.Equal(types.StatusAlive, userOplog0_15.ToStatus())
	assert.Equal(account.UserOpTypeRemoveUserNode, userOplog0_15.Op)

}
