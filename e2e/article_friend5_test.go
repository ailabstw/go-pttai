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
	"io/ioutil"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestArticleFriend5(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
	var marshaledStr string
	var marshaledStr2 string
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
	nodeID0_1 := me0_1.NodeID
	pubKey0_1, _ := nodeID0_1.Pubkey()
	nodeAddr0_1 := crypto.PubkeyToAddress(*pubKey0_1)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)

	// 2. get total weight
	bodyString = `{"id": "testID", "method": "me_getTotalWeight", "params": [""]}`

	var totalWeigtht0_2 uint32
	testCore(t0, bodyString, &totalWeigtht0_2, t, isDebug)

	assert.Equal(uint32(me.WeightDesktop), totalWeigtht0_2)

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

	// 4. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_4 string

	testCore(t0, bodyString, &myKey0_4, t, isDebug)
	if isDebug {
		t.Logf("myKey0_4: %v\n", myKey0_4)
	}

	// 4.1 validate-my-key

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_validateMyKey", "params": ["%v"]}`, string(myKey0_4))

	var isValid0_4_1 bool

	testCore(t0, bodyString, &isValid0_4_1, t, isDebug)
	assert.Equal(true, isValid0_4_1)

	time.Sleep(5 * time.Second)

	// 6. getJoinKeyInfo
	bodyString = `{"id": "testID", "method": "me_getJoinKeyInfos", "params": [""]}`

	dataJoinKeyInfos0_6 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataJoinKeyInfos0_6, t, isDebug)
	assert.NotEqual(0, len(dataJoinKeyInfos0_6.Result))

	// 7. raft-satus
	marshaledID, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRaftStatus", "params": ["%v"]}`, string(marshaledID))

	raftStatus0_7 := &me.RaftStatus{}
	testCore(t0, bodyString, raftStatus0_7, t, isDebug)

	assert.Equal(me0_1.RaftID, raftStatus0_7.Lead)
	assert.Equal(me0_1.RaftID, raftStatus0_7.ConfState.Nodes[0])
	assert.Equal(uint32(me.WeightDesktop), raftStatus0_7.ConfState.Weights[0])

	// 8. getOpKeyInfo
	bodyString = `{"id": "testID", "method": "me_getOpKeyInfos", "params": [""]}`

	dataOpKeyInfos0_8 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfos0_8, t, isDebug)
	assert.Equal(1, len(dataOpKeyInfos0_8.Result))
	opKeyInfo0_8 := dataOpKeyInfos0_8.Result[0]

	// 8.1 ptt.GetOps
	bodyString = `{"id": "testID", "method": "ptt_getOps", "params": []}`

	dataOpKeyInfo0_8_1 := &struct {
		Result map[common.Address]*types.PttID `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyInfo0_8_1, t, isDebug)

	opKeyInfoMap0_8_1 := dataOpKeyInfo0_8_1.Result
	entityID, ok := opKeyInfoMap0_8_1[*opKeyInfo0_8.Hash]

	assert.Equal(true, ok)
	assert.Equal(me0_3.ID, entityID)

	// 9. MasterOplog
	bodyString = `{"id": "testID", "method": "me_getMyMasterOplogList", "params": ["", "", 0, 2]}`

	dataMasterOplogs0_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_9, t, isDebug)
	assert.Equal(1, len(dataMasterOplogs0_9.Result))
	masterOplog0_9 := dataMasterOplogs0_9.Result[0]
	assert.Equal(me0_3.ID[:common.AddressLength], masterOplog0_9.CreatorID[common.AddressLength:])
	assert.Equal(me0_3.ID, masterOplog0_9.ObjID)
	assert.Equal(me.MasterOpTypeAddMaster, masterOplog0_9.Op)
	assert.Equal(nilPttID, masterOplog0_9.PreLogID)
	assert.Equal(types.Bool(true), masterOplog0_9.IsSync)
	assert.Equal(masterOplog0_9.ID, masterOplog0_9.MasterLogID)

	// 9.1. OpKeyOplog
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getOpKeyOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataOpKeyOplogs0_9_1 := &struct {
		Result []*pkgservice.OpKeyOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyOplogs0_9_1, t, isDebug)
	assert.Equal(1, len(dataOpKeyOplogs0_9_1.Result))
	opKeyOplog0_9_1 := dataOpKeyOplogs0_9_1.Result[0]
	assert.Equal(me0_3.ID, opKeyOplog0_9_1.CreatorID)
	assert.Equal(opKeyInfo0_8.ID, opKeyOplog0_9_1.ObjID)
	assert.Equal(pkgservice.OpKeyOpTypeCreateOpKey, opKeyOplog0_9_1.Op)
	assert.Equal(nilPttID, opKeyOplog0_9_1.PreLogID)
	assert.Equal(types.Bool(true), opKeyOplog0_9_1.IsSync)
	assert.Equal(masterOplog0_9.ID, opKeyOplog0_9_1.MasterLogID)
	assert.Equal(1, len(opKeyOplog0_9_1.MasterSigns))
	masterSign0_9_1 := opKeyOplog0_9_1.MasterSigns[0]
	assert.Equal(nodeAddr0_1[:], masterSign0_9_1.ID[:common.AddressLength])
	assert.Equal(me0_3.ID[:common.AddressLength], masterSign0_9_1.ID[common.AddressLength:])

	// 9.2. MeOplog
	bodyString = `{"id": "testID", "method": "me_getMeOplogList", "params": ["", 0, 2]}`

	dataMeOplogs0_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMeOplogs0_9_2, t, isDebug)
	assert.Equal(1, len(dataMeOplogs0_9_2.Result))
	meOplog0_9_2 := dataMeOplogs0_9_2.Result[0]
	assert.Equal(me0_3.ID, meOplog0_9_2.CreatorID)
	assert.Equal(me0_3.ID, meOplog0_9_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog0_9_2.Op)
	assert.Equal(nilPttID, meOplog0_9_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog0_9_2.IsSync)
	assert.Equal(masterOplog0_9.ID, meOplog0_9_2.MasterLogID)
	assert.Equal(me0_3.LogID, meOplog0_9_2.ID)
	masterSign0_9_2 := meOplog0_9_2.MasterSigns[0]
	assert.Equal(nodeAddr0_1[:], masterSign0_9_2.ID[:common.AddressLength])
	assert.Equal(me0_3.ID[:common.AddressLength], masterSign0_9_2.ID[common.AddressLength:])

	// 10. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_10 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_10, t, isDebug)
	assert.Equal(1, len(dataBoardList0_10.Result))
	board0_10_0 := dataBoardList0_10.Result[0]
	assert.Equal(me0_3.BoardID, board0_10_0.ID)
	assert.Equal(types.StatusAlive, board0_10_0.Status)

	defaultTitle0_10_0 := content.DefaultTitleTW(me0_1.ID, me0_1.ID, "")
	assert.Equal(defaultTitle0_10_0, board0_10_0.Title)

	dataBoardList1_10 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t1, bodyString, dataBoardList1_10, t, isDebug)
	assert.Equal(1, len(dataBoardList1_10.Result))
	board1_10_0 := dataBoardList1_10.Result[0]
	assert.Equal(me1_3.BoardID, board1_10_0.ID)
	assert.Equal(types.StatusAlive, board1_10_0.Status)

	defaultTitle1_10_0 := content.DefaultTitleTW(me1_1.ID, me1_1.ID, "")
	assert.Equal(defaultTitle1_10_0, board1_10_0.Title)

	// 10.1.
	marshaledID, _ = board0_10_0.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataGetArticleList0_10_1 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testCore(t0, bodyString, dataGetArticleList0_10_1, t, isDebug)

	assert.Equal(0, len(dataGetArticleList0_10_1.Result))

	// 13. upload file
	marshaledID, _ = me0_3.BoardID.MarshalText()
	file0_13, _ := ioutil.ReadFile("./e2e-test.zip")
	marshaledStr = base64.StdEncoding.EncodeToString(file0_13)
	marshaledStr2 = "e2e-test.zip"

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_uploadFile", "params": ["%v", "%v", "%v"]}`, string(marshaledID), marshaledStr2, marshaledStr)

	dataUploadFile0_13 := &content.BackendUploadFile{}
	testCore(t0, bodyString, dataUploadFile0_13, t, isDebug)

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

	marshaledID, _ = board0_10_0.ID.MarshalText()

	title0_35 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_35)
	marshaledID2, _ = dataUploadFile0_13.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, ["%v"]]}`, string(marshaledID), marshaledStr, string(article), string(marshaledID2))
	dataCreateArticle0_35 := &content.BackendCreateArticle{}
	testCore(t0, bodyString, dataCreateArticle0_35, t, isDebug)
	assert.Equal(board0_10_0.ID, dataCreateArticle0_35.BoardID)
	assert.Equal(3, dataCreateArticle0_35.NBlock)

	// 36. content-get-article-list
	marshaledID, _ = board0_10_0.ID.MarshalText()

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

	// 39. create-article
	article0_39, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試101")),
		base64.StdEncoding.EncodeToString([]byte("測試102")),
		base64.StdEncoding.EncodeToString([]byte("測試103")),
		base64.StdEncoding.EncodeToString([]byte("測試104")),
		base64.StdEncoding.EncodeToString([]byte("測試105")),
		base64.StdEncoding.EncodeToString([]byte("測試106")),
		base64.StdEncoding.EncodeToString([]byte("測試107")),
		base64.StdEncoding.EncodeToString([]byte("測試108")),
		base64.StdEncoding.EncodeToString([]byte("測試109")),
		base64.StdEncoding.EncodeToString([]byte("測試110")),
		base64.StdEncoding.EncodeToString([]byte("測試111")),
		base64.StdEncoding.EncodeToString([]byte("測試112")),
		base64.StdEncoding.EncodeToString([]byte("測試113")),
		base64.StdEncoding.EncodeToString([]byte("測試114")),
		base64.StdEncoding.EncodeToString([]byte("測試115")),
		base64.StdEncoding.EncodeToString([]byte("測試116")),
		base64.StdEncoding.EncodeToString([]byte("測試117")),
		base64.StdEncoding.EncodeToString([]byte("測試118")),
		base64.StdEncoding.EncodeToString([]byte("測試119")),
	})

	marshaledID, _ = board0_10_0.ID.MarshalText()

	title0_39 := []byte("標題0_1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_39)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article0_39))
	dataCreateArticle0_39 := &content.BackendCreateArticle{}
	testCore(t0, bodyString, dataCreateArticle0_39, t, isDebug)
	assert.Equal(board0_10_0.ID, dataCreateArticle0_39.BoardID)
	assert.Equal(2, dataCreateArticle0_39.NBlock)

	// 40. content-get-article-list
	marshaledID, _ = board0_10_0.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_40 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_40, t, isDebug)
	assert.Equal(2, len(dataGetArticleList0_40.Result))
	article0_40_0 := dataGetArticleList0_40.Result[0]
	assert.Equal(article0_36, article0_40_0)
	article0_40_1 := dataGetArticleList0_40.Result[1]

	// 41. get-article-block
	marshaledID2, _ = article0_40_1.ID.MarshalText()
	marshaledID3, _ = article0_40_1.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_41 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_41, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList0_41.Result))

	article41_0 := [][]byte{
		[]byte("測試101"),
	}

	article41_1 := [][]byte{
		[]byte("測試102"),
		[]byte("測試103"),
		[]byte("測試104"),
		[]byte("測試105"),
		[]byte("測試106"),
		[]byte("測試107"),
		[]byte("測試108"),
		[]byte("測試109"),
		[]byte("測試110"),
		[]byte("測試111"),
		[]byte("測試112"),
		[]byte("測試113"),
		[]byte("測試114"),
		[]byte("測試115"),
		[]byte("測試116"),
		[]byte("測試117"),
		[]byte("測試118"),
		[]byte("測試119"),
	}

	assert.Equal(article41_0, dataGetArticleBlockList0_41.Result[0].Buf)
	assert.Equal(article41_1, dataGetArticleBlockList0_41.Result[1].Buf)

	// 39.1. update-article
	article39_1, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試161")),
		base64.StdEncoding.EncodeToString([]byte("測試162")),
		base64.StdEncoding.EncodeToString([]byte("測試163")),
		base64.StdEncoding.EncodeToString([]byte("測試164")),
		base64.StdEncoding.EncodeToString([]byte("測試165")),
		base64.StdEncoding.EncodeToString([]byte("測試166")),
		base64.StdEncoding.EncodeToString([]byte("測試167")),
		base64.StdEncoding.EncodeToString([]byte("測試168")),
		base64.StdEncoding.EncodeToString([]byte("測試169")),
	})

	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_updateArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), string(marshaledID2), string(article39_1))
	dataUpdateArticle0_39_1 := &content.BackendUpdateArticle{}
	testCore(t0, bodyString, dataUpdateArticle0_39_1, t, isDebug)
	assert.Equal(board0_10_0.ID, dataUpdateArticle0_39_1.BoardID)
	assert.Equal(article0_36.ID, dataUpdateArticle0_39_1.ArticleID)
	assert.Equal(2, dataUpdateArticle0_39_1.NBlock)

	// 39.2. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = dataUpdateArticle0_39_1.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_39_2 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_39_2, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList0_39_2.Result))

	article39_2_0 := [][]byte{
		[]byte("測試161"),
	}

	article39_2_1 := [][]byte{
		[]byte("測試162"),
		[]byte("測試163"),
		[]byte("測試164"),
		[]byte("測試165"),
		[]byte("測試166"),
		[]byte("測試167"),
		[]byte("測試168"),
		[]byte("測試169"),
	}

	assert.Equal(article39_2_0, dataGetArticleBlockList0_39_2.Result[0].Buf)
	assert.Equal(article39_2_1, dataGetArticleBlockList0_39_2.Result[1].Buf)

	// 42.0. upload file
	marshaledID, _ = me1_3.BoardID.MarshalText()
	file1_42_0, _ := ioutil.ReadFile("./e2e-test.zip")
	marshaledStr = base64.StdEncoding.EncodeToString(file1_42_0)
	marshaledStr2 = "e2e-test.zip"

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_uploadFile", "params": ["%v", "%v", "%v"]}`, string(marshaledID), marshaledStr2, marshaledStr)

	dataUploadFile1_42_0 := &content.BackendUploadFile{}
	testCore(t1, bodyString, dataUploadFile1_42_0, t, isDebug)

	// 42. create-article
	article1_42, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試201")),
		base64.StdEncoding.EncodeToString([]byte("測試202")),
		base64.StdEncoding.EncodeToString([]byte("測試203")),
		base64.StdEncoding.EncodeToString([]byte("測試204")),
		base64.StdEncoding.EncodeToString([]byte("測試205")),
		base64.StdEncoding.EncodeToString([]byte("測試206")),
		base64.StdEncoding.EncodeToString([]byte("測試207")),
		base64.StdEncoding.EncodeToString([]byte("測試208")),
		base64.StdEncoding.EncodeToString([]byte("測試209")),
	})

	marshaledID, _ = board1_10_0.ID.MarshalText()

	title1_42 := []byte("標題1_42")
	marshaledStr = base64.StdEncoding.EncodeToString(title1_42)
	marshaledID2, _ = dataUploadFile1_42_0.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, ["%v"]]}`, string(marshaledID), marshaledStr, string(article1_42), string(marshaledID2))
	dataCreateArticle1_42 := &content.BackendCreateArticle{}
	testCore(t1, bodyString, dataCreateArticle1_42, t, isDebug)
	assert.Equal(board1_10_0.ID, dataCreateArticle1_42.BoardID)
	assert.Equal(2, dataCreateArticle1_42.NBlock)

	// 42.1. content-get-article-list
	marshaledID, _ = board1_10_0.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList1_42_1 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_42_1, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_42_1.Result))
	article1_42_1_0 := dataGetArticleList1_42_1.Result[0]

	// 42.2. get-article-block
	marshaledID2, _ = article1_42_1_0.ID.MarshalText()
	marshaledID3, _ = article1_42_1_0.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList1_42_2 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_42_2, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_42_2.Result))

	article42_0 := [][]byte{
		[]byte("測試201"),
	}

	article42_1 := [][]byte{
		[]byte("測試202"),
		[]byte("測試203"),
		[]byte("測試204"),
		[]byte("測試205"),
		[]byte("測試206"),
		[]byte("測試207"),
		[]byte("測試208"),
		[]byte("測試209"),
	}

	assert.Equal(article42_0, dataGetArticleBlockList1_42_2.Result[0].Buf)
	assert.Equal(article42_1, dataGetArticleBlockList1_42_2.Result[1].Buf)

	// 43. create-article
	article1_43, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試301")),
		base64.StdEncoding.EncodeToString([]byte("測試302")),
		base64.StdEncoding.EncodeToString([]byte("測試303")),
		base64.StdEncoding.EncodeToString([]byte("測試304")),
		base64.StdEncoding.EncodeToString([]byte("測試305")),
	})

	marshaledID, _ = board1_10_0.ID.MarshalText()

	title1_43 := []byte("標題1_43")
	marshaledStr = base64.StdEncoding.EncodeToString(title1_43)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article1_43))
	dataCreateArticle1_43 := &content.BackendCreateArticle{}
	testCore(t1, bodyString, dataCreateArticle1_43, t, isDebug)
	assert.Equal(board1_10_0.ID, dataCreateArticle1_43.BoardID)
	assert.Equal(2, dataCreateArticle1_43.NBlock)

	// 43.1. content-get-article-list
	marshaledID, _ = board1_10_0.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList1_43_1 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_43_1, t, isDebug)
	assert.Equal(2, len(dataGetArticleList1_43_1.Result))
	article1_43_1_0 := dataGetArticleList1_43_1.Result[0]
	assert.Equal(article1_43_1_0, article1_42_1_0)
	article1_43_1_1 := dataGetArticleList1_43_1.Result[1]

	// 43.2. get-article-block
	marshaledID2, _ = article1_43_1_1.ID.MarshalText()
	marshaledID3, _ = article1_43_1_1.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList1_43_2 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_43_2, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_43_2.Result))

	article43_0 := [][]byte{
		[]byte("測試301"),
	}

	article43_1 := [][]byte{
		[]byte("測試302"),
		[]byte("測試303"),
		[]byte("測試304"),
		[]byte("測試305"),
	}

	assert.Equal(article43_0, dataGetArticleBlockList1_43_2.Result[0].Buf)
	assert.Equal(article43_1, dataGetArticleBlockList1_43_2.Result[1].Buf)

	// 45. show-url
	bodyString = `{"id": "testID", "method": "me_showURL", "params": []}`

	dataShowURL1_45 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowURL1_45, t, isDebug)
	url1_45 := dataShowURL1_45.URL

	// 47. join-friend
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_45)

	dataJoinFriend0_47 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinFriend0_47, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend0_47.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend0_47.NodeID)

	// wait 10
	t.Logf("wait 15 seconds for hand-shaking")
	time.Sleep(15 * time.Second)

	// 8. get-friend-list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList0_48 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetFriendList0_48, t, isDebug)
	assert.Equal(1, len(dataGetFriendList0_48.Result))
	friend0_48 := dataGetFriendList0_48.Result[0]
	assert.Equal(types.StatusAlive, friend0_48.Status)
	assert.Equal(me1_1.ID, friend0_48.FriendID)

	dataGetFriendList1_48 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_48, t, isDebug)
	assert.Equal(1, len(dataGetFriendList1_48.Result))
	friend1_48 := dataGetFriendList1_48.Result[0]
	assert.Equal(types.StatusAlive, friend1_48.Status)
	assert.Equal(me0_1.ID, friend1_48.FriendID)
	assert.Equal(friend0_48.ID, friend1_48.ID)

	// 50.0. get raw article
	t.Logf("50.0: content.GetRawArticle")
	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawArticle", "params": ["%v", "%v"]}`, string(marshaledID), string(marshaledID2))

	article0_50_0 := &content.Article{}
	testCore(t0, bodyString, article0_50_0, t, isDebug)
	assert.Equal(article0_36.ID, article0_50_0.ID)

	article1_50_0 := &content.Article{}
	testCore(t1, bodyString, article1_50_0, t, isDebug)
	assert.Equal(article0_36.ID, article1_50_0.ID)

	// 50. get-article-block
	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_50 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_50, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList0_50.Result))

	assert.Equal(article39_2_0, dataGetArticleBlockList0_50.Result[0].Buf)
	assert.Equal(article39_2_1, dataGetArticleBlockList0_50.Result[1].Buf)

	dataGetArticleBlockList1_50 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_50, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_50.Result))

	assert.Equal(article39_2_0, dataGetArticleBlockList1_50.Result[0].Buf)
	assert.Equal(article39_2_1, dataGetArticleBlockList1_50.Result[1].Buf)

	// 51. get-article-block
	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_40_1.ID.MarshalText()
	marshaledID3, _ = article0_40_1.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_51 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_51, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList0_51.Result))

	assert.Equal(article41_0, dataGetArticleBlockList0_51.Result[0].Buf)
	assert.Equal(article41_1, dataGetArticleBlockList0_51.Result[1].Buf)

	dataGetArticleBlockList1_51 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_51, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_51.Result))

	assert.Equal(article41_0, dataGetArticleBlockList1_51.Result[0].Buf)
	assert.Equal(article41_1, dataGetArticleBlockList1_51.Result[1].Buf)
}
