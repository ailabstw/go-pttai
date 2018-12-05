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
	"io/ioutil"
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

func TestFriendImage(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledStr string
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
	//nodeID1_1 := me1_1.NodeID
	//pubKey1_1, _ := nodeID1_1.Pubkey()
	// nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

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

	// 5. show-url
	bodyString = `{"id": "testID", "method": "me_showURL", "params": []}`

	dataShowURL1_5 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowURL1_5, t, isDebug)
	url1_5 := dataShowURL1_5.URL

	// 7. join-friend
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_5)

	dataJoinFriend0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinFriend0_7, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend0_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(TimeSleepRestart)

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
	assert.Equal(1, len(dataGetFriendList1_8.Result))
	friend1_8 := dataGetFriendList1_8.Result[0]
	assert.Equal(types.StatusAlive, friend1_8.Status)
	assert.Equal(me0_1.ID, friend1_8.FriendID)
	assert.Equal(friend0_8.ID, friend1_8.ID)

	// 9. get-raw-friend
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend0_9 := &friend.Friend{}
	testCore(t0, bodyString, friend0_9, t, isDebug)
	assert.Equal(friend0_8.ID, friend0_9.ID)
	assert.Equal(me1_1.ID, friend0_9.FriendID)

	friend1_9 := &friend.Friend{}
	testCore(t1, bodyString, friend1_9, t, isDebug)
	assert.Equal(friend1_8.ID, friend1_9.ID)
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

	// 13. upload image
	marshaledID, _ = me0_3.BoardID.MarshalText()
	img0_13, _ := ioutil.ReadFile("./btn_confirm.png")
	marshaledStr = base64.StdEncoding.EncodeToString(img0_13)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_uploadImage", "params": ["%v", "image/png", "%v"]}`, string(marshaledID), marshaledStr)

	dataUploadImg0_13 := &content.BackendUploadImg{}
	testCore(t0, bodyString, dataUploadImg0_13, t, isDebug)
	assert.Equal(pkgservice.MediaTypeJPEG, dataUploadImg0_13.Type)

	// time.wait
	time.Sleep(TimeSleepRestart)

	// 14.0
	t.Logf("14.0. getBoardOplogList")
	marshaledID, _ = me0_3.BoardID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataBoardOplogList0_14_0 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataBoardOplogList0_14_0, t, isDebug)
	assert.Equal(2, len(dataBoardOplogList0_14_0.Result))

	dataBoardOplogList1_14_0 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataBoardOplogList1_14_0, t, isDebug)
	assert.Equal(2, len(dataBoardOplogList1_14_0.Result))

	// 14. get image
	marshaledID, _ = me0_3.BoardID.MarshalText()
	marshaledID2, _ = dataUploadImg0_13.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getImage", "params": ["%v", "%v"]}`, string(marshaledID), string(marshaledID2))

	dataGetImage0_14 := &content.BackendGetImg{}
	testCore(t0, bodyString, dataGetImage0_14, t, isDebug)

	outImg, _ := base64.StdEncoding.DecodeString(`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`)

	assert.Equal(outImg, dataGetImage0_14.Buf)
	assert.Equal(pkgservice.MediaTypeJPEG, dataGetImage0_14.Type)
	assert.Equal(dataUploadImg0_13.ID, dataGetImage0_14.ID)

	// t1
	dataGetImage1_14 := &content.BackendGetImg{}
	testCore(t1, bodyString, dataGetImage1_14, t, isDebug)

	assert.Equal(outImg, dataGetImage1_14.Buf)
	assert.Equal(pkgservice.MediaTypeJPEG, dataGetImage1_14.Type)
	assert.Equal(dataUploadImg0_13.ID, dataGetImage1_14.ID)
}
