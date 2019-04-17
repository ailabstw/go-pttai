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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceSyncFriend1(t *testing.T) {
	NNodes = 3
	isDebug := true

	var bodyString string
	var marshaled []byte
	//var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
	//var marshaledStr string
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")
	t2 := baloo.New("http://127.0.0.1:9452")

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

	me2_1 := &me.BackendMyInfo{}
	testCore(t2, bodyString, me2_1, t, isDebug)
	assert.Equal(types.StatusAlive, me2_1.Status)

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

	me2_3 := &me.MyInfo{}
	testCore(t2, bodyString, me2_3, t, isDebug)
	assert.Equal(types.StatusAlive, me2_3.Status)
	assert.Equal(me2_1.ID, me2_3.ID)
	assert.Equal(1, len(me2_3.OwnerIDs))
	assert.Equal(me2_3.ID, me2_3.OwnerIDs[0])
	assert.Equal(true, me2_3.IsOwner(me2_3.ID))

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

	// 7. join-me t0 signed-in as t1
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

	//masterOplog1_9_2 := dataMasterOplogs1_9.Result[2]

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

	// 10. show-url
	bodyString = `{"id": "testID", "method": "me_showURL", "params": []}`

	dataShowURL1_10 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowURL1_10, t, isDebug)
	url1_10 := dataShowURL1_10.URL

	// 11. join-friend t2 add t1(t0) as friend
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_10)

	dataJoinFriend2_11 := &pkgservice.BackendJoinRequest{}
	testCore(t2, bodyString, dataJoinFriend2_11, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend2_11.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend2_11.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(10 * time.Second)

	// 12. get-friend-list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList0_12 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetFriendList0_12, t, isDebug)
	assert.Equal(1, len(dataGetFriendList0_12.Result))
	friend0_12 := dataGetFriendList0_12.Result[0]
	assert.Equal(types.StatusAlive, friend0_12.Status)
	assert.Equal(me2_1.ID, friend0_12.FriendID)

	dataGetFriendList1_12 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_12, t, isDebug)
	assert.Equal(1, len(dataGetFriendList1_12.Result))
	friend1_12 := dataGetFriendList1_12.Result[0]
	assert.Equal(types.StatusAlive, friend1_12.Status)
	assert.Equal(me2_1.ID, friend1_12.FriendID)
	assert.Equal(friend0_12.ID, friend1_12.ID)

	dataGetFriendList2_12 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetFriendList2_12, t, isDebug)
	assert.Equal(1, len(dataGetFriendList2_12.Result))
	friend2_12 := dataGetFriendList2_12.Result[0]
	assert.Equal(types.StatusAlive, friend2_12.Status)
	assert.Equal(me1_1.ID, friend2_12.FriendID)
	assert.Equal(friend0_12.ID, friend2_12.ID)

	// 13. create-msg
	msg, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試1")),
		base64.StdEncoding.EncodeToString([]byte("測試2")),
		base64.StdEncoding.EncodeToString([]byte("測試3")),
		base64.StdEncoding.EncodeToString([]byte("測試4")),
		base64.StdEncoding.EncodeToString([]byte("測試5")),
		base64.StdEncoding.EncodeToString([]byte("測試6")),
		base64.StdEncoding.EncodeToString([]byte("測試7")),
		base64.StdEncoding.EncodeToString([]byte("測試8")),
		base64.StdEncoding.EncodeToString([]byte("測試9")),
		base64.StdEncoding.EncodeToString([]byte("測試10")),
		base64.StdEncoding.EncodeToString([]byte("測試11")),
		base64.StdEncoding.EncodeToString([]byte("測試12")),
		base64.StdEncoding.EncodeToString([]byte("測試13")),
		base64.StdEncoding.EncodeToString([]byte("測試14")),
		base64.StdEncoding.EncodeToString([]byte("測試15")),
		base64.StdEncoding.EncodeToString([]byte("測試16")),
		base64.StdEncoding.EncodeToString([]byte("測試17")),
		base64.StdEncoding.EncodeToString([]byte("測試18")),
		base64.StdEncoding.EncodeToString([]byte("測試19")),
		base64.StdEncoding.EncodeToString([]byte("測試20")),
		base64.StdEncoding.EncodeToString([]byte("測試21")),
		base64.StdEncoding.EncodeToString([]byte("測試22")),
	})

	// 35. friend-create-message
	marshaled, _ = friend0_12.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_createMessage", "params": ["%v", %v, []]}`, string(marshaled), string(msg))
	dataCreateMessage0_35 := &friend.BackendCreateMessage{}
	testCore(t0, bodyString, dataCreateMessage0_35, t, isDebug)
	assert.Equal(friend0_12.ID, dataCreateMessage0_35.FriendID)
	assert.Equal(2, dataCreateMessage0_35.NBlock)

	time.Sleep(20 * time.Second)

	// 36. friend-get-message-list
	marshaled, _ = friend0_12.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMessageList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMessageList0_36 := &struct {
		Result []*friend.BackendGetMessage `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMessageList0_36, t, isDebug)
	assert.Equal(1, len(dataGetMessageList0_36.Result))
	message0_36 := dataGetMessageList0_36.Result[0]

	dataGetMessageList1_36 := &struct {
		Result []*friend.BackendGetMessage `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMessageList1_36, t, isDebug)
	assert.Equal(1, len(dataGetMessageList1_36.Result))

	marshaled, _ = friend0_12.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMessageList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataGetMessageList2_36 := &struct {
		Result []*friend.BackendGetMessage `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetMessageList2_36, t, isDebug)
	assert.Equal(1, len(dataGetMessageList2_36.Result))
	// 37. get friend-oplog list
	marshaled, _ = friend0_12.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendOplogList", "params": ["%v", "", 10, 2]}`, string(marshaled))

	dataFriendOplog0_37 := &struct {
		Result []*friend.FriendOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataFriendOplog0_37, t, isDebug)
	assert.Equal(2, len(dataFriendOplog0_37.Result))
	friendOplog0_37_0 := dataFriendOplog0_37.Result[0]
	assert.Equal(types.StatusAlive, friendOplog0_37_0.ToStatus())
	assert.Equal(friend.FriendOpTypeCreateFriend, friendOplog0_37_0.Op)
	friendOplog0_37_1 := dataFriendOplog0_37.Result[1]
	assert.Equal(types.StatusAlive, friendOplog0_37_1.ToStatus())
	assert.Equal(friend.FriendOpTypeCreateMessage, friendOplog0_37_1.Op)

	dataFriendOplog1_37 := &struct {
		Result []*friend.FriendOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataFriendOplog1_37, t, isDebug)
	assert.Equal(2, len(dataFriendOplog1_37.Result))

	assert.Equal(dataFriendOplog0_37, dataFriendOplog1_37)

	// // 38. get-message-block
	marshaledID2, _ = message0_36.ID.MarshalText()
	marshaledID3, _ = message0_36.BlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMessageBlockList", "params": ["%v", "%v", "%v", 0, 0, 10]}`, string(marshaled), string(marshaledID2), string(marshaledID3))

	dataGetMessageBlockList0_37 := &struct {
		Result []*friend.BackendMessageBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMessageBlockList0_37, t, isDebug)
	assert.Equal(2, len(dataGetMessageBlockList0_37.Result))

	article0 := [][]byte{
		[]byte("測試1"),
		[]byte("測試2"),
		[]byte("測試3"),
		[]byte("測試4"),
		[]byte("測試5"),
		[]byte("測試6"),
		[]byte("測試7"),
		[]byte("測試8"),
		[]byte("測試9"),
		[]byte("測試10"),
		[]byte("測試11"),
		[]byte("測試12"),
		[]byte("測試13"),
		[]byte("測試14"),
		[]byte("測試15"),
		[]byte("測試16"),
		[]byte("測試17"),
		[]byte("測試18"),
		[]byte("測試19"),
		[]byte("測試20"),
	}

	article1 := [][]byte{
		[]byte("測試21"),
		[]byte("測試22"),
	}

	assert.Equal(article0, dataGetMessageBlockList0_37.Result[0].Buf)
	assert.Equal(article1, dataGetMessageBlockList0_37.Result[1].Buf)

	dataGetMessageBlockList1_37 := &struct {
		Result []*friend.BackendMessageBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMessageBlockList1_37, t, isDebug)
	assert.Equal(2, len(dataGetMessageBlockList1_37.Result))

	assert.Equal(article0, dataGetMessageBlockList1_37.Result[0].Buf)
	assert.Equal(article1, dataGetMessageBlockList1_37.Result[1].Buf)

	dataGetMessageBlockList2_37 := &struct {
		Result []*friend.BackendMessageBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetMessageBlockList2_37, t, isDebug)
	assert.Equal(2, len(dataGetMessageBlockList2_37.Result))

	assert.Equal(article0, dataGetMessageBlockList2_37.Result[0].Buf)
	assert.Equal(article1, dataGetMessageBlockList2_37.Result[1].Buf)
}
