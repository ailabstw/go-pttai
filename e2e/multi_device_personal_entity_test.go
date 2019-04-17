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
	"reflect"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDevicePersonalEntity(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
	var marshaledID []byte
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
	boardID0_3 := me0_3.BoardID

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))

	profileID1_3 := me1_3.ProfileID
	boardID1_3 := me1_3.BoardID

	// 4. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_4 string

	testCore(t0, bodyString, &myKey0_4, t, isDebug)
	if isDebug {
		t.Logf("myKey0_4: %v\n", myKey0_4)
	}

	// 4.1. profile
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaledID))

	profile0_4_1 := &account.Profile{}
	testCore(t0, bodyString, profile0_4_1, t, isDebug)

	assert.Equal(types.StatusAlive, profile0_4_1.Status)

	// t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaledID))

	profile1_4_1 := &account.Profile{}
	testCore(t1, bodyString, profile1_4_1, t, isDebug)

	assert.Equal(types.StatusAlive, profile1_4_1.Status)

	// 4.1. profile-member: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMembers0_4_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMembers0_4_1, t, isDebug)

	assert.Equal(1, len(dataMembers0_4_1.Result))
	member0_4_1_0 := dataMembers0_4_1.Result[0]
	assert.Equal(me0_3.ID, member0_4_1_0.ID)
	assert.Equal(types.StatusAlive, member0_4_1_0.Status)

	// 4.1. profile-member: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMembers1_4_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMembers1_4_1, t, isDebug)

	assert.Equal(1, len(dataMembers1_4_1.Result))
	member1_4_1_0 := dataMembers1_4_1.Result[0]
	assert.Equal(me1_3.ID, member1_4_1_0.ID)
	assert.Equal(types.StatusAlive, member1_4_1_0.Status)

	// 4.1.1. profile-member-oplog: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberOplogs0_4_1_1 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogs0_4_1_1, t, isDebug)

	assert.Equal(1, len(dataMemberOplogs0_4_1_1.Result))
	memberOplog0_4_1_1_0 := dataMemberOplogs0_4_1_1.Result[0]
	assert.Equal(me0_3.ID, memberOplog0_4_1_1_0.ObjID)
	assert.Equal(pkgservice.MemberOpTypeAddMember, memberOplog0_4_1_1_0.Op)
	assert.Equal(types.StatusAlive, memberOplog0_4_1_1_0.ToStatus())

	// 4.1.2. profile-member-oplog: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberOplogs1_4_1_1 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberOplogs1_4_1_1, t, isDebug)

	assert.Equal(1, len(dataMemberOplogs1_4_1_1.Result))
	memberOplog1_4_1_1_0 := dataMemberOplogs1_4_1_1.Result[0]
	assert.Equal(me1_3.ID, memberOplog1_4_1_1_0.ObjID)
	assert.Equal(pkgservice.MemberOpTypeAddMember, memberOplog1_4_1_1_0.Op)
	assert.Equal(types.StatusAlive, memberOplog1_4_1_1_0.ToStatus())

	// 4.1.2. profile-master: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasters0_4_1_2 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasters0_4_1_2, t, isDebug)

	assert.Equal(1, len(dataMasters0_4_1_2.Result))
	master0_4_1_2_0 := dataMasters0_4_1_2.Result[0]
	assert.Equal(me0_3.ID, master0_4_1_2_0.ID)
	assert.Equal(types.StatusAlive, master0_4_1_2_0.Status)

	// 4.1.2 profile-master: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasters1_4_1_2 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasters1_4_1_2, t, isDebug)

	assert.Equal(1, len(dataMasters1_4_1_2.Result))
	master1_4_1_2_0 := dataMasters1_4_1_2.Result[0]
	assert.Equal(me1_3.ID, master1_4_1_2_0.ID)
	assert.Equal(types.StatusAlive, master1_4_1_2_0.Status)

	// 4.1.3. profile-master-oplog: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterOplogs0_4_1_3 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_4_1_3, t, isDebug)

	assert.Equal(1, len(dataMasterOplogs0_4_1_3.Result))
	masterOplog0_4_1_3_0 := dataMasterOplogs0_4_1_3.Result[0]
	assert.Equal(me0_3.ID, masterOplog0_4_1_3_0.ObjID)
	assert.Equal(pkgservice.MasterOpTypeAddMaster, masterOplog0_4_1_3_0.Op)
	assert.Equal(types.StatusAlive, masterOplog0_4_1_3_0.ToStatus())

	// 4.1.3. profile-master-oplog: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterOplogs1_4_1_3 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogs1_4_1_3, t, isDebug)

	assert.Equal(1, len(dataMasterOplogs1_4_1_3.Result))
	masterOplog1_4_1_3_0 := dataMasterOplogs1_4_1_3.Result[0]
	assert.Equal(me1_3.ID, masterOplog1_4_1_3_0.ObjID)
	assert.Equal(pkgservice.MasterOpTypeAddMaster, masterOplog1_4_1_3_0.Op)
	assert.Equal(types.StatusAlive, masterOplog1_4_1_3_0.ToStatus())

	// 4.2. board
	marshaledID, _ = boardID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board0_4_2 := &content.Board{}
	testCore(t0, bodyString, board0_4_2, t, isDebug)

	assert.Equal(types.StatusAlive, board0_4_2.Status)

	// t1
	marshaledID, _ = boardID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board1_4_2 := &content.Board{}
	testCore(t1, bodyString, board1_4_2, t, isDebug)

	assert.Equal(types.StatusAlive, board1_4_2.Status)

	// 4.2. board-member
	marshaledID, _ = boardID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMembers0_4_2 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMembers0_4_2, t, isDebug)

	assert.Equal(1, len(dataMembers0_4_2.Result))
	member0_4_2_0 := dataMembers0_4_2.Result[0]
	assert.Equal(me0_3.ID, member0_4_2_0.ID)
	assert.Equal(types.StatusAlive, member0_4_2_0.Status)

	// t1
	marshaledID, _ = boardID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMembers1_4_2 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMembers1_4_2, t, isDebug)

	assert.Equal(1, len(dataMembers1_4_2.Result))
	member1_4_2_0 := dataMembers1_4_2.Result[0]
	assert.Equal(me1_3.ID, member1_4_2_0.ID)
	assert.Equal(types.StatusAlive, member1_4_2_0.Status)

	// 4.2.1. board-member-oplog
	marshaledID, _ = boardID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberOplogs0_4_2_1 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogs0_4_2_1, t, isDebug)

	assert.Equal(1, len(dataMemberOplogs0_4_2_1.Result))
	memberOplog0_4_2_1_0 := dataMemberOplogs0_4_2_1.Result[0]
	assert.Equal(me0_3.ID, memberOplog0_4_2_1_0.ObjID)
	assert.Equal(pkgservice.MemberOpTypeAddMember, memberOplog0_4_2_1_0.Op)
	assert.Equal(types.StatusAlive, memberOplog0_4_2_1_0.ToStatus())

	// 4.2.2. board-master
	marshaledID, _ = boardID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasters0_4_2_2 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasters0_4_2_2, t, isDebug)

	assert.Equal(1, len(dataMasters0_4_2_2.Result))
	master0_4_2_2_0 := dataMasters0_4_2_2.Result[0]
	assert.Equal(me0_3.ID, master0_4_2_2_0.ID)
	assert.Equal(types.StatusAlive, master0_4_2_2_0.Status)

	// 4.2.3. board-master-oplog
	marshaledID, _ = boardID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterOplogs0_4_2_3 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_4_2_3, t, isDebug)

	assert.Equal(1, len(dataMasterOplogs0_4_2_3.Result))
	masterOplog0_4_2_3_0 := dataMasterOplogs0_4_2_3.Result[0]
	assert.Equal(me0_3.ID, masterOplog0_4_2_3_0.ObjID)
	assert.Equal(pkgservice.MasterOpTypeAddMaster, masterOplog0_4_2_3_0.Op)
	assert.Equal(types.StatusAlive, masterOplog0_4_2_3_0.ToStatus())

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
	t.Logf("wait 15 seconds for hand-shaking")
	time.Sleep(TimeSleepRestart)

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

	// 8.2. get-raw-board
	marshaled, _ = me0_3.BoardID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaled))

	board0_8_2 := &content.Board{}
	testCore(t0, bodyString, board0_8_2, t, isDebug)
	assert.Equal(me0_1.ID, board0_8_2.CreatorID)
	assert.Equal(me1_1.ID, board0_8_2.OwnerIDs[0])

	// 8.3. get-raw-board
	marshaled, _ = me1_3.BoardID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaled))

	board0_8_3 := &content.Board{}
	testCore(t0, bodyString, board0_8_3, t, isDebug)
	assert.Equal(me1_1.ID, board0_8_3.CreatorID)
	assert.Equal(me1_1.ID, board0_8_3.OwnerIDs[0])

	board1_8_3 := &content.Board{}
	testCore(t1, bodyString, board1_8_3, t, isDebug)
	assert.Equal(me1_1.ID, board1_8_3.CreatorID)
	assert.Equal(me1_1.ID, board1_8_3.OwnerIDs[0])

	// 8.4. get-raw-account
	t.Logf("8.4 get-raw-account: profileID0_3")
	marshaled, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_8_4 := &account.Profile{}
	testCore(t0, bodyString, profile0_8_4, t, isDebug)
	assert.Equal(me0_1.ID, profile0_8_4.CreatorID)
	assert.Equal(me1_1.ID, profile0_8_4.OwnerIDs[0])
	assert.Equal(types.StatusTerminal, profile0_8_4.Status)

	profile1_8_4 := &account.Profile{}
	testCore(t1, bodyString, profile1_8_4, t, isDebug)
	// assert.Equal(nilProfile, profile1_8_4)

	// 8.5. get-raw-account
	t.Logf("8.4 get-raw-account: profileID1_3")
	marshaled, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_8_5 := &account.Profile{}
	testCore(t0, bodyString, profile0_8_5, t, isDebug)
	assert.Equal(me1_1.ID, profile0_8_5.CreatorID)
	assert.Equal(me1_1.ID, profile0_8_5.OwnerIDs[0])
	assert.Equal(types.StatusAlive, profile0_8_5.Status)

	profile1_8_5 := &account.Profile{}
	testCore(t1, bodyString, profile1_8_5, t, isDebug)
	assert.Equal(me1_1.ID, profile1_8_5.CreatorID)
	assert.Equal(me1_1.ID, profile1_8_5.OwnerIDs[0])
	assert.Equal(types.StatusAlive, profile0_8_5.Status)

	// 9. MasterOplog
	t.Logf("getMyMasterOplogList")
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

	// 9.1. getRawMe
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

	// 10.1. profile-member: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMembers0_10_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMembers0_10_1, t, isDebug)

	assert.Equal(2, len(dataMembers0_10_1.Result))
	member0_10_1_0 := dataMembers0_10_1.Result[0]
	member0_10_1_1 := dataMembers0_10_1.Result[1]

	member0_10_1_orig := member0_10_1_0
	member0_10_1_new := member0_10_1_1

	if reflect.DeepEqual(me1_1.ID, member0_10_1_orig.ID) {
		member0_10_1_orig = member0_10_1_1
		member0_10_1_new = member0_10_1_0
	}

	assert.Equal(me0_3.ID, member0_10_1_orig.ID)
	assert.Equal(types.StatusMigrated, member0_10_1_orig.Status)

	assert.Equal(me1_3.ID, member0_10_1_new.ID)
	assert.Equal(types.StatusAlive, member0_10_1_new.Status)

	// 10.1. profile-member: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMembers0_10_1 = &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMembers0_10_1, t, isDebug)

	assert.Equal(1, len(dataMembers0_10_1.Result))
	member0_10_1_0 = dataMembers0_10_1.Result[0]
	assert.Equal(me1_3.ID, member0_10_1_0.ID)
	assert.Equal(types.StatusAlive, member0_10_1_0.Status)

	dataMembers1_10_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMembers1_10_1, t, isDebug)

	assert.Equal(1, len(dataMembers1_10_1.Result))
	member1_10_1_0 := dataMembers1_10_1.Result[0]
	assert.Equal(me1_3.ID, member1_10_1_0.ID)
	assert.Equal(types.StatusAlive, member1_10_1_0.Status)

	// 10.1.1. profile-member-oplog: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberOplogs0_10_1_1 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogs0_10_1_1, t, isDebug)

	assert.Equal(3, len(dataMemberOplogs0_10_1_1.Result))
	assert.Equal(dataMemberOplogs0_4_1_1.Result, dataMemberOplogs0_10_1_1.Result[:1])

	memberOplog0_10_1_1_1 := dataMemberOplogs0_10_1_1.Result[1]
	assert.Equal(me1_3.ID, memberOplog0_10_1_1_1.ObjID)
	assert.Equal(pkgservice.MemberOpTypeAddMember, memberOplog0_10_1_1_1.Op)
	assert.Equal(types.StatusAlive, memberOplog0_10_1_1_1.ToStatus())

	memberOplog0_10_1_1_2 := dataMemberOplogs0_10_1_1.Result[2]
	assert.Equal(me0_3.ID, memberOplog0_10_1_1_2.ObjID)
	assert.Equal(pkgservice.MemberOpTypeMigrateMember, memberOplog0_10_1_1_2.Op)
	assert.Equal(types.StatusAlive, memberOplog0_10_1_1_2.ToStatus())

	// 10.1.2. profile-member-oplog: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberOplogs0_10_1_2 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogs0_10_1_2, t, isDebug)

	assert.Equal(1, len(dataMemberOplogs0_10_1_2.Result))
	assert.Equal(dataMemberOplogs1_4_1_1.Result, dataMemberOplogs0_10_1_2.Result)

	dataMemberOplogs1_10_1_2 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberOplogs1_10_1_2, t, isDebug)

	assert.Equal(1, len(dataMemberOplogs1_10_1_2.Result))
	assert.Equal(dataMemberOplogs1_4_1_1.Result, dataMemberOplogs1_10_1_2.Result)

	// 10.1.3. profile-master: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasters0_10_1_3 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasters0_10_1_3, t, isDebug)

	assert.Equal(2, len(dataMasters0_10_1_3.Result))
	master0_10_1_3_0 := dataMasters0_10_1_3.Result[0]
	master0_10_1_3_1 := dataMasters0_10_1_3.Result[1]

	master0_10_1_orig := master0_10_1_3_0
	master0_10_1_new := master0_10_1_3_1

	if reflect.DeepEqual(me1_1.ID, master0_10_1_orig.ID) {
		master0_10_1_orig = master0_10_1_3_1
		master0_10_1_new = master0_10_1_3_0
	}

	assert.Equal(me0_3.ID, master0_10_1_orig.ID)
	assert.Equal(types.StatusMigrated, master0_10_1_orig.Status)

	assert.Equal(me1_3.ID, master0_10_1_new.ID)
	assert.Equal(types.StatusAlive, master0_10_1_new.Status)

	// 10.1.4 profile-master: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasters0_10_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasters0_10_1, t, isDebug)

	assert.Equal(1, len(dataMasters0_10_1.Result))
	master0_10_1_0 := dataMasters0_10_1.Result[0]
	assert.Equal(me1_3.ID, master0_10_1_0.ID)
	assert.Equal(types.StatusAlive, master0_10_1_0.Status)

	dataMasters1_10_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasters1_10_1, t, isDebug)

	assert.Equal(1, len(dataMasters1_10_1.Result))
	master1_10_1_0 := dataMasters1_10_1.Result[0]
	assert.Equal(me1_3.ID, master1_10_1_0.ID)
	assert.Equal(types.StatusAlive, master1_10_1_0.Status)

	// 10.1.5. profile-master-oplog: t0
	marshaledID, _ = profileID0_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterOplogs0_10_1_5 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_10_1_5, t, isDebug)

	assert.Equal(2, len(dataMasterOplogs0_10_1_5.Result))
	assert.Equal(dataMasterOplogs0_4_1_3.Result, dataMasterOplogs0_10_1_5.Result[:1])
	masterOplog0_10_1_5_1 := dataMasterOplogs0_10_1_5.Result[1]
	assert.Equal(me0_3.ID, masterOplog0_10_1_5_1.ObjID)
	assert.Equal(pkgservice.MasterOpTypeMigrateMaster, masterOplog0_10_1_5_1.Op)
	assert.Equal(types.StatusAlive, masterOplog0_10_1_5_1.ToStatus())

	// 4.1.6. profile-master-oplog: t1
	marshaledID, _ = profileID1_3.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterOplogs0_10_1_6 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_10_1_6, t, isDebug)

	assert.Equal(1, len(dataMasterOplogs0_10_1_6.Result))
	assert.Equal(dataMasterOplogs1_4_1_3.Result, dataMasterOplogs0_10_1_6.Result)

	dataMasterOplogs1_10_1_6 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogs1_10_1_6, t, isDebug)

	assert.Equal(1, len(dataMasterOplogs1_10_1_6.Result))
	assert.Equal(dataMasterOplogs1_4_1_3.Result, dataMasterOplogs1_10_1_6.Result)
}
