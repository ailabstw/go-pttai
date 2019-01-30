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
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceSyncBoard2(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
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
	nodeID1_1 := me1_1.NodeID
	pubKey1_1, _ := nodeID1_1.Pubkey()
	nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

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

	// 7. join-me
	log.Debug("7. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey0_4)

	dataJoinMe0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
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

	masterOplog1_9_2 := dataMasterOplogs1_9.Result[2]

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

	// 13. create-board
	title := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createBoard", "params": ["%v", true]}`, marshaledStr)

	dataCreateBoard0_13 := &content.BackendCreateBoard{}

	testCore(t0, bodyString, dataCreateBoard0_13, t, isDebug)
	assert.Equal(pkgservice.EntityTypePrivate, dataCreateBoard0_13.BoardType)
	assert.Equal(title, dataCreateBoard0_13.Title)
	assert.Equal(types.StatusAlive, dataCreateBoard0_13.Status)
	assert.Equal(me1_1.ID, dataCreateBoard0_13.CreatorID)
	assert.Equal(me1_1.ID, dataCreateBoard0_13.UpdaterID)

	// 13.1. board
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board0_13_1 := &content.Board{}

	testCore(t0, bodyString, board0_13_1, t, isDebug)
	assert.Equal(board0_13_1.ID[common.AddressLength:], me1_3.ID[:common.AddressLength])
	assert.Equal(board0_13_1.CreatorID, me1_3.ID)
	assert.Equal(types.StatusAlive, board0_13_1.Status)
	assert.Equal(pkgservice.EntityTypePrivate, board0_13_1.EntityType)

	// 13.2 set title
	marshaled, _ = board0_13_1.ID.MarshalText()

	title0_13_2 := []byte("標題2")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_13_2)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_setTitle", "params": ["%v", "%v"]}`, string(marshaled), marshaledStr)

	dataSetTitle0_13_2 := &content.BackendGetBoard{}
	testCore(t0, bodyString, dataSetTitle0_13_2, t, isDebug)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 13.3 get title
	marshaled, _ = board0_13_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawTitle", "params": ["%v"]}`, string(marshaled))

	dataGetTitle0_13_3 := &content.Title{}
	testCore(t0, bodyString, dataGetTitle0_13_3, t, isDebug)

	assert.Equal(title0_13_2, dataGetTitle0_13_3.Title)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 14. MeOplog
	bodyString = `{"id": "testID", "method": "me_getMeOplogList", "params": ["", 0, 2]}`

	dataMeOplogs0_14 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMeOplogs0_14, t, isDebug)
	assert.Equal(2, len(dataMeOplogs0_14.Result))
	assert.Equal(dataMeOplogs0_9_2.Result, dataMeOplogs0_14.Result[:1])

	meOplog0_14 := dataMeOplogs0_14.Result[1]
	assert.Equal(me1_3.ID, meOplog0_14.CreatorID)
	assert.Equal(dataCreateBoard0_13.ID, meOplog0_14.ObjID)
	assert.Equal(me.MeOpTypeCreateBoard, meOplog0_14.Op)
	assert.Equal(nilPttID, meOplog0_14.PreLogID)
	assert.Equal(types.Bool(true), meOplog0_14.IsSync)
	assert.Equal(masterOplog1_9_2.ID, meOplog0_14.MasterLogID)
	masterSign0_14 := meOplog0_14.MasterSigns[0]
	assert.Equal(nodeAddr0_1[:], masterSign0_14.ID[:common.AddressLength])
	assert.Equal(me1_3.ID[:common.AddressLength], masterSign0_14.ID[common.AddressLength:])

	dataMeOplogs1_14 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMeOplogs1_14, t, isDebug)
	assert.Equal(2, len(dataMeOplogs1_14.Result))
	assert.Equal(dataMeOplogs0_9_2.Result, dataMeOplogs1_14.Result[:1])

	meOplog1_14 := dataMeOplogs1_14.Result[1]
	assert.Equal(me1_3.ID, meOplog1_14.CreatorID)
	assert.Equal(dataCreateBoard0_13.ID, meOplog1_14.ObjID)
	assert.Equal(me.MeOpTypeCreateBoard, meOplog1_14.Op)
	assert.Equal(nilPttID, meOplog1_14.PreLogID)
	assert.Equal(types.Bool(true), meOplog1_14.IsSync)
	assert.Equal(masterOplog1_9_2.ID, meOplog1_14.MasterLogID)
	masterSign1_14 := meOplog1_14.MasterSigns[0]
	assert.Equal(nodeAddr0_1[:], masterSign1_14.ID[:common.AddressLength])
	assert.Equal(me1_3.ID[:common.AddressLength], masterSign1_14.ID[common.AddressLength:])

	// 15. board
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board1_15 := content.NewEmptyBoard()
	testCore(t1, bodyString, board1_15, t, isDebug)

	assert.Equal(board1_15.ID[common.AddressLength:], me1_3.ID[:common.AddressLength])
	assert.Equal(board1_15.CreatorID, me1_3.ID)
	assert.Equal(types.StatusAlive, board1_15.Status)
	assert.Equal(pkgservice.EntityTypePrivate, board1_15.EntityType)

	// 16. master-oplog
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterOplogs0_16 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_16, t, isDebug)
	assert.Equal(1, len(dataMasterOplogs0_16.Result))

	masterOplog0_16 := dataMasterOplogs0_16.Result[0]
	assert.Equal(me1_3.ID, masterOplog0_16.CreatorID)
	assert.Equal(me1_3.ID, masterOplog0_16.ObjID)
	assert.Equal(pkgservice.MasterOpTypeAddMaster, masterOplog0_16.Op)
	assert.Equal(nilPttID, masterOplog0_16.PreLogID)
	assert.Equal(types.Bool(true), masterOplog0_16.IsSync)
	assert.Equal(masterOplog0_16.ID, masterOplog0_16.MasterLogID)
	masterSign0_16 := masterOplog0_16.MasterSigns[0]
	assert.Equal(me1_3.ID, masterSign0_16.ID)

	dataMasterOplogs1_16 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogs1_16, t, isDebug)
	assert.Equal(1, len(dataMasterOplogs1_16.Result))

	masterOplog1_16 := dataMasterOplogs1_16.Result[0]
	assert.Equal(me1_3.ID, masterOplog1_16.CreatorID)
	assert.Equal(me1_3.ID, masterOplog1_16.ObjID)
	assert.Equal(pkgservice.MasterOpTypeAddMaster, masterOplog1_16.Op)
	assert.Equal(nilPttID, masterOplog1_16.PreLogID)
	assert.Equal(types.Bool(true), masterOplog1_16.IsSync)
	assert.Equal(masterOplog1_16.ID, masterOplog1_16.MasterLogID)
	masterSign1_16 := masterOplog1_16.MasterSigns[0]
	assert.Equal(me1_3.ID, masterSign1_16.ID)

	// 16. BoardOplog
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardOplogList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataBoardOplogs0_16_1 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataBoardOplogs0_16_1, t, isDebug)
	assert.Equal(2, len(dataBoardOplogs0_16_1.Result))

	boardOplog0_16_1 := dataBoardOplogs0_16_1.Result[0]
	assert.Equal(me1_3.ID, boardOplog0_16_1.CreatorID)
	assert.Equal(dataCreateBoard0_13.ID, boardOplog0_16_1.ObjID)
	assert.Equal(content.BoardOpTypeCreateBoard, boardOplog0_16_1.Op)
	assert.Equal(nilPttID, boardOplog0_16_1.PreLogID)
	assert.Equal(types.Bool(true), boardOplog0_16_1.IsSync)
	assert.Equal(masterOplog0_16.ID, boardOplog0_16_1.MasterLogID)
	masterSign0_16_1 := boardOplog0_16_1.MasterSigns[0]
	assert.Equal(me1_3.ID, masterSign0_16_1.ID)

	dataBoardOplogs1_16_1 := &struct {
		Result []*content.BoardOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataBoardOplogs1_16_1, t, isDebug)
	assert.Equal(2, len(dataBoardOplogs1_16_1.Result))

	boardOplog1_16_1 := dataBoardOplogs1_16_1.Result[0]
	assert.Equal(me1_3.ID, boardOplog1_16_1.CreatorID)
	assert.Equal(dataCreateBoard0_13.ID, boardOplog1_16_1.ObjID)
	assert.Equal(content.BoardOpTypeCreateBoard, boardOplog1_16_1.Op)
	assert.Equal(nilPttID, boardOplog1_16_1.PreLogID)
	assert.Equal(types.Bool(true), boardOplog1_16_1.IsSync)
	assert.Equal(masterOplog1_16.ID, boardOplog1_16_1.MasterLogID)
	masterSign1_16_1 := boardOplog1_16_1.MasterSigns[0]
	assert.Equal(me1_3.ID, masterSign1_16_1.ID)

	// 42.0. upload file
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	file0_42_0, _ := ioutil.ReadFile("./e2e-test.zip")
	marshaledStr = base64.StdEncoding.EncodeToString(file0_42_0)
	marshaledStr2 = "e2e-test.zip"

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_uploadFile", "params": ["%v", "%v", "%v"]}`, string(marshaledID), marshaledStr2, marshaledStr)

	dataUploadFile0_42_0 := &content.BackendUploadFile{}
	testCore(t0, bodyString, dataUploadFile0_42_0, t, isDebug)

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

	marshaled, _ = dataCreateBoard0_13.ID.MarshalText()

	title0_35 := []byte("標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_35)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaled), marshaledStr, string(article))
	dataCreateArticle0_35 := &content.BackendCreateArticle{}
	testCore(t0, bodyString, dataCreateArticle0_35, t, isDebug)
	assert.Equal(dataCreateBoard0_13.ID, dataCreateArticle0_35.BoardID)
	assert.Equal(3, dataCreateArticle0_35.NBlock)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 36. content-get-article-list
	marshaled, _ = dataCreateBoard0_13.ID.MarshalText()

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

	// 38. get-article-block
	marshaled, _ = dataCreateBoard0_13.ID.MarshalText()
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

	// 49. update-article
	article48, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("測試61")),
		base64.StdEncoding.EncodeToString([]byte("測試62")),
		base64.StdEncoding.EncodeToString([]byte("測試63")),
		base64.StdEncoding.EncodeToString([]byte("測試64")),
		base64.StdEncoding.EncodeToString([]byte("測試65")),
		base64.StdEncoding.EncodeToString([]byte("測試66")),
		base64.StdEncoding.EncodeToString([]byte("測試67")),
		base64.StdEncoding.EncodeToString([]byte("測試68")),
		base64.StdEncoding.EncodeToString([]byte("測試69")),
	})

	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_updateArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), string(marshaledID2), string(article48))
	dataUpdateArticle0_48 := &content.BackendUpdateArticle{}
	testCore(t0, bodyString, dataUpdateArticle0_48, t, isDebug)
	assert.Equal(dataCreateBoard0_13.ID, dataUpdateArticle0_48.BoardID)
	assert.Equal(article0_36.ID, dataUpdateArticle0_48.ArticleID)
	assert.Equal(2, dataUpdateArticle0_48.NBlock)

	// wait 10 seconds
	time.Sleep(10 * time.Second)

	// 49. content-get-article-list
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_49 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_49, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_49.Result))
	article0_49_0 := dataGetArticleList0_49.Result[0]
	assert.Equal(types.StatusAlive, article0_49_0.Status)
	assert.Equal(dataUpdateArticle0_48.ContentBlockID, article0_49_0.ContentBlockID)

	dataGetArticleList1_49 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_49, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_49.Result))
	article1_49_0 := dataGetArticleList1_49.Result[0]
	assert.Equal(types.StatusAlive, article1_49_0.Status)
	assert.Equal(article0_36.ID, article1_49_0.ID)
	assert.Equal(dataUpdateArticle0_48.ContentBlockID, article1_49_0.ContentBlockID)

	// 50. get-article-block
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	marshaledID2, _ = article0_49_0.ID.MarshalText()
	marshaledID3, _ = article0_49_0.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_50 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_50, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList0_50.Result))

	article50_0 := [][]byte{
		[]byte("測試61"),
	}

	article50_1 := [][]byte{
		[]byte("測試62"),
		[]byte("測試63"),
		[]byte("測試64"),
		[]byte("測試65"),
		[]byte("測試66"),
		[]byte("測試67"),
		[]byte("測試68"),
		[]byte("測試69"),
	}

	assert.Equal(article50_0, dataGetArticleBlockList0_50.Result[0].Buf)
	assert.Equal(article50_1, dataGetArticleBlockList0_50.Result[1].Buf)

	dataGetArticleBlockList1_50 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_50, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_50.Result))

	assert.Equal(article50_0, dataGetArticleBlockList1_50.Result[0].Buf)
	assert.Equal(article50_1, dataGetArticleBlockList1_50.Result[1].Buf)

	// 41. content-create-comment
	comment := []byte("這是comment")
	commentStr := base64.StdEncoding.EncodeToString(comment)

	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createComment", "params": ["%v", "%v", 0, "%v", ""]}`, string(marshaledID), string(marshaledID2), commentStr)
	t.Logf("41. content_createComment: bodyString: %v", bodyString)
	dataCreateComment1_41 := &content.BackendCreateComment{}
	testCore(t0, bodyString, dataCreateComment1_41, t, isDebug)
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
	assert.Equal(3, len(dataGetArticleBlockList0_42.Result))
	articleBlock0_42 := dataGetArticleBlockList0_42.Result[2]
	assert.Equal(types.StatusAlive, articleBlock0_42.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_42.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock0_42.CommentType)
	assert.Equal([][]byte{comment}, articleBlock0_42.Buf)

	dataGetArticleBlockList1_42 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_42, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList1_42.Result))
	articleBlock1_42 := dataGetArticleBlockList1_42.Result[2]
	assert.Equal(types.StatusAlive, articleBlock1_42.Status)
	assert.Equal(content.ContentTypeComment, articleBlock1_42.ContentType)
	assert.Equal(content.CommentTypePush, articleBlock1_42.CommentType)
	assert.Equal([][]byte{comment}, articleBlock1_42.Buf)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 48. content-delete-comment
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = articleBlock0_42.RefID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_deleteComment", "params": ["%v", "%v", "%v"]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))
	dataDeleteComment0_48 := &content.BackendDeleteComment{}
	testCore(t0, bodyString, dataDeleteComment0_48, t, isDebug)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 49. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2))

	dataGetArticleBlockList0_49 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_49, t, isDebug)
	assert.Equal(3, len(dataGetArticleBlockList0_49.Result))
	articleBlock0_49 := dataGetArticleBlockList0_49.Result[2]

	assert.Equal(types.StatusDeleted, articleBlock0_49.Status)
	assert.Equal(content.ContentTypeComment, articleBlock0_49.ContentType)
	assert.Equal(content.DefaultDeletedComment, articleBlock0_49.Buf)

	// 50. content-delete-article
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	marshaledID2, _ = article0_36.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_deleteArticle", "params": ["%v", "%v"]}`, string(marshaledID), string(marshaledID2))
	dataDeleteArticle0_50 := &content.BackendDeleteArticle{}
	testCore(t0, bodyString, dataDeleteArticle0_50, t, isDebug)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 51. get-article-block
	marshaledID2, _ = article0_36.ID.MarshalText()
	marshaledID3, _ = article0_36.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList0_51 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleBlockList0_51, t, isDebug)
	assert.Equal(0, len(dataGetArticleBlockList0_51.Result))

	// // 52. content-get-article-list
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_52 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_52, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_52.Result))
	article0_52 := dataGetArticleList0_52.Result[0]

	assert.Equal(0, article0_52.NPush)
	assert.Equal(0, article0_52.NBoo)
	assert.Equal(types.StatusDeleted, article0_52.Status)
}
