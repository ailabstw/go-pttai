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
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDevice3ForceSyncArticle4(t *testing.T) {
	NNodes = 3
	isDebug := true

	var bodyString string
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
	var marshaledStr string
	var offsetSecond int64
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

	var myKey2_4 string

	testCore(t2, bodyString, &myKey2_4, t, isDebug)
	if isDebug {
		t.Logf("myKey2_4: %v\n", myKey2_4)
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

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes2_6 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetMyNodes2_6, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes2_6.Result))

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
	time.Sleep(10 * time.Second)

	// 7.1. join-me
	log.Debug("7.1. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey2_4)

	dataJoinMe2_7 := &pkgservice.BackendJoinRequest{}
	testCore(t2, bodyString, dataJoinMe2_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe2_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe2_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(10 * time.Second)

	// 8. me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_8, t, isDebug)
	assert.Equal(3, len(dataGetMyNodes0_8.Result))
	myNode0_8_0 := dataGetMyNodes0_8.Result[0]
	myNode0_8_1 := dataGetMyNodes0_8.Result[1]
	myNode0_8_2 := dataGetMyNodes0_8.Result[2]

	assert.Equal(types.StatusAlive, myNode0_8_0.Status)
	assert.Equal(types.StatusAlive, myNode0_8_1.Status)
	assert.Equal(types.StatusAlive, myNode0_8_2.Status)

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes1_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_8, t, isDebug)
	assert.Equal(3, len(dataGetMyNodes1_8.Result))
	myNode1_8_0 := dataGetMyNodes1_8.Result[0]
	myNode1_8_1 := dataGetMyNodes1_8.Result[1]
	myNode1_8_2 := dataGetMyNodes1_8.Result[2]

	assert.Equal(types.StatusAlive, myNode1_8_0.Status)
	assert.Equal(types.StatusAlive, myNode1_8_1.Status)
	assert.Equal(types.StatusAlive, myNode1_8_2.Status)

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes2_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetMyNodes2_8, t, isDebug)
	assert.Equal(3, len(dataGetMyNodes2_8.Result))
	myNode2_8_0 := dataGetMyNodes2_8.Result[0]
	myNode2_8_1 := dataGetMyNodes2_8.Result[1]
	myNode2_8_2 := dataGetMyNodes2_8.Result[2]

	assert.Equal(types.StatusAlive, myNode2_8_0.Status)
	assert.Equal(types.StatusAlive, myNode2_8_1.Status)
	assert.Equal(types.StatusAlive, myNode2_8_2.Status)

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

	me2_8_1 := &me.MyInfo{}
	testCore(t2, bodyString, me2_8_1, t, isDebug)
	assert.Equal(types.StatusAlive, me2_8_1.Status)
	assert.Equal(1, len(me2_8_1.OwnerIDs))
	assert.Equal(me1_3.ID, me2_8_1.OwnerIDs[0])
	assert.Equal(true, me2_8_1.IsOwner(me1_3.ID))

	// 9. MasterOplog
	bodyString = `{"id": "testID", "method": "me_getMyMasterOplogList", "params": ["", "", 0, 2]}`

	dataMasterOplogs0_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_9, t, isDebug)
	assert.Equal(5, len(dataMasterOplogs0_9.Result))
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
	assert.Equal(5, len(dataMasterOplogs1_9.Result))
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

	dataMasterOplogs2_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t2, bodyString, dataMasterOplogs2_9, t, isDebug)
	assert.Equal(5, len(dataMasterOplogs2_9.Result))
	masterOplog2_9 := dataMasterOplogs2_9.Result[0]
	assert.Equal(me1_3.ID[:common.AddressLength], masterOplog2_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_3.ID, masterOplog2_9.ObjID)
	assert.Equal(me.MasterOpTypeAddMaster, masterOplog2_9.Op)
	assert.Equal(nilPttID, masterOplog2_9.PreLogID)
	assert.Equal(types.Bool(true), masterOplog2_9.IsSync)
	assert.Equal(masterOplog2_9.ID, masterOplog2_9.MasterLogID)

	// 9.1. getRawMe
	marshaledID, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaledID))

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

	// 10. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_10 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_10, t, isDebug)
	assert.Equal(2, len(dataBoardList0_10.Result))
	board0_10_0 := dataBoardList0_10.Result[0]
	assert.Equal(me0_3.BoardID, board0_10_0.ID)
	assert.Equal(types.StatusMigrated, board0_10_0.Status)

	defaultTitle0_10_0 := content.DefaultTitleTW(me0_1.ID, me1_1.ID, "")
	assert.Equal(defaultTitle0_10_0, board0_10_0.Title)

	board0_10_1 := dataBoardList0_10.Result[1]
	assert.Equal(me1_3.BoardID, board0_10_1.ID)
	assert.Equal(types.StatusAlive, board0_10_1.Status)

	defaultTitle0_10_1 := content.DefaultTitleTW(me1_1.ID, me1_1.ID, "")
	assert.Equal(defaultTitle0_10_1, board0_10_1.Title)

	dataBoardList2_10 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t2, bodyString, dataBoardList2_10, t, isDebug)
	assert.Equal(2, len(dataBoardList2_10.Result))
	board2_10_0 := dataBoardList2_10.Result[0]
	assert.Equal(me2_3.BoardID, board2_10_0.ID)
	assert.Equal(types.StatusMigrated, board2_10_0.Status)

	defaultTitle2_10_0 := content.DefaultTitleTW(me2_1.ID, me1_1.ID, "")
	assert.Equal(defaultTitle2_10_0, board2_10_0.Title)

	board2_10_1 := dataBoardList2_10.Result[1]
	assert.Equal(me1_3.BoardID, board2_10_1.ID)
	assert.Equal(types.StatusAlive, board2_10_1.Status)

	defaultTitle2_10_1 := content.DefaultTitleTW(me1_1.ID, me1_1.ID, "")
	assert.Equal(defaultTitle2_10_1, board2_10_1.Title)

	// 10.1.
	marshaledID, _ = board0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataGetArticleList0_10_1 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testCore(t0, bodyString, dataGetArticleList0_10_1, t, isDebug)

	assert.Equal(0, len(dataGetArticleList0_10_1.Result))

	// 35. t1 create article.
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

	marshaledID, _ = board0_10_1.ID.MarshalText()

	title1_35 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title1_35)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article))
	dataCreateArticle1_35 := &content.BackendCreateArticle{}
	testCore(t1, bodyString, dataCreateArticle1_35, t, isDebug)
	assert.Equal(board0_10_1.ID, dataCreateArticle1_35.BoardID)
	assert.Equal(3, dataCreateArticle1_35.NBlock)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 36. content-get-article-list
	marshaledID, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_36 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_36, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_36.Result))
	article0_36 := dataGetArticleList0_36.Result[0]
	assert.Equal(types.StatusAlive, article0_36.Status)

	dataGetArticleList1_36 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_36, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_36.Result))
	article1_36 := dataGetArticleList1_36.Result[0]
	assert.Equal(types.StatusAlive, article1_36.Status)
	assert.Equal(article0_36.ID, article1_36.ID)

	dataGetArticleList2_36 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleList2_36, t, isDebug)
	assert.Equal(1, len(dataGetArticleList2_36.Result))
	article2_36 := dataGetArticleList2_36.Result[0]
	assert.Equal(types.StatusAlive, article2_36.Status)

	// 37. get-article-block
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

	dataGetArticleBlockList1_37 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_37, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList1_37.Result))

	assert.Equal(article0, dataGetArticleBlockList1_37.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList1_37.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList1_37.Result[2].Buf)

	dataGetArticleBlockList2_37 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleBlockList2_37, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList2_37.Result))

	assert.Equal(article0, dataGetArticleBlockList2_37.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList2_37.Result[1].Buf)
	assert.Equal(article2, dataGetArticleBlockList2_37.Result[2].Buf)

	// 41. time-forward
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_getTimestamp", "params": []}`)

	var sec0_41_0 types.Timestamp
	testCore(t0, bodyString, &sec0_41_0, t, isDebug)

	// 41.1. time-forward.
	offsetSecond = 3600
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_setOffsetSecond", "params": [%v]}`, offsetSecond)

	bool0_41_1 := false
	testCore(t0, bodyString, &bool0_41_1, t, isDebug)

	bool1_41_1 := false
	testCore(t1, bodyString, &bool1_41_1, t, isDebug)

	bool2_41_1 := false
	testCore(t2, bodyString, &bool2_41_1, t, isDebug)

	// 41.2. get offset-second
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_getTimestamp", "params": []}`)

	var sec0_41_2 types.Timestamp
	testCore(t0, bodyString, &sec0_41_2, t, isDebug)
	assert.Equal(sec0_41_0.Ts+offsetSecond, sec0_41_2.Ts)

	// 42.0 get merkle
	marshaledID, _ = board0_10_1.ID.MarshalText()
	t.Logf("42.0 get merkle: marshaledID: %v", marshaledID)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogMerkle", "params": ["%v"]}`, string(marshaledID))

	merkle0_42_0 := &pkgservice.BackendMerkle{}
	testCore(t0, bodyString, merkle0_42_0, t, isDebug)

	merkle1_42_0 := &pkgservice.BackendMerkle{}
	testCore(t1, bodyString, merkle1_42_0, t, isDebug)

	merkle2_42_0 := &pkgservice.BackendMerkle{}
	testCore(t2, bodyString, merkle2_42_0, t, isDebug)

	// 42. sync.
	t.Logf("42. force sync")
	marshaledID, _ = board0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_forceSync", "params": ["%v"]}`, string(marshaledID))

	bool0_42 := false
	testCore(t0, bodyString, &bool0_42, t, isDebug)

	time.Sleep(5 * time.Second)

	bool1_42 := false
	testCore(t1, bodyString, &bool1_42, t, isDebug)

	time.Sleep(5 * time.Second)

	bool2_42 := false
	testCore(t2, bodyString, &bool2_42, t, isDebug)

	time.Sleep(5 * time.Second)

	// 42.2 get merkle
	marshaledID, _ = board0_10_1.ID.MarshalText()
	t.Logf("42.2 get merkle: marshaledID: %v", marshaledID)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogMerkle", "params": ["%v"]}`, string(marshaledID))

	merkle0_42_2 := &pkgservice.BackendMerkle{}
	testCore(t0, bodyString, merkle0_42_2, t, isDebug)

	merkle1_42_2 := &pkgservice.BackendMerkle{}
	testCore(t1, bodyString, merkle1_42_2, t, isDebug)
	assert.Equal(merkle0_42_2.LastSyncTS, merkle1_42_2.LastSyncTS)

	merkle2_42_2 := &pkgservice.BackendMerkle{}
	testCore(t2, bodyString, merkle2_42_2, t, isDebug)
	assert.Equal(merkle0_42_2.LastSyncTS, merkle2_42_2.LastSyncTS)

	// 43. shutdown t0. t2.
	bodyString = `{"id": "testID", "method": "ptt_shutdown", "params": []}`

	resultString := `{"jsonrpc":"2.0","id":"testID","result":true}`
	testBodyEqualCore(t0, bodyString, resultString, t)

	time.Sleep(5 * time.Second)

	testBodyEqualCore(t2, bodyString, resultString, t)

	time.Sleep(5 * time.Second)

	// 44. time-forward.
	now := time.Now()
	remainSeconds := now.Second() % 3600
	offsetSecond = int64(remainSeconds - 10 + 14400)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_setOffsetSecond", "params": [%v]}`, offsetSecond)

	merkle1_44_0 := &pkgservice.BackendMerkle{}
	testCore(t1, bodyString, merkle1_44_0, t, isDebug)

	// 45. t1 create article.
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
	})

	marshaledID, _ = board0_10_1.ID.MarshalText()

	title1_45 := []byte("標題2")
	marshaledStr = base64.StdEncoding.EncodeToString(title1_45)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article))
	dataCreateArticle1_45 := &content.BackendCreateArticle{}
	testCore(t1, bodyString, dataCreateArticle1_45, t, isDebug)
	assert.Equal(board0_10_1.ID, dataCreateArticle1_45.BoardID)
	assert.Equal(2, dataCreateArticle1_45.NBlock)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 46. content-get-article-list
	marshaledID, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataGetArticleList1_46 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_46, t, isDebug)
	assert.Equal(2, len(dataGetArticleList1_46.Result))
	assert.Equal(dataGetArticleList1_36.Result, dataGetArticleList1_46.Result[:1])
	article1_46 := dataGetArticleList1_46.Result[1]

	// 47. get-article-block
	marshaledID2, _ = article1_46.ID.MarshalText()
	marshaledID3, _ = article1_46.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList1_47 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_47, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_47.Result))

	article0 = [][]byte{
		[]byte("測試31"),
	}

	article1 = [][]byte{
		[]byte("測試32"),
		[]byte("測試33"),
		[]byte("測試34"),
		[]byte("測試35"),
		[]byte("測試36"),
		[]byte("測試37"),
		[]byte("測試38"),
		[]byte("測試39"),
		[]byte("測試40"),
	}

	assert.Equal(article0, dataGetArticleBlockList1_47.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList1_47.Result[1].Buf)

	// 48.0. shutdown t1.
	bodyString = `{"id": "testID", "method": "ptt_shutdown", "params": []}`

	resultString = `{"jsonrpc":"2.0","id":"testID","result":true}`
	testBodyEqualCore(t1, bodyString, resultString, t)

	time.Sleep(5 * time.Second)

	// 48. restart t0
	startNode(t, 0, offsetSecond)

	time.Sleep(TimeSleepRestart)

	// 48.1. restart t1
	startNode(t, 1, offsetSecond)

	time.Sleep(TimeSleepRestart)

	// 48.2 offset second
	t.Logf("48.2. get offset second")
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_getOffsetSecond", "params": []}`)

	var offsetSecond0_48_2 int64
	testCore(t0, bodyString, &offsetSecond0_48_2, t, isDebug)
	assert.Equal(offsetSecond, offsetSecond0_48_2)

	var offsetSecond1_48_2 int64
	testCore(t1, bodyString, &offsetSecond1_48_2, t, isDebug)
	assert.Equal(offsetSecond, offsetSecond1_48_2)

	// 49. content-get-article-list
	marshaledID, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataGetArticleList0_49 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_49, t, isDebug)
	assert.Equal(2, len(dataGetArticleList0_49.Result))

	// 50. get-article-block
	marshaledID2, _ = article1_46.ID.MarshalText()
	marshaledID3, _ = article1_46.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_50 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_50, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList0_50.Result))

	assert.Equal(article0, dataGetArticleBlockList0_50.Result[0].Buf)
	assert.Equal(article1, dataGetArticleBlockList0_50.Result[1].Buf)

	// 51.0 get merkle
	marshaledID, _ = board0_10_1.ID.MarshalText()
	t.Logf("51.0 get merkle: marshaledID: %v", marshaledID)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogMerkle", "params": ["%v"]}`, string(marshaledID))

	merkle0_51_0 := &pkgservice.BackendMerkle{}
	testCore(t0, bodyString, merkle0_51_0, t, isDebug)

	merkle1_51_0 := &pkgservice.BackendMerkle{}
	testCore(t1, bodyString, merkle1_51_0, t, isDebug)

	// 52. sync.
	t.Logf("52. force sync")
	marshaledID, _ = board0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_forceSync", "params": ["%v"]}`, string(marshaledID))

	bool0_52 := false
	testCore(t0, bodyString, &bool0_52, t, isDebug)

	time.Sleep(5 * time.Second)

	bool1_52 := false
	testCore(t1, bodyString, &bool1_52, t, isDebug)

	time.Sleep(5 * time.Second)

	// 53.0 get merkle
	marshaledID, _ = board0_10_1.ID.MarshalText()
	t.Logf("53.0 get merkle: marshaledID: %v", board0_10_1.ID)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogMerkle", "params": ["%v"]}`, string(marshaledID))

	merkle0_53_0 := &pkgservice.BackendMerkle{}
	testCore(t0, bodyString, merkle0_53_0, t, isDebug)

	merkle1_53_0 := &pkgservice.BackendMerkle{}
	testCore(t1, bodyString, merkle1_53_0, t, isDebug)

	assert.Equal(merkle0_53_0.LastSyncTS, merkle1_53_0.LastSyncTS)

	// 54. get board oplogs
	marshaledID, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataGetBoardOplogList0_54 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetBoardOplogList0_54, t, isDebug)
	assert.Equal(3, len(dataGetBoardOplogList0_54.Result))
	boardOplog0_54_2 := dataGetBoardOplogList0_54.Result[2]

	dataGetBoardOplogList1_54 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetBoardOplogList1_54, t, isDebug)
	assert.Equal(3, len(dataGetBoardOplogList1_54.Result))
	boardOplog1_54_2 := dataGetBoardOplogList1_54.Result[2]

	assert.Equal(boardOplog0_54_2, boardOplog1_54_2)
}
