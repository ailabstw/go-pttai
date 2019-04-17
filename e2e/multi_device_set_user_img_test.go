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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceSetUserImg(t *testing.T) {
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

	// 3.1 getRawProfile
	marshaled, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_3_1 := &account.Profile{}
	testCore(t0, bodyString, profile0_3_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile0_3_1.Status)
	assert.Equal(me0_3.ID, profile0_3_1.MyID)
	assert.Equal(me0_3.ProfileID, profile0_3_1.ID)

	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile1_3_1 := &account.Profile{}
	testCore(t1, bodyString, profile1_3_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile1_3_1.Status)
	assert.Equal(me1_3.ID, profile1_3_1.MyID)
	assert.Equal(me1_3.ProfileID, profile1_3_1.ID)

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
	assert.Equal(5, len(dataGetUserOplogList0_3_4.Result))
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
	assert.Equal(5, len(dataGetUserOplogList1_3_4.Result))
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
	bodyString = `{"id": "testID", "method": "me_getJoinKeyInfos", "params": [""]}`
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
	time.Sleep(TimeSleepRestart)

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

	// 9. MasterOplog
	bodyString = `{"id": "testID", "method": "me_getMyMasterOplogList", "params": ["", "", 0, 2]}`

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
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaled))

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
	assert.Equal(me0_3.ProfileID, profile0_10_1.ID)
	assert.Equal(false, profile0_10_1.IsOwner(me0_1.ID))
	assert.Equal(true, profile0_10_1.IsOwner(me1_1.ID))

	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile1_10_1 := &account.Profile{}
	testCore(t1, bodyString, profile1_10_1, t, isDebug)
	assert.Equal(types.StatusAlive, profile1_10_1.Status)
	assert.Equal(me1_3.ID, profile1_10_1.MyID)
	assert.Equal(me1_3.ProfileID, profile1_10_1.ID)

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

	// 10.4. getUserOplogList
	t.Logf("10.4 GetUserOplogList: t0")
	marshaled, _ = profile0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_10_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_10_4, t, isDebug)
	assert.Equal(6, len(dataGetUserOplogList0_10_4.Result))
	userOplog0_10_4_0 := dataGetUserOplogList0_10_4.Result[0]
	masterSigns0_10_4_0 := userOplog0_10_4_0.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_0))
	assert.Equal(me0_3.ID, masterSigns0_10_4_0[0].ID)
	assert.Equal(masterOplog0_10_2.ID, userOplog0_10_4_0.MasterLogID)
	assert.Equal(account.UserOpTypeCreateProfile, userOplog0_10_4_0.Op)

	userOplog0_10_4_1 := dataGetUserOplogList0_10_4.Result[1]
	masterSigns0_10_4_1 := userOplog0_10_4_1.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_1))
	assert.Equal(me0_3.ID, masterSigns0_10_4_1[0].ID)
	assert.Equal(masterOplog0_10_2.ID, userOplog0_10_4_1.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserName, userOplog0_10_4_1.Op)

	userOplog0_10_4_2 := dataGetUserOplogList0_10_4.Result[2]
	masterSigns0_10_4_2 := userOplog0_10_4_2.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_2))
	assert.Equal(me0_3.ID, masterSigns0_10_4_2[0].ID)
	assert.Equal(masterOplog0_10_2.ID, userOplog0_10_4_2.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserImg, userOplog0_10_4_2.Op)

	userOplog0_10_4_4 := dataGetUserOplogList0_10_4.Result[4]
	masterSigns0_10_4_4 := userOplog0_10_4_4.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_4))
	assert.Equal(me0_3.ID, masterSigns0_10_4_4[0].ID)
	assert.Equal(masterOplog0_10_2.ID, userOplog0_10_4_4.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog0_10_4_4.Op)

	userOplog0_10_4_5 := dataGetUserOplogList0_10_4.Result[5]
	masterSigns0_10_4_5 := userOplog0_10_4_5.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_5))
	assert.Equal(me0_3.ID, masterSigns0_10_4_5[0].ID)
	assert.Equal(masterOplog0_10_2.ID, userOplog0_10_4_5.MasterLogID)
	assert.Equal(account.UserOpTypeDeleteProfile, userOplog0_10_4_5.Op)

	// t1
	t.Logf("10.4 GetUserOplogList: t1")
	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList1_10_4 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetUserOplogList1_10_4, t, isDebug)
	assert.Equal(6, len(dataGetUserOplogList1_10_4.Result))
	userOplog1_10_4_0 := dataGetUserOplogList1_10_4.Result[0]
	masterSigns1_10_4_0 := userOplog1_10_4_0.MasterSigns
	assert.Equal(1, len(masterSigns1_10_4_0))
	assert.Equal(me1_3.ID, masterSigns1_10_4_0[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog1_10_4_0.MasterLogID)
	assert.Equal(account.UserOpTypeCreateProfile, userOplog1_10_4_0.Op)

	userOplog1_10_4_1 := dataGetUserOplogList1_10_4.Result[1]
	masterSigns1_10_4_1 := userOplog1_10_4_1.MasterSigns
	assert.Equal(1, len(masterSigns1_10_4_1))
	assert.Equal(me1_3.ID, masterSigns1_10_4_1[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog1_10_4_1.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserName, userOplog1_10_4_1.Op)

	userOplog1_10_4_2 := dataGetUserOplogList1_10_4.Result[2]
	masterSigns1_10_4_2 := userOplog1_10_4_2.MasterSigns
	assert.Equal(1, len(masterSigns1_10_4_2))
	assert.Equal(me1_3.ID, masterSigns1_10_4_2[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog1_10_4_2.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserImg, userOplog1_10_4_2.Op)

	userOplog1_10_4_4 := dataGetUserOplogList1_10_4.Result[4]
	masterSigns1_10_4_4 := userOplog1_10_4_4.MasterSigns
	assert.Equal(1, len(masterSigns1_10_4_4))
	assert.Equal(me1_3.ID, masterSigns1_10_4_4[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog1_10_4_4.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog1_10_4_4.Op)

	userOplog1_10_4_5 := dataGetUserOplogList1_10_4.Result[5]
	masterSigns1_10_4_5 := userOplog1_10_4_5.MasterSigns
	assert.Equal(1, len(masterSigns1_10_4_5))
	assert.Equal(me1_3.ID, masterSigns1_10_4_5[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog1_10_4_5.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog1_10_4_5.Op)

	// new t0 (should be the same as t1)
	t.Logf("10.4 GetUserOplogList: new t0")
	marshaled, _ = profile1_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getUserOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetUserOplogList0_10_4_1 := &struct {
		Result []*account.UserOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetUserOplogList0_10_4_1, t, isDebug)
	assert.Equal(6, len(dataGetUserOplogList0_10_4_1.Result))
	userOplog0_10_4_1_0 := dataGetUserOplogList0_10_4_1.Result[0]
	mastersigns0_10_4_1_0 := userOplog0_10_4_1_0.MasterSigns
	assert.Equal(1, len(mastersigns0_10_4_1_0))
	assert.Equal(me1_3.ID, mastersigns0_10_4_1_0[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog0_10_4_1_0.MasterLogID)
	assert.Equal(account.UserOpTypeCreateProfile, userOplog0_10_4_1_0.Op)

	userOplog0_10_4_1_1 := dataGetUserOplogList0_10_4_1.Result[1]
	masterSigns0_10_4_1_1 := userOplog0_10_4_1_1.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_1_1))
	assert.Equal(me1_3.ID, masterSigns0_10_4_1_1[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog0_10_4_1_1.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserName, userOplog0_10_4_1_1.Op)

	userOplog0_10_4_1_2 := dataGetUserOplogList0_10_4_1.Result[2]
	masterSigns0_10_4_1_2 := userOplog0_10_4_1_2.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_1_2))
	assert.Equal(me1_3.ID, masterSigns0_10_4_1_2[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog0_10_4_1_2.MasterLogID)
	assert.Equal(account.UserOpTypeCreateUserImg, userOplog0_10_4_1_2.Op)

	userOplog0_10_4_1_4 := dataGetUserOplogList0_10_4_1.Result[4]
	masterSigns0_10_4_1_4 := userOplog0_10_4_1_4.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_1_4))
	assert.Equal(me1_3.ID, masterSigns0_10_4_1_4[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog0_10_4_1_4.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog0_10_4_1_4.Op)
	opData0_10_4_1_3 := &account.UserOpAddUserNode{}
	userOplog0_10_4_1_4.GetData(opData0_10_4_1_3)
	assert.Equal(me1_1.NodeID, opData0_10_4_1_3.NodeID)

	userOplog0_10_4_1_5 := dataGetUserOplogList0_10_4_1.Result[5]
	masterSigns0_10_4_1_5 := userOplog0_10_4_1_5.MasterSigns
	assert.Equal(1, len(masterSigns0_10_4_1_5))
	assert.Equal(me1_3.ID, masterSigns0_10_4_1_5[0].ID)
	assert.Equal(masterOplog1_10_2.ID, userOplog0_10_4_1_5.MasterLogID)
	assert.Equal(account.UserOpTypeAddUserNode, userOplog0_10_4_1_5.Op)
	opData0_10_4_1_4 := &account.UserOpAddUserNode{}
	userOplog0_10_4_1_5.GetData(opData0_10_4_1_4)
	assert.Equal(me0_1.NodeID, opData0_10_4_1_4.NodeID)

	assert.Equal(dataGetUserOplogList1_10_4, dataGetUserOplogList0_10_4_1)

	// 16. my_setMyImg
	normalizedImgStr := "data:image/png;base64,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"

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_setMyImage", "params": ["%v"]}`, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==")
	t.Logf("setMyImage: bodyString: %v", bodyString)
	dataSetMyImg0_16 := &account.UserImg{}
	testCore(t0, bodyString, dataSetMyImg0_16, t, false)
	assert.Equal(me1_1.ID, dataSetMyImg0_16.ID)

	t.Logf("wait 10 seconds for sync")
	time.Sleep(10 * time.Second)

	// 17. get user img
	marshaled, _ = me1_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawUserImg", "params": ["%v"]}`, string(marshaled))

	dataGetUserImg0_17 := &account.UserImg{}
	testCore(t0, bodyString, dataGetUserImg0_17, t, isDebug)
	assert.Equal(me1_1.ID, dataGetUserImg0_17.ID)
	assert.Equal(account.ImgTypePNG, dataGetUserImg0_17.ImgType)
	assert.Equal(uint16(account.MaxProfileImgWidth), dataGetUserImg0_17.Width)
	assert.Equal(uint16(account.MaxProfileImgHeight), dataGetUserImg0_17.Height)
	assert.Equal(normalizedImgStr, dataGetUserImg0_17.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg0_17.Status)

	var nilSyncUserImgInfo *account.SyncUserImgInfo
	assert.Equal(nilSyncUserImgInfo, dataGetUserImg0_17.SyncInfo)

	dataGetUserImg1_17 := &account.UserImg{}
	testCore(t1, bodyString, dataGetUserImg1_17, t, isDebug)
	assert.Equal(me1_1.ID, dataGetUserImg1_17.ID)
	assert.Equal(account.ImgTypePNG, dataGetUserImg1_17.ImgType)
	assert.Equal(uint16(account.MaxProfileImgWidth), dataGetUserImg1_17.Width)
	assert.Equal(uint16(account.MaxProfileImgHeight), dataGetUserImg1_17.Height)
	assert.Equal(normalizedImgStr, dataGetUserImg1_17.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg1_17.Status)

	assert.Equal(nilSyncUserImgInfo, dataGetUserImg1_17.SyncInfo)
}
