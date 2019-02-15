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
	"reflect"
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

func TestFriendDeleteMember4(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
	var marshaled2 []byte
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
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

	// 13. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_13 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_13, t, isDebug)
	assert.Equal(2, len(dataBoardList0_13.Result))
	board0_13_0 := dataBoardList0_13.Result[0]
	assert.Equal(me0_3.BoardID, board0_13_0.ID)
	assert.Equal(types.StatusAlive, board0_13_0.Status)

	board0_13_1 := dataBoardList0_13.Result[1]
	assert.Equal(me1_3.BoardID, board0_13_1.ID)
	assert.Equal(types.StatusAlive, board0_13_1.Status)

	// t1
	dataBoardList1_13 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t1, bodyString, dataBoardList1_13, t, isDebug)
	assert.Equal(2, len(dataBoardList1_13.Result))
	board1_13_0 := dataBoardList1_13.Result[0]
	assert.Equal(me1_3.BoardID, board1_13_0.ID)
	assert.Equal(types.StatusAlive, board1_13_0.Status)

	board1_13_1 := dataBoardList1_13.Result[1]
	assert.Equal(me0_3.BoardID, board1_13_1.ID)
	assert.Equal(types.StatusAlive, board1_13_1.Status)

	// 13.1. create-article
	article, _ := json.Marshal([]string{
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

	marshaledID, _ = board1_13_1.ID.MarshalText()

	title0_13_1 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_13_1)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article))
	dataCreateArticle0_13_1 := &content.BackendCreateArticle{}
	testCore(t0, bodyString, dataCreateArticle0_13_1, t, isDebug)
	assert.Equal(board1_13_1.ID, dataCreateArticle0_13_1.BoardID)
	assert.Equal(3, dataCreateArticle0_13_1.NBlock)

	// sleep
	time.Sleep(TimeSleepDefault)

	// 13.2. content-get-article-list
	marshaledID, _ = board1_13_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_13_2 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_13_2, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_13_2.Result))
	article0_13_2 := dataGetArticleList0_13_2.Result[0]

	dataGetArticleList1_13_2 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_13_2, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_13_2.Result))

	// 13.3. get-article-block
	marshaledID2, _ = article0_13_2.ID.MarshalText()
	marshaledID3, _ = article0_13_2.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_13_3 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_13_3, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList0_13_3.Result))

	article0 := [][]byte{
		[]byte("測試1"),
	}

	article1 := [][]byte{
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
		[]byte("測試21"),
	}

	article2 := [][]byte{
		[]byte("測試22"),
	}

	assert.Equal(article0, dataGetArticleBlockList0_13_3.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList0_13_3.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList0_13_3.Result[2].Buf)

	dataGetArticleBlockList1_13_3 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_13_3, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList1_13_3.Result))

	assert.Equal(article0, dataGetArticleBlockList1_13_3.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList1_13_3.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList1_13_3.Result[2].Buf)

	// 14. delete-member
	t.Logf("14. delete-member")
	marshaled, _ = board1_13_1.ID.MarshalText()
	marshaled2, _ = me1_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_deleteMember", "params": ["%v", "%v"]}`, string(marshaled), string(marshaled2))

	dataDeleteMember0_14 := false
	testCore(t0, bodyString, &dataDeleteMember0_14, t, isDebug)
	assert.Equal(true, dataDeleteMember0_14)

	time.Sleep(TimeSleepDefault)

	// 14.1. get peers
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getPeers", "params": ["%v"]}`, string(marshaled))

	dataPeers0_14_1 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_14_1, t, isDebug)
	assert.Equal(1, len(dataPeers0_14_1.Result))

	dataPeers1_14_1 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPeers1_14_1, t, isDebug)
	assert.Equal(0, len(dataPeers1_14_1.Result))

	// 15. get board list
	t.Logf("15. get board list")
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_15 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_15, t, isDebug)
	assert.Equal(2, len(dataBoardList0_15.Result))
	board0_15_0 := dataBoardList0_15.Result[0]
	assert.Equal(me0_3.BoardID, board0_15_0.ID)
	assert.Equal(types.StatusAlive, board0_15_0.Status)

	board0_15_1 := dataBoardList0_15.Result[1]
	assert.Equal(me1_3.BoardID, board0_15_1.ID)
	assert.Equal(types.StatusAlive, board0_15_1.Status)

	// t1
	dataBoardList1_15 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t1, bodyString, dataBoardList1_15, t, isDebug)
	assert.Equal(2, len(dataBoardList1_15.Result))
	board1_15_0 := dataBoardList1_15.Result[0]
	assert.Equal(me1_3.BoardID, board1_15_0.ID)
	assert.Equal(types.StatusAlive, board1_15_0.Status)

	board1_15_1 := dataBoardList1_15.Result[1]
	assert.Equal(me0_3.BoardID, board1_15_1.ID)
	assert.Equal(types.StatusDeleted, board1_15_1.Status)

	// 15.2. content-get-article-list
	marshaledID, _ = board1_13_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_15_2 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_15_2, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_15_2.Result))

	dataGetArticleList1_15_2 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_15_2, t, isDebug)
	assert.Equal(0, len(dataGetArticleList1_15_2.Result))

	// 15.3. get-article-block
	marshaledID2, _ = article0_13_2.ID.MarshalText()
	marshaledID3, _ = article0_13_2.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_15_3 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_15_3, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList0_15_3.Result))

	assert.Equal(article0, dataGetArticleBlockList0_15_3.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList0_15_3.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList0_15_3.Result[2].Buf)

	dataGetArticleBlockList1_15_3 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_15_3, t, isDebug)
	assert.Equal(0, len(dataGetArticleBlockList1_15_3.Result))

	// 16.0. sync.
	t.Logf("16.0. force sync")
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_forceSync", "params": ["%v"]}`, string(marshaled))

	bool0_16_0 := false
	testCore(t0, bodyString, &bool0_16_0, t, isDebug)
	assert.Equal(true, bool0_16_0)

	time.Sleep(TimeSleepDefault)

	bool1_16_0 := false
	testCore(t1, bodyString, &bool1_16_0, t, isDebug)
	assert.Equal(true, bool1_16_0)

	time.Sleep(TimeSleepDefault)

	// 16. get board-oplog
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataBoardOplogList0_16 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataBoardOplogList0_16, t, isDebug)
	assert.Equal(2, len(dataBoardOplogList0_16.Result))

	dataBoardOplogList1_16 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataBoardOplogList1_16, t, isDebug)
	assert.Equal(1, len(dataBoardOplogList1_16.Result))

	// 17. get member-oplog
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMemberOplogList0_17 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogList0_17, t, isDebug)
	assert.Equal(3, len(dataMemberOplogList0_17.Result))

	dataMemberOplogList1_17 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberOplogList1_17, t, isDebug)
	assert.Equal(1, len(dataMemberOplogList1_17.Result))

	// 18. get master-oplog
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMasterOplogList0_18 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogList0_18, t, isDebug)
	assert.Equal(1, len(dataMasterOplogList0_18.Result))

	dataMasterOplogList1_18 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogList1_18, t, isDebug)
	assert.Equal(0, len(dataMasterOplogList1_18.Result))

	// 19. get members
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMembers0_19 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMembers0_19, t, isDebug)
	assert.Equal(2, len(dataMembers0_19.Result))
	member0_19_0 := dataMembers0_19.Result[0]
	member0_19_1 := dataMembers0_19.Result[1]
	var member0_19_me *pkgservice.Member
	var member0_19_other *pkgservice.Member
	if reflect.DeepEqual(member0_19_0.ID, me0_1.ID) {
		member0_19_me = member0_19_0
		member0_19_other = member0_19_1
	} else {
		member0_19_me = member0_19_1
		member0_19_other = member0_19_0
	}

	assert.Equal(types.StatusAlive, member0_19_me.Status)
	assert.Equal(types.StatusDeleted, member0_19_other.Status)

	dataMembers1_19 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMembers1_19, t, isDebug)
	assert.Equal(1, len(dataMembers1_19.Result))

	member1_19_0 := dataMembers1_19.Result[0]
	assert.Equal(member1_19_0.ID, me1_1.ID)
	assert.Equal(types.StatusDeleted, member1_19_0.Status)

	// 20. get peers
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getPeers", "params": ["%v"]}`, string(marshaled))

	dataPeers0_19 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_19, t, isDebug)
	assert.Equal(1, len(dataPeers0_19.Result))

	dataPeers1_19 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPeers1_19, t, isDebug)
	assert.Equal(0, len(dataPeers1_19.Result))

	// 21. content show url
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_showBoardURL", "params": ["%v"]}`, string(marshaled))

	dataShowURL0_21 := &pkgservice.BackendJoinURL{}
	testCore(t0, bodyString, dataShowURL0_21, t, isDebug)
	url0_21 := dataShowURL0_21.URL

	// 22. join-board
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinBoard", "params": ["%v"]}`, url0_21)

	dataJoinBoard1_22 := &pkgservice.BackendJoinRequest{}
	testCore(t1, bodyString, dataJoinBoard1_22, t, isDebug)

	assert.Equal(me0_1.ID, dataJoinBoard1_22.CreatorID)

	// sleep
	time.Sleep(TimeSleepRestart)

	// 23. get board list
	t.Logf("23. get board list")
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_23 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_23, t, isDebug)
	assert.Equal(2, len(dataBoardList0_23.Result))
	board0_23_0 := dataBoardList0_23.Result[0]
	assert.Equal(me0_3.BoardID, board0_23_0.ID)
	assert.Equal(types.StatusAlive, board0_23_0.Status)

	board0_23_1 := dataBoardList0_23.Result[1]
	assert.Equal(me1_3.BoardID, board0_23_1.ID)
	assert.Equal(types.StatusAlive, board0_23_1.Status)

	// t1
	dataBoardList1_23 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t1, bodyString, dataBoardList1_23, t, isDebug)
	assert.Equal(2, len(dataBoardList1_23.Result))
	board1_23_0 := dataBoardList1_23.Result[0]
	assert.Equal(me1_3.BoardID, board1_23_0.ID)
	assert.Equal(types.StatusAlive, board1_23_0.Status)

	board1_23_1 := dataBoardList1_23.Result[1]
	assert.Equal(me0_3.BoardID, board1_23_1.ID)
	assert.Equal(types.StatusAlive, board1_23_1.Status)

	// 24. get board-oplog
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataBoardOplogList0_24 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataBoardOplogList0_24, t, isDebug)
	assert.Equal(2, len(dataBoardOplogList0_24.Result))

	dataBoardOplogList1_24 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataBoardOplogList1_24, t, isDebug)
	assert.Equal(2, len(dataBoardOplogList1_24.Result))

	// 25. get member-oplog
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMemberOplogList0_25 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogList0_25, t, isDebug)
	assert.Equal(4, len(dataMemberOplogList0_25.Result))

	dataMemberOplogList1_25 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberOplogList1_25, t, isDebug)
	assert.Equal(4, len(dataMemberOplogList1_25.Result))

	// 26. get master-oplog
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMasterOplogList0_26 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogList0_26, t, isDebug)
	assert.Equal(1, len(dataMasterOplogList0_26.Result))

	dataMasterOplogList1_26 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogList1_26, t, isDebug)
	assert.Equal(1, len(dataMasterOplogList1_26.Result))

	// 27. get peers
	marshaled, _ = board1_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getPeers", "params": ["%v"]}`, string(marshaled))

	dataPeers0_27 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPeers0_27, t, isDebug)
	assert.Equal(1, len(dataPeers0_27.Result))

	dataPeers1_27 := &struct {
		Result []*pkgservice.BackendPeer `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPeers1_27, t, isDebug)
	assert.Equal(1, len(dataPeers1_27.Result))

	// 28.2 content-get-article-list
	marshaledID, _ = board1_13_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_28_2 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_28_2, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_28_2.Result))

	dataGetArticleList1_28_2 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_28_2, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_28_2.Result))

	// 28.3. get-article-block
	marshaledID2, _ = article0_13_2.ID.MarshalText()
	marshaledID3, _ = article0_13_2.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_28_3 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_28_3, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList0_28_3.Result))

	assert.Equal(article0, dataGetArticleBlockList0_28_3.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList0_28_3.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList0_28_3.Result[2].Buf)

	dataGetArticleBlockList1_28_3 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_28_3, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList1_28_3.Result))

	assert.Equal(article0, dataGetArticleBlockList1_28_3.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList1_28_3.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList1_28_3.Result[2].Buf)
}
