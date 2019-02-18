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

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestFriend4GetListByMsgCreateTs(t *testing.T) {
	NNodes = 4
	isDebug := true

	var bodyString string
	var marshaled []byte
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")
	t2 := baloo.New("http://127.0.0.1:9452")
	t3 := baloo.New("http://127.0.0.1:9453")

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}
	testCore(t0, bodyString, me0_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_1.Status)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)
	//nodeID1_1 := me1_1.NodeID
	//pubKey1_1, _ := nodeID1_1.Pubkey()
	// nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

	me2_1 := &me.BackendMyInfo{}
	testCore(t2, bodyString, me2_1, t, isDebug)
	assert.Equal(types.StatusAlive, me2_1.Status)

	me3_1 := &me.BackendMyInfo{}
	testCore(t3, bodyString, me3_1, t, isDebug)
	assert.Equal(types.StatusAlive, me3_1.Status)

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

	me3_3 := &me.MyInfo{}
	testCore(t3, bodyString, me3_3, t, isDebug)
	assert.Equal(types.StatusAlive, me3_3.Status)
	assert.Equal(me3_1.ID, me3_3.ID)
	assert.Equal(1, len(me3_3.OwnerIDs))
	assert.Equal(me3_3.ID, me3_3.OwnerIDs[0])
	assert.Equal(true, me3_3.IsOwner(me3_3.ID))

	// 5. show-url
	bodyString = `{"id": "testID", "method": "me_showURL", "params": []}`

	dataShowURL1_5 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowURL1_5, t, isDebug)
	url1_5 := dataShowURL1_5.URL

	// 7. join-friend: t0
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_5)

	dataJoinFriend0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinFriend0_7, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend0_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(15 * time.Second)

	// 7.1. join-friend: t2
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_5)

	dataJoinFriend2_7 := &pkgservice.BackendJoinRequest{}
	testCore(t2, bodyString, dataJoinFriend2_7, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend2_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend2_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(15 * time.Second)

	// 7.2. join-friend: t3
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_5)

	dataJoinFriend3_7 := &pkgservice.BackendJoinRequest{}
	testCore(t3, bodyString, dataJoinFriend3_7, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend3_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend3_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(15 * time.Second)

	// 8. get-friend-list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList0_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetFriendList0_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList0_8.Result))
	friend0_8 := dataGetFriendList0_8.Result[0]
	assert.Equal(types.StatusAlive, friend0_8.Status)
	assert.Equal(me1_1.ID, friend0_8.FriendID)

	dataGetFriendList1_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_8, t, isDebug)
	assert.Equal(3, len(dataGetFriendList1_8.Result))
	friend1_8_0 := dataGetFriendList1_8.Result[0]
	assert.Equal(types.StatusAlive, friend1_8_0.Status)
	assert.Equal(me0_1.ID, friend1_8_0.FriendID)
	assert.Equal(friend0_8.ID, friend1_8_0.ID)
	friend1_8_2 := dataGetFriendList1_8.Result[1]
	assert.Equal(types.StatusAlive, friend1_8_2.Status)
	assert.Equal(me2_1.ID, friend1_8_2.FriendID)
	friend1_8_3 := dataGetFriendList1_8.Result[2]
	assert.Equal(types.StatusAlive, friend1_8_3.Status)
	assert.Equal(me3_1.ID, friend1_8_3.FriendID)

	dataGetFriendList2_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetFriendList2_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList2_8.Result))
	friend2_8 := dataGetFriendList2_8.Result[0]
	assert.Equal(types.StatusAlive, friend2_8.Status)
	assert.Equal(me1_1.ID, friend2_8.FriendID)

	dataGetFriendList3_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t3, bodyString, dataGetFriendList3_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList3_8.Result))
	friend3_8 := dataGetFriendList3_8.Result[0]
	assert.Equal(types.StatusAlive, friend3_8.Status)
	assert.Equal(me1_1.ID, friend3_8.FriendID)

	// 8.1. get-friend-list-by-msg-create-ts
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendListByMsgCreateTS", "params": ["", 0, 1]}`)

	dataGetFriendList1_8_1 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_8_1, t, isDebug)
	assert.Equal(3, len(dataGetFriendList1_8_1.Result))
	friend1_8_1_3 := dataGetFriendList1_8_1.Result[0]
	assert.Equal(types.StatusAlive, friend1_8_1_3.Status)
	assert.Equal(me3_1.ID, friend1_8_1_3.FriendID)

	friend1_8_1_2 := dataGetFriendList1_8_1.Result[1]
	assert.Equal(types.StatusAlive, friend1_8_1_2.Status)
	assert.Equal(me2_1.ID, friend1_8_1_2.FriendID)

	friend1_8_1_0 := dataGetFriendList1_8_1.Result[2]
	assert.Equal(types.StatusAlive, friend1_8_1_0.Status)
	assert.Equal(me0_1.ID, friend1_8_1_0.FriendID)
	assert.Equal(friend0_8.ID, friend1_8_0.ID)

	// 9. get-raw-friend
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend0_9 := &friend.Friend{}
	testCore(t0, bodyString, friend0_9, t, isDebug)
	assert.Equal(friend0_8.ID, friend0_9.ID)
	assert.Equal(me1_1.ID, friend0_9.FriendID)

	friend1_9 := &friend.Friend{}
	testCore(t1, bodyString, friend1_9, t, isDebug)
	assert.Equal(friend1_8_0.ID, friend1_9.ID)
	assert.Equal(friend0_9.Friend0ID, friend1_9.Friend0ID)
	assert.Equal(friend0_9.Friend1ID, friend1_9.Friend1ID)
	assert.Equal(me0_1.ID, friend1_9.FriendID)

	// 10. master-oplog
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMasterOplogList0_10 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogList0_10, t, isDebug)
	assert.Equal(2, len(dataMasterOplogList0_10.Result))
	masterOplog0_10_0 := dataMasterOplogList0_10.Result[0]
	masterOplog0_10_1 := dataMasterOplogList0_10.Result[1]
	assert.Equal(types.StatusAlive, masterOplog0_10_0.ToStatus())
	assert.Equal(types.StatusAlive, masterOplog0_10_1.ToStatus())
	assert.Equal(masterOplog0_10_0.ObjID, me1_1.ID)
	assert.Equal(masterOplog0_10_1.ObjID, me0_1.ID)

	dataMasterOplogList1_10 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogList1_10, t, isDebug)
	assert.Equal(2, len(dataMasterOplogList1_10.Result))
	assert.Equal(dataMasterOplogList0_10, dataMasterOplogList1_10)
	masterOplog1_10_0 := dataMasterOplogList1_10.Result[0]
	masterOplog1_10_1 := dataMasterOplogList1_10.Result[1]
	assert.Equal(types.StatusAlive, masterOplog1_10_0.ToStatus())
	assert.Equal(types.StatusAlive, masterOplog1_10_1.ToStatus())
	assert.Equal(masterOplog1_10_0.ID, masterOplog1_10_1.MasterLogID)
	assert.Equal(1, len(masterOplog1_10_0.MasterSigns))
	masterSign1_10_0_0 := masterOplog1_10_0.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_10_0_0.ID)
	masterSign1_10_1_0 := masterOplog1_10_1.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_10_1_0.ID)

	// 11. masters
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMasterListFromCache", "params": ["%v"]}`, string(marshaled))

	dataMasterList0_11 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11, t, isDebug)
	assert.Equal(2, len(dataMasterList0_11.Result))

	dataMasterList1_11 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterList1_11, t, isDebug)
	assert.Equal(2, len(dataMasterList1_11.Result))

	// 11.1
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMasterList0_11_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11_1, t, isDebug)
	assert.Equal(2, len(dataMasterList0_11_1.Result))

	dataMasterList1_11_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterList1_11_1, t, isDebug)
	assert.Equal(2, len(dataMasterList1_11_1.Result))
	assert.Equal(dataMasterList0_11_1, dataMasterList1_11_1)

	// 12. member-oplog
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMemberOplogList0_12 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogList0_12, t, isDebug)
	assert.Equal(2, len(dataMemberOplogList0_12.Result))
	memberOplog0_12_0 := dataMemberOplogList0_12.Result[0]
	memberOplog0_12_1 := dataMemberOplogList0_12.Result[1]
	assert.Equal(types.StatusAlive, memberOplog0_12_0.ToStatus())
	assert.Equal(types.StatusAlive, memberOplog0_12_1.ToStatus())
	assert.Equal(memberOplog0_12_0.ObjID, me1_1.ID)
	assert.Equal(memberOplog0_12_1.ObjID, me0_1.ID)

	dataMemberOplogList1_12 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberOplogList1_12, t, isDebug)
	assert.Equal(2, len(dataMemberOplogList1_12.Result))
	assert.Equal(dataMemberOplogList0_12, dataMemberOplogList1_12)
	memberOplog1_12_0 := dataMemberOplogList1_12.Result[0]
	memberOplog1_12_1 := dataMemberOplogList1_12.Result[1]
	assert.Equal(types.StatusAlive, memberOplog1_12_0.ToStatus())
	assert.Equal(types.StatusAlive, memberOplog1_12_1.ToStatus())
	assert.Equal(masterOplog0_10_0.ID, memberOplog1_12_0.MasterLogID)
	assert.Equal(masterOplog0_10_0.ID, memberOplog1_12_1.MasterLogID)
	assert.Equal(1, len(memberOplog1_12_0.MasterSigns))
	masterSign1_12_0_0 := memberOplog1_12_0.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_12_0_0.ID)
	masterSign1_12_1_0 := memberOplog1_12_1.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_12_1_0.ID)

	// 12.1
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMemberList0_12_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberList0_12_1, t, isDebug)
	assert.Equal(2, len(dataMemberList0_12_1.Result))

	dataMemberList1_12_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberList1_12_1, t, isDebug)
	assert.Equal(2, len(dataMemberList1_12_1.Result))
	assert.Equal(dataMemberList0_12_1, dataMemberList1_12_1)

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
	marshaled, _ = friend2_8.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_createMessage", "params": ["%v", %v, []]}`, string(marshaled), string(msg))
	dataCreateMessage2_35 := &friend.BackendCreateMessage{}
	testCore(t2, bodyString, dataCreateMessage2_35, t, isDebug)
	assert.Equal(friend2_8.ID, dataCreateMessage2_35.FriendID)
	assert.Equal(2, dataCreateMessage2_35.NBlock)

	time.Sleep(10 * time.Second)

	// 36. friend-get-message-list
	marshaled, _ = friend2_8.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMessageList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetMessageList2_36 := &struct {
		Result []*friend.BackendGetMessage `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetMessageList2_36, t, isDebug)
	assert.Equal(1, len(dataGetMessageList2_36.Result))

	dataGetMessageList1_36 := &struct {
		Result []*friend.BackendGetMessage `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMessageList1_36, t, isDebug)
	assert.Equal(1, len(dataGetMessageList1_36.Result))

	// 37. get-friend-list-by-msg-create-ts
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendListByMsgCreateTS", "params": ["", 0, 1]}`)

	dataGetFriendList1_37 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_37, t, isDebug)
	assert.Equal(3, len(dataGetFriendList1_37.Result))

	friend1_37_2 := dataGetFriendList1_37.Result[0]
	assert.Equal(types.StatusAlive, friend1_37_2.Status)
	assert.Equal(me2_1.ID, friend1_37_2.FriendID)

	friend1_37_3 := dataGetFriendList1_37.Result[1]
	assert.Equal(types.StatusAlive, friend1_37_3.Status)
	assert.Equal(me3_1.ID, friend1_37_3.FriendID)

	friend1_37_0 := dataGetFriendList1_37.Result[2]
	assert.Equal(types.StatusAlive, friend1_37_0.Status)
	assert.Equal(me0_1.ID, friend1_37_0.FriendID)
	assert.Equal(friend0_8.ID, friend1_8_0.ID)

	// 38. get-friend-list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList1_38 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_38, t, isDebug)
	assert.Equal(3, len(dataGetFriendList1_38.Result))

	friend1_38_0 := dataGetFriendList1_38.Result[0]
	assert.Equal(types.StatusAlive, friend1_38_0.Status)
	assert.Equal(me0_1.ID, friend1_38_0.FriendID)
	assert.Equal(friend0_8.ID, friend1_8_0.ID)

	friend1_38_2 := dataGetFriendList1_38.Result[1]
	assert.Equal(types.StatusAlive, friend1_38_2.Status)
	assert.Equal(me2_1.ID, friend1_38_2.FriendID)

	friend1_38_3 := dataGetFriendList1_38.Result[2]
	assert.Equal(types.StatusAlive, friend1_38_3.Status)
	assert.Equal(me3_1.ID, friend1_38_3.FriendID)

	// 39. friend-create-message
	marshaled, _ = friend3_8.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_createMessage", "params": ["%v", %v, []]}`, string(marshaled), string(msg))
	dataCreateMessage3_39 := &friend.BackendCreateMessage{}
	testCore(t3, bodyString, dataCreateMessage3_39, t, isDebug)
	assert.Equal(friend3_8.ID, dataCreateMessage3_39.FriendID)
	assert.Equal(2, dataCreateMessage3_39.NBlock)

	time.Sleep(10 * time.Second)

	// 40. get-friend-list-by-msg-create-ts
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendListByMsgCreateTS", "params": ["", 0, 1]}`)

	dataGetFriendList1_40 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_40, t, isDebug)
	assert.Equal(3, len(dataGetFriendList1_40.Result))

	friend1_40_3 := dataGetFriendList1_40.Result[0]
	assert.Equal(types.StatusAlive, friend1_40_3.Status)
	assert.Equal(me3_1.ID, friend1_40_3.FriendID)

	friend1_40_2 := dataGetFriendList1_40.Result[1]
	assert.Equal(types.StatusAlive, friend1_40_2.Status)
	assert.Equal(me2_1.ID, friend1_40_2.FriendID)

	friend1_40_0 := dataGetFriendList1_40.Result[2]
	assert.Equal(types.StatusAlive, friend1_40_0.Status)
	assert.Equal(me0_1.ID, friend1_40_0.FriendID)
	assert.Equal(friend0_8.ID, friend1_8_0.ID)

	// 41. get-friend-list
	t.Logf("41. get-friend-list")
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList1_41 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_41, t, isDebug)
	assert.Equal(3, len(dataGetFriendList1_41.Result))

	friend1_41_0 := dataGetFriendList1_41.Result[0]
	assert.Equal(types.StatusAlive, friend1_41_0.Status)
	assert.Equal(me0_1.ID, friend1_41_0.FriendID)
	assert.Equal(friend0_8.ID, friend1_8_0.ID)

	friend1_41_2 := dataGetFriendList1_41.Result[1]
	assert.Equal(types.StatusAlive, friend1_41_2.Status)
	assert.Equal(me2_1.ID, friend1_41_2.FriendID)

	friend1_41_3 := dataGetFriendList1_41.Result[2]
	assert.Equal(types.StatusAlive, friend1_41_3.Status)
	assert.Equal(me3_1.ID, friend1_41_3.FriendID)

}
