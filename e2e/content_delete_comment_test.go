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
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestContentDeleteComment(t *testing.T) {
	NNodes = 1
	isDebug := true

	var bodyString string
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
	var marshaledStr string
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}

	testCore(t0, bodyString, me0_1, t, isDebug)

	assert.Equal(types.StatusAlive, me0_1.Status)
	nodeID0_1 := me0_1.NodeID
	pubKey0_1, _ := nodeID0_1.Pubkey()
	nodeAddr0_1 := crypto.PubkeyToAddress(*pubKey0_1)

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
	assert.Equal(types.ZeroTimestamp, board0_10_0.ArticleCreateTS)

	// 10.1.
	marshaledID, _ = board0_10_0.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

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

	marshaledID, _ = board0_10_0.ID.MarshalText()

	title0_35 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_35)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), marshaledStr, string(article))
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

	assert.Equal(0, article0_36.NPush)
	assert.Equal(0, article0_36.NBoo)
	assert.Equal(types.ZeroTimestamp, article0_36.CommentCreateTS)

	// 36.1 content-get-board
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_36_1 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_36_1, t, isDebug)
	assert.Equal(1, len(dataBoardList0_36_1.Result))
	board0_36_1_0 := dataBoardList0_36_1.Result[0]
	assert.Equal(me0_3.BoardID, board0_36_1_0.ID)
	assert.Equal(types.StatusAlive, board0_36_1_0.Status)

	defaultTitle0_36_1_0 := content.DefaultTitleTW(me0_1.ID, me0_1.ID, "")
	assert.Equal(defaultTitle0_36_1_0, board0_36_1_0.Title)
	assert.Equal(article0_36.UpdateTS, board0_36_1_0.ArticleCreateTS)

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

	// 39. content-create-comment
	comment := []byte("這是comment")
	commentStr := base64.StdEncoding.EncodeToString(comment)

	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("39. content_createComment: bodyString: %v", bodyString)
	dataCreateComment0_39 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment0_39, t, isDebug)
	assert.Equal(dataCreateArticle0_35.ArticleID, dataCreateComment0_39.ArticleID)
	assert.Equal(dataCreateArticle0_35.BoardID, dataCreateComment0_39.BoardID)

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

	// 42. content-get-article-list
	marshaledID, _ = board0_10_0.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_42 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_42, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_42.Result))
	article0_42 := dataGetArticleList0_42.Result[0]

	assert.Equal(1, article0_42.NPush)
	assert.Equal(0, article0_42.NBoo)
	assert.Equal(articleBlock0_40.UpdateTS, article0_42.CommentCreateTS)
	assert.Equal(true, article0_42.CreateTS.IsLess(article0_42.CommentCreateTS))

	// 43. content-create-comment
	commentBytes0_43 := []byte("這是comment43")
	commentStr = base64.StdEncoding.EncodeToString(commentBytes0_43)

	marshaledID, _ = board0_10_0.ID.MarshalText()
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

	marshaledID, _ = board0_10_0.ID.MarshalText()
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

	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("45. content_createComment: bodyString: %v", bodyString)
	dataCreateComment0_45 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment0_45, t, isDebug)
	assert.Equal(dataCreateArticle0_35.ArticleID, dataCreateComment0_45.ArticleID)
	assert.Equal(dataCreateArticle0_35.BoardID, dataCreateComment0_45.BoardID)

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

	// 47. content-get-article-list
	marshaledID, _ = board0_10_0.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_47 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_47, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_47.Result))
	article0_47 := dataGetArticleList0_47.Result[0]

	assert.Equal(3, article0_47.NPush)
	assert.Equal(1, article0_47.NBoo)
	assert.Equal(articleBlock0_46_6.CreateTS, article0_47.CommentCreateTS)

	// 48. content-delete-comment
	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = articleBlock0_46_5.RefID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_deleteComment", "params": ["%v", "%v", "%v"]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))
	t.Logf("45. content_createComment: bodyString: %v", bodyString)
	dataDeleteComment0_48 := &content.BackendDeleteComment{}
	testCore(t0, bodyString, dataDeleteComment0_48, t, isDebug)

	// 49. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_49 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_49, t, isDebug)
	assert.Equal(7, len(dataGetArticleBlockList0_49.Result))
	articleBlock0_49_4 := dataGetArticleBlockList0_49.Result[4]
	assert.Equal(types.StatusAlive, articleBlock0_49_4.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_49_4.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_49_4.CommentType)
	assert.Equal([][]byte{commentBytes0_43}, articleBlock0_49_4.Buf)

	articleBlock0_49_5 := dataGetArticleBlockList0_49.Result[5]
	assert.Equal(types.StatusDeleted, articleBlock0_49_5.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_49_5.ContentType)
	assert.Equal(content.CommentTypeBoo, articleBlock0_49_5.CommentType)
	assert.Equal(content.DefaultDeletedComment, articleBlock0_49_5.Buf)

	articleBlock0_49_6 := dataGetArticleBlockList0_49.Result[6]
	assert.Equal(types.StatusAlive, articleBlock0_49_6.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_49_6.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_49_6.CommentType)
	assert.Equal([][]byte{commentBytes0_45}, articleBlock0_49_6.Buf)

	// 50. content-delete-comment
	marshaledID, _ = board0_10_0.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = articleBlock0_46_6.RefID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_deleteComment", "params": ["%v", "%v", "%v"]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))
	t.Logf("45. content_createComment: bodyString: %v", bodyString)
	dataDeleteComment0_50 := &content.BackendDeleteComment{}
	testCore(t0, bodyString, dataDeleteComment0_50, t, isDebug)

	// 51. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_51 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_51, t, isDebug)
	assert.Equal(7, len(dataGetArticleBlockList0_51.Result))
	articleBlock0_51_4 := dataGetArticleBlockList0_51.Result[4]
	assert.Equal(types.StatusAlive, articleBlock0_51_4.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_51_4.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_51_4.CommentType)
	assert.Equal([][]byte{commentBytes0_43}, articleBlock0_51_4.Buf)

	articleBlock0_51_5 := dataGetArticleBlockList0_51.Result[5]
	assert.Equal(types.StatusDeleted, articleBlock0_51_5.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_51_5.ContentType)
	assert.Equal(content.CommentTypeBoo, articleBlock0_51_5.CommentType)
	assert.Equal(content.DefaultDeletedComment, articleBlock0_51_5.Buf)

	articleBlock0_51_6 := dataGetArticleBlockList0_51.Result[6]
	assert.Equal(types.StatusDeleted, articleBlock0_51_6.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_51_6.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_51_6.CommentType)
	assert.Equal(content.DefaultDeletedComment, articleBlock0_51_6.Buf)
}
