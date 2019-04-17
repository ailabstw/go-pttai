// Copyright 2019 The go-pttai Authors
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
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDevice3DeleteArticle(t *testing.T) {
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
	marshaled, _ = board0_10_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataGetArticleList0_10_1 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testCore(t0, bodyString, dataGetArticleList0_10_1, t, isDebug)

	assert.Equal(0, len(dataGetArticleList0_10_1.Result))

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

	marshaled, _ = board0_10_1.ID.MarshalText()

	title0_35 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_35)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaled), marshaledStr, string(article))
	dataCreateArticle0_35 := &content.BackendCreateArticle{}
	testCore(t0, bodyString, dataCreateArticle0_35, t, isDebug)
	assert.Equal(board0_10_1.ID, dataCreateArticle0_35.BoardID)
	assert.Equal(3, dataCreateArticle0_35.NBlock)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 36. content-get-article-list
	marshaled, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaled))
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
	testListCore(t0, bodyString, dataGetArticleList2_36, t, isDebug)
	assert.Equal(1, len(dataGetArticleList2_36.Result))
	article2_36 := dataGetArticleList2_36.Result[0]
	assert.Equal(types.StatusAlive, article2_36.Status)

	// 38. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaled), string(marshaledID2), string(marshaledID3))

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

	// 39. content-create-comment
	t.Logf("39. content-create-comment")
	comment := []byte("這是comment")
	commentStr := base64.StdEncoding.EncodeToString(comment)

	marshaledID, _ = board0_10_1.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("39. content_createComment: bodyString: %v", bodyString)
	dataCreateComment0_39 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment0_39, t, isDebug)
	assert.Equal(dataCreateArticle0_35.ArticleID, dataCreateComment0_39.ArticleID)
	assert.Equal(dataCreateArticle0_35.BoardID, dataCreateComment0_39.BoardID)

	// wait 10 seconds
	time.Sleep(10 * time.Second)

	// 40. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_40 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_40, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList0_40.Result))
	articleBlock0_40 := dataGetArticleBlockList0_40.Result[3]
	assert.Equal(types.StatusAlive, articleBlock0_40.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_40.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_40.CommentType)
	assert.Equal([][]byte{comment}, articleBlock0_40.Buf)

	dataGetArticleBlockList1_40 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_40, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList1_40.Result))
	articleBlock1_40 := dataGetArticleBlockList1_40.Result[3]
	assert.Equal(types.StatusAlive, articleBlock1_40.Status)
	assert.Equal(content.ContentTypeComment, articleBlock1_40.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock1_40.CommentType)
	assert.Equal([][]byte{comment}, articleBlock1_40.Buf)

	dataGetArticleBlockList2_40 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleBlockList2_40, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList2_40.Result))
	articleBlock2_40 := dataGetArticleBlockList2_40.Result[3]
	assert.Equal(types.StatusAlive, articleBlock2_40.Status)
	assert.Equal(content.ContentTypeComment, articleBlock2_40.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock2_40.CommentType)
	assert.Equal([][]byte{comment}, articleBlock2_40.Buf)

	// 41. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 1, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), "")

	dataGetArticleBlockList0_41 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_41, t, isDebug)
	assert.Equal(1, len(dataGetArticleBlockList0_41.Result))
	articleBlock0_41 := dataGetArticleBlockList0_41.Result[0]
	assert.Equal(articleBlock0_40, articleBlock0_41)

	dataGetArticleBlockList1_41 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_41, t, isDebug)
	assert.Equal(1, len(dataGetArticleBlockList1_41.Result))
	articleBlock1_41 := dataGetArticleBlockList1_41.Result[0]
	assert.Equal(articleBlock1_40, articleBlock1_41)

	dataGetArticleBlockList2_41 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleBlockList2_41, t, isDebug)
	assert.Equal(1, len(dataGetArticleBlockList2_41.Result))
	articleBlock2_41 := dataGetArticleBlockList2_41.Result[0]
	assert.Equal(articleBlock2_40, articleBlock2_41)

	// 42. content-get-article-list
	marshaledID, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_42 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_42, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_42.Result))
	article0_42 := dataGetArticleList0_42.Result[0]

	assert.Equal(article0_36.ID, article0_42.ID)
	assert.Equal(1, article0_42.NPush)
	assert.Equal(0, article0_42.NBoo)
	assert.Equal(articleBlock0_40.UpdateTS, article0_42.CommentCreateTS)
	assert.Equal(true, article0_42.CreateTS.IsLess(article0_42.CommentCreateTS))

	dataGetArticleList1_42 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_42, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_42.Result))
	article1_42 := dataGetArticleList1_42.Result[0]

	assert.Equal(article1_36.ID, article1_42.ID)
	assert.Equal(1, article1_42.NPush)
	assert.Equal(0, article1_42.NBoo)
	assert.Equal(articleBlock1_40.UpdateTS, article1_42.CommentCreateTS)
	assert.Equal(true, article1_42.CreateTS.IsLess(article1_42.CommentCreateTS))

	dataGetArticleList2_42 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleList2_42, t, isDebug)
	assert.Equal(1, len(dataGetArticleList2_42.Result))
	article2_42 := dataGetArticleList2_42.Result[0]

	assert.Equal(article2_36.ID, article2_42.ID)
	assert.Equal(1, article2_42.NPush)
	assert.Equal(0, article2_42.NBoo)
	assert.Equal(articleBlock2_40.UpdateTS, article2_42.CommentCreateTS)
	assert.Equal(true, article2_42.CreateTS.IsLess(article2_42.CommentCreateTS))

	// 43. content-create-comment
	commentBytes0_43 := []byte("這是comment43")
	commentStr = base64.StdEncoding.EncodeToString(commentBytes0_43)

	marshaledID, _ = board0_10_1.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("43. content_createComment: bodyString: %v", bodyString)
	dataCreateComment0_43 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment0_43, t, isDebug)
	assert.Equal(dataCreateArticle0_35.ArticleID, dataCreateComment0_43.ArticleID)
	assert.Equal(dataCreateArticle0_35.BoardID, dataCreateComment0_43.BoardID)

	// 44. content-create-comment
	commentBytes0_44 := []byte("這是comment44")
	commentStr = base64.StdEncoding.EncodeToString(commentBytes0_44)

	marshaledID, _ = board0_10_1.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 1, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("44. content_createComment: bodyString: %v", bodyString)
	dataCreateComment0_44 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment0_44, t, isDebug)
	assert.Equal(dataCreateArticle0_35.ArticleID, dataCreateComment0_44.ArticleID)
	assert.Equal(dataCreateArticle0_35.BoardID, dataCreateComment0_44.BoardID)

	// 45. content-create-comment
	commentBytes0_45 := []byte("這是comment45")
	commentStr = base64.StdEncoding.EncodeToString(commentBytes0_45)

	marshaledID, _ = board0_10_1.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("45. content_createComment: bodyString: %v", bodyString)
	dataCreateComment0_45 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment0_45, t, isDebug)
	assert.Equal(dataCreateArticle0_35.ArticleID, dataCreateComment0_45.ArticleID)
	assert.Equal(dataCreateArticle0_35.BoardID, dataCreateComment0_45.BoardID)

	// wait 10 seconds
	time.Sleep(10 * time.Second)

	// 46. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_46 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_46, t, isDebug)
	assert.Equal(7, len(dataGetArticleBlockList0_46.Result))
	articleBlock0_46_4 := dataGetArticleBlockList0_46.Result[4]
	assert.Equal(types.StatusAlive, articleBlock0_46_4.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_46_4.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_46_4.CommentType)
	assert.Equal([][]byte{commentBytes0_43}, articleBlock0_46_4.Buf)

	articleBlock0_46_5 := dataGetArticleBlockList0_46.Result[5]
	assert.Equal(types.StatusAlive, articleBlock0_46_5.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_46_5.ContentType)
	assert.Equal(content.CommentTypeBoo, articleBlock0_46_5.CommentType)
	assert.Equal([][]byte{commentBytes0_44}, articleBlock0_46_5.Buf)

	articleBlock0_46_6 := dataGetArticleBlockList0_46.Result[6]
	assert.Equal(types.StatusAlive, articleBlock0_46_6.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_46_6.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_46_6.CommentType)
	assert.Equal([][]byte{commentBytes0_45}, articleBlock0_46_6.Buf)

	dataGetArticleBlockList1_46 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_46, t, isDebug)
	assert.Equal(7, len(dataGetArticleBlockList1_46.Result))
	articleBlock1_46_4 := dataGetArticleBlockList1_46.Result[4]
	assert.Equal(types.StatusAlive, articleBlock1_46_4.Status)
	assert.Equal(content.ContentTypeComment, articleBlock1_46_4.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock1_46_4.CommentType)
	assert.Equal([][]byte{commentBytes0_43}, articleBlock1_46_4.Buf)

	articleBlock1_46_5 := dataGetArticleBlockList1_46.Result[5]
	assert.Equal(types.StatusAlive, articleBlock1_46_5.Status)
	assert.Equal(content.ContentTypeComment, articleBlock1_46_5.ContentType)
	assert.Equal(content.CommentTypeBoo, articleBlock1_46_5.CommentType)
	assert.Equal([][]byte{commentBytes0_44}, articleBlock1_46_5.Buf)

	articleBlock1_46_6 := dataGetArticleBlockList1_46.Result[6]
	assert.Equal(types.StatusAlive, articleBlock1_46_6.Status)
	assert.Equal(content.ContentTypeComment, articleBlock1_46_6.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock1_46_6.CommentType)
	assert.Equal([][]byte{commentBytes0_45}, articleBlock1_46_6.Buf)

	dataGetArticleBlockList2_46 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleBlockList2_46, t, isDebug)
	assert.Equal(7, len(dataGetArticleBlockList2_46.Result))
	articleBlock2_46_4 := dataGetArticleBlockList2_46.Result[4]
	assert.Equal(types.StatusAlive, articleBlock2_46_4.Status)
	assert.Equal(content.ContentTypeComment, articleBlock2_46_4.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock2_46_4.CommentType)
	assert.Equal([][]byte{commentBytes0_43}, articleBlock2_46_4.Buf)

	articleBlock2_46_5 := dataGetArticleBlockList2_46.Result[5]
	assert.Equal(types.StatusAlive, articleBlock2_46_5.Status)
	assert.Equal(content.ContentTypeComment, articleBlock2_46_5.ContentType)
	assert.Equal(content.CommentTypeBoo, articleBlock2_46_5.CommentType)
	assert.Equal([][]byte{commentBytes0_44}, articleBlock2_46_5.Buf)

	articleBlock2_46_6 := dataGetArticleBlockList2_46.Result[6]
	assert.Equal(types.StatusAlive, articleBlock2_46_6.Status)
	assert.Equal(content.ContentTypeComment, articleBlock2_46_6.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock2_46_6.CommentType)
	assert.Equal([][]byte{commentBytes0_45}, articleBlock2_46_6.Buf)

	// 46.1. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "", 1, 0, 10, 2]}`, string(marshaledID), string(marshaledID2))

	dataGetArticleBlockList0_46_1 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}

	testListCore(t0, bodyString, dataGetArticleBlockList0_46_1, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList0_46_1.Result))
	assert.Equal(dataGetArticleBlockList0_46.Result[3:], dataGetArticleBlockList0_46_1.Result)

	dataGetArticleBlockList1_46_1 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}

	testListCore(t1, bodyString, dataGetArticleBlockList1_46_1, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList1_46_1.Result))
	assert.Equal(dataGetArticleBlockList1_46.Result[3:], dataGetArticleBlockList1_46_1.Result)

	dataGetArticleBlockList2_46_1 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}

	testListCore(t2, bodyString, dataGetArticleBlockList2_46_1, t, isDebug)
	assert.Equal(4, len(dataGetArticleBlockList2_46_1.Result))
	assert.Equal(dataGetArticleBlockList2_46.Result[3:], dataGetArticleBlockList2_46_1.Result)

	// 48. content-delete-article
	marshaledID, _ = board0_10_1.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_deleteArticle", "params": ["%v", "%v"]}`, string(marshaledID), string(marshaledID2))
	dataDeleteArticle0_48 := &content.BackendDeleteArticle{}
	testCore(t0, bodyString, dataDeleteArticle0_48, t, isDebug)

	// wait 10 seconds
	time.Sleep(10 * time.Second)

	// 49. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_49 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_49, t, isDebug)
	assert.Equal(0, len(dataGetArticleBlockList0_49.Result))

	dataGetArticleBlockList1_49 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_49, t, isDebug)
	assert.Equal(0, len(dataGetArticleBlockList1_49.Result))

	dataGetArticleBlockList2_49 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleBlockList2_49, t, isDebug)
	assert.Equal(0, len(dataGetArticleBlockList2_49.Result))

	// 50. content-get-article-list
	marshaledID, _ = board0_10_1.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_50 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_50, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_50.Result))
	article0_50 := dataGetArticleList0_50.Result[0]

	assert.Equal(0, article0_50.NPush)
	assert.Equal(0, article0_50.NBoo)
	assert.Equal(types.StatusDeleted, article0_50.Status)

	dataGetArticleList1_50 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_50, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_50.Result))
	article1_50 := dataGetArticleList1_50.Result[0]

	assert.Equal(0, article1_50.NPush)
	assert.Equal(0, article1_50.NBoo)
	assert.Equal(types.StatusDeleted, article1_50.Status)

	dataGetArticleList2_50 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t2, bodyString, dataGetArticleList2_50, t, isDebug)
	assert.Equal(1, len(dataGetArticleList2_50.Result))
	article2_50 := dataGetArticleList2_50.Result[0]

	assert.Equal(0, article2_50.NPush)
	assert.Equal(0, article2_50.NBoo)
	assert.Equal(types.StatusDeleted, article2_50.Status)
}
