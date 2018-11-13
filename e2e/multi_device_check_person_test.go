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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceCheckPerson(t *testing.T) {
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

	//nodeID0_1 := me0_1.NodeID
	//pubKey0_1, _ := nodeID0_1.Pubkey()
	//nodeAddr0_1 := crypto.PubkeyToAddress(*pubKey0_1)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)
	nodeID1_1 := me1_1.NodeID
	pubKey1_1, _ := nodeID1_1.Pubkey()
	nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

	// 3. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": []}`

	me0_3 := &me.MyInfo{}
	testCore(t0, bodyString, me0_3, t, isDebug)
	assert.Equal(types.StatusAlive, me0_3.Status)
	assert.Equal(me0_1.ID, me0_3.ID)
	assert.Equal(1, len(me0_3.OwnerIDs))
	assert.Equal(me0_3.ID, me0_3.OwnerIDs[0])
	assert.Equal(true, me0_3.IsOwner(me0_3.ID))
	profileID0_3 := me0_3.MyProfileID

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))
	profileID1_3 := me1_3.MyProfileID

	// 3.1 getRawProfile
	marshaled, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_3_1 := &account.Profile{}
	testCore(t0, bodyString, profile0_3_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile0_3_1.Status)
	assert.Equal(me0_3.ID, profile0_3_1.MyID)
	assert.Equal(me0_3.MyProfileID, profile0_3_1.ID)

	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile1_3_1 := &account.Profile{}
	testCore(t1, bodyString, profile1_3_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile1_3_1.Status)
	assert.Equal(me1_3.ID, profile1_3_1.MyID)
	assert.Equal(me1_3.MyProfileID, profile1_3_1.ID)

	// 3.2. getMasterOplogList
	marshaled, _ = profile0_3_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList0_3_2 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMasterOplogList0_3_2, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogList0_3_2.Result))
	masterOplog0_3_2 := dataGetMasterOplogList0_3_2.Result[0]
	masterSigns0_3_2 := masterOplog0_3_2.MasterSigns
	assert.Equal(1, len(masterSigns0_3_2))
	assert.Equal(me0_3.ID, masterSigns0_3_2[0].ID)
	assert.Equal(masterOplog0_3_2.ID, masterOplog0_3_2.MasterLogID)

	marshaled, _ = profile1_3_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList1_3_2 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMasterOplogList1_3_2, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogList1_3_2.Result))
	masterOplog1_3_2 := dataGetMasterOplogList1_3_2.Result[0]
	masterSigns1_3_2 := masterOplog1_3_2.MasterSigns
	assert.Equal(1, len(masterSigns0_3_2))
	assert.Equal(me1_3.ID, masterSigns1_3_2[0].ID)
	assert.Equal(masterOplog1_3_2.ID, masterOplog1_3_2.MasterLogID)

	// 3.3. getMemberOplogList
	marshaled, _ = profile0_3_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMembergOplogList0_3_3 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMembergOplogList0_3_3, t, isDebug)
	assert.Equal(1, len(dataGetMembergOplogList0_3_3.Result))
	memberOplog0_3_3 := dataGetMembergOplogList0_3_3.Result[0]
	masterSigns0_3_3 := memberOplog0_3_3.MasterSigns
	assert.Equal(1, len(masterSigns0_3_3))
	assert.Equal(me0_3.ID, masterSigns0_3_3[0].ID)

	marshaled, _ = profile1_3_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMemberOplogList1_3_3 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMemberOplogList1_3_3, t, isDebug)
	assert.Equal(1, len(dataGetMemberOplogList1_3_3.Result))
	memberOplog1_3_3 := dataGetMemberOplogList1_3_3.Result[0]
	masterSigns1_3_3 := memberOplog1_3_3.MasterSigns
	assert.Equal(1, len(masterSigns1_3_3))
	assert.Equal(me1_3.ID, masterSigns1_3_3[0].ID)

	// 3.4. getUserOplogList
	marshaled, _ = profile0_3_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_3_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_3_4, t, isDebug)
	assert.Equal(2, len(dataGetUserOplogList0_3_4.Result))
	userOplog0_3_4 := dataGetUserOplogList0_3_4.Result[0]
	masterSigns0_3_4 := userOplog0_3_4.MasterSigns
	assert.Equal(1, len(masterSigns0_3_4))
	assert.Equal(me0_3.ID, masterSigns0_3_4[0].ID)

	marshaled, _ = profile1_3_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList1_3_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetUserOplogList1_3_4, t, isDebug)
	assert.Equal(2, len(dataGetUserOplogList1_3_4.Result))
	userOplog1_3_4 := dataGetUserOplogList1_3_4.Result[0]
	masterSigns1_3_4 := userOplog1_3_4.MasterSigns
	assert.Equal(1, len(masterSigns1_3_4))
	assert.Equal(me1_3.ID, masterSigns1_3_4[0].ID)

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
	bodyString = `{"id": "testID", "method": "me_getJoinKeyInfos", "params": []}`
	dataGetJoinKeys0_6_1 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetJoinKeys0_6_1, t, isDebug)
	assert.Equal(1, len(dataGetJoinKeys0_6_1.Result))

	// 7. join-me
	log.Debug("7. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey0_4)

	dataJoinMe0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(10 * time.Second)

	// 8. me_GetMyNodes
	log.Debug("8. me_GetMyNodes: start")
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
	log.Debug("8.1. getRawMe: start")
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": []}`

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

	// 9. MasterOplog
	bodyString = `{"id": "testID", "method": "me_getMasterOplogList", "params": ["", 0, 2]}`

	dataMasterOplogs0_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_9, t, isDebug)
	assert.Equal(3, len(dataMasterOplogs0_9.Result))
	masterOplog0_9 := dataMasterOplogs0_9.Result[0]
	assert.Equal(me1_3.ID[:common.AddressLength], masterOplog0_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_3.ID, masterOplog0_9.ObjID)
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
	assert.Equal(me1_3.ID[:common.AddressLength], masterOplog1_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_3.ID, masterOplog1_9.ObjID)
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

	// 9.1. getRawMeByID
	marshaled, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMeByID", "params": ["%v"]}`, string(marshaled))

	me0_9_1 := &me.MyInfo{}
	testCore(t0, bodyString, me0_9_1, t, isDebug)
	assert.Equal(types.StatusMigrated, me0_9_1.Status)
	assert.Equal(2, len(me0_9_1.OwnerIDs))
	assert.Equal(true, me0_9_1.IsOwner(me1_3.ID))
	assert.Equal(true, me0_9_1.IsOwner(me0_3.ID))

	// 9.2. MeOplog
	bodyString = `{"id": "testID", "method": "me_getMeOplogList", "params": ["", 0, 2]}`

	dataMeOplogs0_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMeOplogs0_9_2, t, isDebug)
	assert.Equal(1, len(dataMeOplogs0_9_2.Result))
	meOplog0_9_2 := dataMeOplogs0_9_2.Result[0]
	assert.Equal(me1_3.ID, meOplog0_9_2.CreatorID)
	assert.Equal(me1_3.ID, meOplog0_9_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog0_9_2.Op)
	assert.Equal(nilPttID, meOplog0_9_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog0_9_2.IsSync)
	assert.Equal(masterOplog1_9.ID, meOplog0_9_2.MasterLogID)
	assert.Equal(me1_3.LogID, meOplog0_9_2.ID)
	masterSign0_9_2 := meOplog0_9_2.MasterSigns[0]
	assert.Equal(nodeAddr1_1[:], masterSign0_9_2.ID[:common.AddressLength])
	assert.Equal(me1_3.ID[:common.AddressLength], masterSign0_9_2.ID[common.AddressLength:])
	assert.Equal(me0_8_1.LogID, meOplog0_9_2.ID)

	dataMeOplogs1_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMeOplogs1_9_2, t, isDebug)
	assert.Equal(1, len(dataMeOplogs1_9_2.Result))
	meOplog1_9_2 := dataMeOplogs1_9_2.Result[0]
	assert.Equal(me1_3.ID, meOplog1_9_2.CreatorID)
	assert.Equal(me1_3.ID, meOplog1_9_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog1_9_2.Op)
	assert.Equal(nilPttID, meOplog1_9_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog1_9_2.IsSync)
	assert.Equal(masterOplog1_9.ID, meOplog1_9_2.MasterLogID)
	assert.Equal(me1_3.LogID, meOplog1_9_2.ID)
	masterSign1_9_2 := meOplog1_9_2.MasterSigns[0]
	assert.Equal(nodeAddr1_1[:], masterSign1_9_2.ID[:common.AddressLength])
	assert.Equal(me1_3.ID[:common.AddressLength], masterSign1_9_2.ID[common.AddressLength:])
	assert.Equal(meOplog0_9_2, meOplog1_9_2)
	assert.Equal(me1_8_1.LogID, meOplog1_9_2.ID)

	// 10.1 getRawProfile
	t.Logf("10.1 getRawProfile")

	marshaled, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_10_1 := &account.Profile{}
	testCore(t0, bodyString, profile0_10_1, t, isDebug)
	assert.Equal(types.StatusDeleted, profile0_10_1.Status)
	assert.Equal(me0_3.ID, profile0_10_1.MyID)
	assert.Equal(me0_3.MyProfileID, profile0_10_1.ID)
	assert.Equal(false, profile0_10_1.IsOwner(me0_1.ID))
	assert.Equal(true, profile0_10_1.IsOwner(me1_1.ID))

	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile1_10_1 := &account.Profile{}
	testCore(t1, bodyString, profile1_10_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile1_10_1.Status)
	assert.Equal(me1_3.ID, profile1_10_1.MyID)
	assert.Equal(me1_3.MyProfileID, profile1_10_1.ID)

	// 10.2. getMasterOplogList
	t.Logf("10.2 getMasterOplogList")
	marshaled, _ = profile0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList0_10_2 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMasterOplogList0_10_2, t, isDebug)
	assert.Equal(2, len(dataGetMasterOplogList0_10_2.Result))
	masterOplog0_10_2 := dataGetMasterOplogList0_10_2.Result[0]
	masterSigns0_10_2 := masterOplog0_10_2.MasterSigns
	assert.Equal(1, len(masterSigns0_10_2))
	assert.Equal(me0_3.ID, masterSigns0_10_2[0].ID)
	assert.Equal(masterOplog0_10_2.ID, masterOplog0_10_2.MasterLogID)

	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList1_10_2 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMasterOplogList1_10_2, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogList1_10_2.Result))
	masterOplog1_10_2 := dataGetMasterOplogList1_10_2.Result[0]
	masterSigns1_10_2 := masterOplog1_10_2.MasterSigns
	assert.Equal(1, len(masterSigns0_10_2))
	assert.Equal(me1_3.ID, masterSigns1_10_2[0].ID)
	assert.Equal(masterOplog1_10_2.ID, masterOplog1_10_2.MasterLogID)
	assert.Equal(types.StatusAlive, masterOplog1_10_2.ToStatus())
	assert.Equal(types.Bool(true), masterOplog1_10_2.IsSync)

	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogList0_10_2_1 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMasterOplogList0_10_2_1, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogList0_10_2_1.Result))
	masterOplog0_10_2_1 := dataGetMasterOplogList0_10_2_1.Result[0]
	masterSigns0_10_2_1 := masterOplog0_10_2_1.MasterSigns
	assert.Equal(1, len(masterSigns0_10_2_1))
	assert.Equal(me1_3.ID, masterSigns0_10_2_1[0].ID)
	assert.Equal(masterOplog1_10_2.ID, masterOplog0_10_2_1.MasterLogID)
	assert.Equal(types.StatusAlive, masterOplog0_10_2_1.ToStatus())
	assert.Equal(types.Bool(true), masterOplog0_10_2.IsSync)

	// 10.3. getMemberOplogList
	t.Logf("10.3 getMemberOplogList")

	marshaled, _ = profile0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMembergOplogList0_10_3 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMembergOplogList0_10_3, t, isDebug)
	assert.Equal(3, len(dataGetMembergOplogList0_10_3.Result))
	memberOplog0_10_3 := dataGetMembergOplogList0_10_3.Result[0]
	masterSigns0_10_3 := memberOplog0_10_3.MasterSigns
	assert.Equal(1, len(masterSigns0_10_3))
	assert.Equal(me0_3.ID, masterSigns0_10_3[0].ID)

	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMemberOplogList1_10_3 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMemberOplogList1_10_3, t, isDebug)
	assert.Equal(1, len(dataGetMemberOplogList1_10_3.Result))
	memberOplog1_10_3 := dataGetMemberOplogList1_10_3.Result[0]
	masterSigns1_10_3 := memberOplog1_10_3.MasterSigns
	assert.Equal(1, len(masterSigns1_10_3))
	assert.Equal(me1_3.ID, masterSigns1_10_3[0].ID)

	// 10.4. getUserOplogList
	marshaled, _ = profile0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_10_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_10_4, t, isDebug)
	assert.Equal(3, len(dataGetUserOplogList0_10_4.Result))
	userOplog0_10_4 := dataGetUserOplogList0_10_4.Result[0]
	masterSigns0_10_4 := userOplog0_10_4.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4))
	assert.Equal(me0_3.ID, masterSigns0_10_4[0].ID)

	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList1_10_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetUserOplogList1_10_4, t, isDebug)
	assert.Equal(3, len(dataGetUserOplogList1_10_4.Result))
	userOplog1_10_4 := dataGetUserOplogList1_10_4.Result[0]
	masterSigns1_10_4 := userOplog1_10_4.MasterSigns
	assert.Equal(1, len(masterSigns1_10_4))
	assert.Equal(me1_3.ID, masterSigns1_10_4[0].ID)

	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_10_4_1 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_10_4_1, t, isDebug)
	assert.Equal(3, len(dataGetUserOplogList0_10_4_1.Result))
	userOplog0_10_4_1 := dataGetUserOplogList0_10_4_1.Result[0]
	masterSigns0_10_4_1 := userOplog0_10_4_1.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_1))
	assert.Equal(me1_3.ID, masterSigns0_10_4_1[0].ID)

	// 10.5. getMasterMerkleList
	t.Logf("10.5 getMasterMerkleList")
	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogMerkleNodeList", "params": ["%v", 1, "", 0, 2]}`, string(marshaled))
	dataGetMasterOplogMerkleNodeList0_10_5 := &struct {
		Result []*pkgservice.MerkleNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMasterOplogMerkleNodeList0_10_5, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogMerkleNodeList0_10_5.Result))

	dataGetMasterOplogMerkleNodeList1_10_5 := &struct {
		Result []*pkgservice.MerkleNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMasterOplogMerkleNodeList1_10_5, t, isDebug)
	assert.Equal(1, len(dataGetMasterOplogMerkleNodeList1_10_5.Result))

	// 10.6. getUserNode
	t.Logf("10.6 GetUserNodeList")
	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserNodeList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserNodeList0_10_6 := &struct {
		Result []*account.UserNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserNodeList0_10_6, t, isDebug)
	assert.Equal(2, len(dataGetUserNodeList0_10_6.Result))
	userNode0_10_6_0 := dataGetUserNodeList0_10_6.Result[0]
	userNode0_10_6_1 := dataGetUserNodeList0_10_6.Result[1]
	assert.Equal(types.StatusAlive, userNode0_10_6_0.Status)
	assert.Equal(types.StatusAlive, userNode0_10_6_1.Status)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserNodeList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserNodeList1_10_6 := &struct {
		Result []*account.UserNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetUserNodeList1_10_6, t, isDebug)
	assert.Equal(2, len(dataGetUserNodeList1_10_6.Result))
	userNode1_10_6_0 := dataGetUserNodeList1_10_6.Result[0]
	userNode1_10_6_1 := dataGetUserNodeList1_10_6.Result[1]
	assert.Equal(types.StatusAlive, userNode1_10_6_0.Status)
	assert.Equal(types.StatusAlive, userNode1_10_6_1.Status)

}
