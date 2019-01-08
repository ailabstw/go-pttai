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
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestFriendJoinBoard3Basic(t *testing.T) {
	NNodes = 3
	isDebug := true

	var err error
	var bodyString string
	var marshaledStr string
	var marshaled []byte
	//var marshaledID []byte
	//var dummyBool bool
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")
	t2 := baloo.New("http://127.0.0.1:9452")

	// 0 test-error
	err = testError("http://127.0.0.1:9450")
	assert.Equal(nil, err)

	err = testError("http://127.0.0.1:9451")
	assert.Equal(nil, err)

	err = testError("http://127.0.0.1:9452")
	assert.Equal(nil, err)

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}
	testCore(t0, bodyString, me0_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_1.Status)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)

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
	//profileID0_3 := me0_3.ProfileID

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))
	//profileID1_3 := me1_3.ProfileID

	me2_3 := &me.MyInfo{}
	testCore(t2, bodyString, me2_3, t, isDebug)
	assert.Equal(types.StatusAlive, me2_3.Status)
	assert.Equal(me2_1.ID, me2_3.ID)
	assert.Equal(1, len(me2_3.OwnerIDs))
	assert.Equal(me2_3.ID, me2_3.OwnerIDs[0])
	assert.Equal(true, me2_3.IsOwner(me2_3.ID))
	//profileID2_3 := me2_3.ProfileID

	// 4. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_4 string

	testCore(t0, bodyString, &myKey0_4, t, isDebug)
	if isDebug {
		t.Logf("myKey0_4: %v\n", myKey0_4)
	}

	var myKey2_4 string

	testCore(t2, bodyString, &myKey2_4, t, isDebug)
	if isDebug {
		t.Logf("myKey2_4: %v\n", myKey2_4)
	}

	// 5. create-board
	title := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createBoard", "params": ["%v", true]}`, marshaledStr)

	dataCreateBoard0_5 := &content.BackendCreateBoard{}

	testCore(t0, bodyString, dataCreateBoard0_5, t, isDebug)
	assert.Equal(pkgservice.EntityTypePrivate, dataCreateBoard0_5.BoardType)
	assert.Equal(title, dataCreateBoard0_5.Title)
	assert.Equal(types.StatusAlive, dataCreateBoard0_5.Status)
	assert.Equal(me0_1.ID, dataCreateBoard0_5.CreatorID)
	assert.Equal(me0_1.ID, dataCreateBoard0_5.UpdaterID)

	// 5. show-board-url
	marshaled, _ = dataCreateBoard0_5.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_showBoardURL", "params": ["%v"]}`, string(marshaled))

	dataShowBoardURL0_5 := &pkgservice.BackendJoinURL{}
	testCore(t0, bodyString, dataShowBoardURL0_5, t, isDebug)
	boardUrl0_5 := dataShowBoardURL0_5.URL

	// 5. show-url
	bodyString = `{"id": "testID", "method": "me_showURL", "params": []}`

	dataShowURL0_5 := &pkgservice.BackendJoinURL{}
	testCore(t0, bodyString, dataShowURL0_5, t, isDebug)
	url0_5 := dataShowURL0_5.URL

	// 7. join-friend
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url0_5)

	dataJoinFriend1_7 := &pkgservice.BackendJoinRequest{}
	testCore(t1, bodyString, dataJoinFriend1_7, t, isDebug)

	assert.Equal(me0_3.ID, dataJoinFriend1_7.CreatorID)
	assert.Equal(me0_1.NodeID, dataJoinFriend1_7.NodeID)

	dataJoinFriend2_7 := &pkgservice.BackendJoinRequest{}
	testCore(t2, bodyString, dataJoinFriend2_7, t, isDebug)

	assert.Equal(me0_3.ID, dataJoinFriend2_7.CreatorID)
	assert.Equal(me0_1.NodeID, dataJoinFriend2_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(30 * time.Second)

	// 8. get-friend-list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList0_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetFriendList0_8, t, isDebug)
	assert.Equal(2, len(dataGetFriendList0_8.Result))
	friend0_8_1 := dataGetFriendList0_8.Result[0]
	assert.Equal(types.StatusAlive, friend0_8_1.Status)
	assert.Equal(me2_1.ID, friend0_8_1.FriendID)
	friend0_8_2 := dataGetFriendList0_8.Result[1]
	assert.Equal(types.StatusAlive, friend0_8_2.Status)
	assert.Equal(me1_1.ID, friend0_8_2.FriendID)

	dataGetFriendList1_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList1_8.Result))
	friend1_8 := dataGetFriendList1_8.Result[0]
	assert.Equal(types.StatusAlive, friend1_8.Status)
	assert.Equal(me0_1.ID, friend1_8.FriendID)
	assert.Equal(friend0_8_2.ID, friend1_8.ID)

	dataGetFriendList2_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetFriendList2_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList2_8.Result))
	friend2_8 := dataGetFriendList2_8.Result[0]
	assert.Equal(types.StatusAlive, friend2_8.Status)
	assert.Equal(me0_1.ID, friend2_8.FriendID)
	assert.Equal(friend0_8_1.ID, friend2_8.ID)

	// 9. get-raw-friend
	marshaled, _ = friend0_8_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend0_9_1 := &friend.Friend{}
	testCore(t0, bodyString, friend0_9_1, t, isDebug)
	assert.Equal(friend0_8_1.ID, friend0_9_1.ID)

	marshaled, _ = friend0_8_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend0_9_2 := &friend.Friend{}
	testCore(t0, bodyString, friend0_9_2, t, isDebug)
	assert.Equal(friend0_8_2.ID, friend0_9_2.ID)

	marshaled, _ = friend0_8_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend1_9 := &friend.Friend{}
	testCore(t1, bodyString, friend1_9, t, isDebug)
	assert.Equal(friend0_8_2.ID, friend1_9.ID)
	assert.Equal(me0_1.ID, friend1_9.FriendID)

	marshaled, _ = friend0_8_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend2_9 := &friend.Friend{}
	testCore(t2, bodyString, friend2_9, t, isDebug)
	assert.Equal(friend0_8_1.ID, friend2_9.ID)
	assert.Equal(me0_1.ID, friend2_9.FriendID)

	// 15. join-board
	t.Logf("15. join-board")
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinBoard", "params": ["%v"]}`, boardUrl0_5)

	dataJoinBoard1_15 := &pkgservice.BackendJoinRequest{}
	rbody, _ := testCore(t1, bodyString, dataJoinBoard1_15, t, isDebug)
	t.Logf("15. join-board: rbody: %v", rbody)

	dataJoinBoard2_15 := &pkgservice.BackendJoinRequest{}
	rbody, _ = testCore(t2, bodyString, dataJoinBoard2_15, t, isDebug)
	t.Logf("15. join-board: rbody: %v", rbody)

	// wait 10 secs
	time.Sleep(30 * time.Second)

	// 16. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_16 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_16, t, isDebug)
	assert.Equal(4, len(dataBoardList0_16.Result))

	dataBoardList1_16 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t1, bodyString, dataBoardList1_16, t, isDebug)
	assert.Equal(3, len(dataBoardList1_16.Result))
	dataBoard1_16_0 := dataBoardList1_16.Result[0]
	dataBoard1_16_1 := dataBoardList1_16.Result[1]
	dataBoard1_16_2 := dataBoardList1_16.Result[2]

	assert.Equal(types.StatusAlive, dataBoard1_16_0.Status)
	assert.Equal(types.StatusAlive, dataBoard1_16_1.Status)
	assert.Equal(types.StatusAlive, dataBoard1_16_2.Status)

	dataBoardList2_16 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t2, bodyString, dataBoardList2_16, t, isDebug)
	assert.Equal(3, len(dataBoardList2_16.Result))
}
