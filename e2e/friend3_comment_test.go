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
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestFriend3Comment(t *testing.T) {
	NNodes = 3
	isDebug := true

	var bodyString string
	var marshaled []byte
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
	var marshaledStr string
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

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)
	//nodeID1_1 := me1_1.NodeID
	//pubKey1_1, _ := nodeID1_1.Pubkey()
	// nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

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
	testCore(t1, bodyString, me2_3, t, isDebug)
	assert.Equal(types.StatusAlive, me2_3.Status)
	assert.Equal(me1_1.ID, me2_3.ID)
	assert.Equal(1, len(me2_3.OwnerIDs))
	assert.Equal(me2_3.ID, me2_3.OwnerIDs[0])
	assert.Equal(true, me2_3.IsOwner(me2_3.ID))

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

	dataJoinFriend2_7 := &pkgservice.BackendJoinRequest{}
	testCore(t2, bodyString, dataJoinFriend2_7, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend2_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend2_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(10 * time.Second)

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
	assert.Equal(2, len(dataGetFriendList1_8.Result))
	friend1_8 := dataGetFriendList1_8.Result[0]
	assert.Equal(types.StatusAlive, friend1_8.Status)
	assert.Equal(me0_1.ID, friend1_8.FriendID)
	assert.Equal(friend0_8.ID, friend1_8.ID)
	friend1_8_1 := dataGetFriendList1_8.Result[1]
	assert.Equal(types.StatusAlive, friend1_8_1.Status)
	assert.Equal(me2_1.ID, friend1_8_1.FriendID)

	dataGetFriendList2_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetFriendList2_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList2_8.Result))
	friend2_8 := dataGetFriendList2_8.Result[0]
	assert.Equal(types.StatusAlive, friend2_8.Status)
	assert.Equal(me1_1.ID, friend2_8.FriendID)
	assert.Equal(friend1_8_1.ID, friend2_8.ID)

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

	// 13. create-board
	title := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createBoard", "params": ["%v", false]}`, marshaledStr)

	dataCreateBoard1_13 := &content.BackendCreateBoard{}

	testCore(t1, bodyString, dataCreateBoard1_13, t, isDebug)
	assert.Equal(pkgservice.EntityTypePrivate, dataCreateBoard1_13.BoardType)
	assert.Equal(title, dataCreateBoard1_13.Title)
	assert.Equal(types.StatusAlive, dataCreateBoard1_13.Status)
	assert.Equal(me1_1.ID, dataCreateBoard1_13.CreatorID)
	assert.Equal(me1_1.ID, dataCreateBoard1_13.UpdaterID)

	// 14. show-board-url
	marshaled, _ = dataCreateBoard1_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_showBoardURL", "params": ["%v"]}`, string(marshaled))

	dataShowBoardURL1_13 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowBoardURL1_13, t, isDebug)
	url1_13 := dataShowBoardURL1_13.URL

	// 15. join-board
	t.Logf("15. join-board")
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinBoard", "params": ["%v"]}`, url1_13)

	dataJoinBoard0_15 := &pkgservice.BackendJoinRequest{}
	rbody, _ := testCore(t0, bodyString, dataJoinBoard0_15, t, isDebug)
	t.Logf("15. join-board: rbody: %v", rbody)

	dataJoinBoard2_15 := &pkgservice.BackendJoinRequest{}
	rbody, _ = testCore(t2, bodyString, dataJoinBoard2_15, t, isDebug)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 16. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_16 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_16, t, isDebug)
	assert.Equal(3, len(dataBoardList0_16.Result))

	dataBoardList1_16 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t1, bodyString, dataBoardList1_16, t, isDebug)
	assert.Equal(4, len(dataBoardList1_16.Result))
	board1_16_0 := dataBoardList1_16.Result[0]
	board1_16_1 := dataBoardList1_16.Result[1]
	board1_16_2 := dataBoardList1_16.Result[2]
	board1_16_3 := dataBoardList1_16.Result[3]

	assert.Equal(types.StatusAlive, board1_16_0.Status)
	assert.Equal(pkgservice.EntityTypePersonal, board1_16_0.BoardType)
	assert.Equal(me1_1.ID, board1_16_0.CreatorID)

	assert.Equal(types.StatusAlive, board1_16_1.Status)
	assert.Equal(pkgservice.EntityTypePersonal, board1_16_1.BoardType)

	assert.Equal(types.StatusAlive, board1_16_2.Status)
	assert.Equal(pkgservice.EntityTypePersonal, board1_16_2.BoardType)

	assert.Equal(types.StatusAlive, board1_16_3.Status)
	assert.Equal(pkgservice.EntityTypePrivate, board1_16_3.BoardType)

	dataBoardList2_16 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t2, bodyString, dataBoardList2_16, t, isDebug)
	assert.Equal(3, len(dataBoardList2_16.Result))

	// 17. get count peers0
	marshaled, _ = board1_16_0.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_countPeers", "params": ["%v"]}`, string(marshaled))

	count0_17, _ := testIntCore(t0, bodyString, t, isDebug)
	assert.Equal(1, count0_17)

	count1_17, _ := testIntCore(t1, bodyString, t, isDebug)
	assert.Equal(2, count1_17)

	// 17.1. count peers1
	marshaled, _ = board1_16_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_countPeers", "params": ["%v"]}`, string(marshaled))

	count0_17_1, _ := testIntCore(t0, bodyString, t, isDebug)
	assert.Equal(1, count0_17_1)

	count1_17_1, _ := testIntCore(t1, bodyString, t, isDebug)
	assert.Equal(1, count1_17_1)

	// 17.2. count peers2
	marshaled, _ = board1_16_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_countPeers", "params": ["%v"]}`, string(marshaled))

	count0_17_2, _ := testIntCore(t0, bodyString, t, isDebug)
	assert.Equal(1, count0_17_2)

	count1_17_2, _ := testIntCore(t1, bodyString, t, isDebug)
	assert.Equal(2, count1_17_2)

	count2_17_2, _ := testIntCore(t0, bodyString, t, isDebug)
	assert.Equal(1, count2_17_2)

	// 35. create-article
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

	marshaledID, _ = board1_16_3.ID.MarshalText()

	title0_35 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_35)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article))
	dataCreateArticle0_35 := &content.BackendCreateArticle{}
	testCore(t0, bodyString, dataCreateArticle0_35, t, isDebug)
	assert.Equal(board1_16_3.ID, dataCreateArticle0_35.BoardID)
	assert.Equal(3, dataCreateArticle0_35.NBlock)

	// 36. content-get-article-list
	marshaledID, _ = board1_16_3.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_36 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_36, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_36.Result))
	article0_36 := dataGetArticleList0_36.Result[0]

	// 38. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_37 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_37, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList0_37.Result))

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

	assert.Equal(article0, dataGetArticleBlockList0_37.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList0_37.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList0_37.Result[2].Buf)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 39. content-get-article-list
	marshaledID, _ = board1_16_3.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_39 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_39, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_39.Result))

	dataGetArticleList1_39 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_39, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_39.Result))
	article1_39 := dataGetArticleList1_39.Result[0]

	dataGetArticleList2_39 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleList2_39, t, isDebug)
	assert.Equal(1, len(dataGetArticleList2_39.Result))

	// 40. ptt-oplog
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_getPttOplogList", "params": ["", 0, 2]}`)

	dataPttOplogList0_40 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPttOplogList0_40, t, isDebug)
	assert.Equal(1, len(dataPttOplogList0_40.Result))
	pttOplog0_40_0 := dataPttOplogList0_40.Result[0]
	assert.Equal(pkgservice.PttOpTypeCreateFriend, pttOplog0_40_0.Op)
	assert.Equal(me1_3.ID, pttOplog0_40_0.CreatorID)

	dataPttOplogList1_40 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPttOplogList1_40, t, isDebug)
	assert.Equal(3, len(dataPttOplogList1_40.Result))
	pttOplog1_40_0 := dataPttOplogList1_40.Result[0]
	assert.Equal(pkgservice.PttOpTypeCreateFriend, pttOplog1_40_0.Op)

	pttOplog1_40_1 := dataPttOplogList1_40.Result[1]
	assert.Equal(pkgservice.PttOpTypeCreateFriend, pttOplog1_40_1.Op)

	pttOplog1_40_2 := dataPttOplogList1_40.Result[2]
	assert.Equal(pkgservice.PttOpTypeCreateArticle, pttOplog1_40_2.Op)
	assert.Equal(article1_39.ID, pttOplog1_40_2.ObjID)

	dataPttOplogList2_40 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPttOplogList2_40, t, isDebug)
	assert.Equal(1, len(dataPttOplogList2_40.Result))
	pttOplog2_40_0 := dataPttOplogList2_40.Result[0]
	assert.Equal(pkgservice.PttOpTypeCreateFriend, pttOplog2_40_0.Op)
	assert.Equal(me1_3.ID, pttOplog2_40_0.CreatorID)

	// 41. content-create-comment
	comment := []byte("這是comment")
	commentStr := base64.StdEncoding.EncodeToString(comment)

	marshaledID, _ = board1_16_3.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("41. content_createComment: bodyString: %v", bodyString)
	dataCreateComment1_41 := &content.BackendCreateComment{}
	testCore(t1, bodyString, dataCreateComment1_41, t, isDebug)
	assert.Equal(article0_36.ID, dataCreateComment1_41.ArticleID)
	assert.Equal(article0_36.BoardID, dataCreateComment1_41.BoardID)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 42. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2))

	dataGetArticleBlockList0_42 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_42, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList0_42.Result))
	articleBlock0_42 := dataGetArticleBlockList0_42.Result[3]
	assert.Equal(types.StatusAlive, articleBlock0_42.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_42.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_42.CommentType)
	assert.Equal([][]byte{comment}, articleBlock0_42.Buf)

	dataGetArticleBlockList1_42 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_42, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList1_42.Result))
	articleBlock1_42 := dataGetArticleBlockList1_42.Result[3]
	assert.Equal(types.StatusAlive, articleBlock1_42.Status)
	assert.Equal(content.ContentTypeComment, articleBlock1_42.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock1_42.CommentType)
	assert.Equal([][]byte{comment}, articleBlock1_42.Buf)

	dataGetArticleBlockList2_42 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleBlockList2_42, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList2_42.Result))
	articleBlock2_42 := dataGetArticleBlockList2_42.Result[3]
	assert.Equal(types.StatusAlive, articleBlock2_42.Status)
	assert.Equal(content.ContentTypeComment, articleBlock2_42.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock2_42.CommentType)
	assert.Equal([][]byte{comment}, articleBlock2_42.Buf)

	// 43. ptt-oplog
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_getPttOplogList", "params": ["", 0, 2]}`)

	dataPttOplogList0_43 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPttOplogList0_43, t, isDebug)
	assert.Equal(2, len(dataPttOplogList0_43.Result))
	pttOplog0_43_1 := dataPttOplogList0_43.Result[1]
	assert.Equal(pkgservice.PttOpTypeCreateComment, pttOplog0_43_1.Op)
	assert.Equal(me1_3.ID, pttOplog0_43_1.CreatorID)

	dataPttOplogList1_43 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPttOplogList1_43, t, isDebug)
	assert.Equal(3, len(dataPttOplogList1_43.Result))
	pttOplog1_43_0 := dataPttOplogList1_43.Result[0]
	assert.Equal(pkgservice.PttOpTypeCreateFriend, pttOplog1_43_0.Op)

	pttOplog1_43_1 := dataPttOplogList1_43.Result[1]
	assert.Equal(pkgservice.PttOpTypeCreateFriend, pttOplog1_43_1.Op)

	pttOplog1_43_2 := dataPttOplogList1_43.Result[2]
	assert.Equal(pkgservice.PttOpTypeCreateArticle, pttOplog1_43_2.Op)
	assert.Equal(article1_39.ID, pttOplog1_43_2.ObjID)
	assert.Equal(me0_3.ID, pttOplog1_43_2.CreatorID)

	dataPttOplogList2_43 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t2, bodyString, dataPttOplogList2_43, t, isDebug)
	assert.Equal(1, len(dataPttOplogList2_43.Result))

	// 44. create-article
	article, _ = json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試31")),
		base64.StdEncoding.EncodeToString([]byte("測試32")),
		base64.StdEncoding.EncodeToString([]byte("測試33")),
		base64.StdEncoding.EncodeToString([]byte("測試34")),
		base64.StdEncoding.EncodeToString([]byte("測試35")),
		base64.StdEncoding.EncodeToString([]byte("測試36")),
		base64.StdEncoding.EncodeToString([]byte("測試37")),
		base64.StdEncoding.EncodeToString([]byte("測試38")),
		base64.StdEncoding.EncodeToString([]byte("測試39")),
		base64.StdEncoding.EncodeToString([]byte("測試40")),
		base64.StdEncoding.EncodeToString([]byte("測試41")),
		base64.StdEncoding.EncodeToString([]byte("測試42")),
		base64.StdEncoding.EncodeToString([]byte("測試43")),
		base64.StdEncoding.EncodeToString([]byte("測試44")),
		base64.StdEncoding.EncodeToString([]byte("測試45")),
		base64.StdEncoding.EncodeToString([]byte("測試46")),
		base64.StdEncoding.EncodeToString([]byte("測試47")),
		base64.StdEncoding.EncodeToString([]byte("測試48")),
		base64.StdEncoding.EncodeToString([]byte("測試49")),
		base64.StdEncoding.EncodeToString([]byte("測試50")),
		base64.StdEncoding.EncodeToString([]byte("測試51")),
		base64.StdEncoding.EncodeToString([]byte("測試52")),
	})

	marshaledID, _ = board1_16_3.ID.MarshalText()

	title0_44 := []byte("標題2")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_44)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article))
	dataCreateArticle1_44 := &content.BackendCreateArticle{}
	testCore(t1, bodyString, dataCreateArticle1_44, t, isDebug)
	assert.Equal(board1_16_3.ID, dataCreateArticle1_44.BoardID)
	assert.Equal(3, dataCreateArticle1_44.NBlock)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 45. content-get-article-list
	marshaledID, _ = board1_16_3.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_45 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_45, t, isDebug)
	assert.Equal(2, len(dataGetArticleList0_45.Result))
	article0_45_1 := dataGetArticleList0_45.Result[1]

	dataGetArticleList1_45 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_45, t, isDebug)
	assert.Equal(2, len(dataGetArticleList1_45.Result))

	// 46. ptt-oplog
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_getPttOplogList", "params": ["", 0, 2]}`)

	dataPttOplogList0_46 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataPttOplogList0_46, t, isDebug)
	assert.Equal(3, len(dataPttOplogList0_46.Result))
	pttOplog0_46_2 := dataPttOplogList0_46.Result[2]
	opData0_46_2 := &pkgservice.PttOpCreateArticle{}
	pttOplog0_46_2.GetData(opData0_46_2)
	assert.Equal(pkgservice.PttOpTypeCreateArticle, pttOplog0_46_2.Op)
	assert.Equal(me1_3.ID, pttOplog0_46_2.CreatorID)
	assert.Equal(article0_45_1.ID, pttOplog0_46_2.ObjID)
	assert.Equal(board1_16_3.ID, opData0_46_2.BoardID)
	assert.Equal(title0_44, opData0_46_2.Title)

	dataPttOplogList1_46 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataPttOplogList1_46, t, isDebug)
	assert.Equal(3, len(dataPttOplogList1_46.Result))

	dataPttOplogList2_46 := &struct {
		Result []*pkgservice.PttOplog `json:"result"`
	}{}
	testListCore(t2, bodyString, dataPttOplogList2_46, t, isDebug)
	assert.Equal(2, len(dataPttOplogList2_46.Result))
	pttOplog2_46_1 := dataPttOplogList2_46.Result[1]
	assert.Equal(me1_3.ID, pttOplog2_46_1.CreatorID)
	assert.Equal(pkgservice.PttOpTypeCreateArticle, pttOplog2_46_1.Op)
}
