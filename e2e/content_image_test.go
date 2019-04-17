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
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-etherum/common"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestContentImage(t *testing.T) {
	NNodes = 1
	isDebug := true

	var bodyString string
	var marshaledID []byte
	var marshaledID2 []byte
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
	assert.Equal(me0_3.BoardID[common.AddressLength:], me0_3.ID[:common.AddressLength])

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
	bodyString = `{"id": "testID", "method": "me_getOpKeyOplogList", "params": ["", "", 0, 2]}`

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

	// 10. board
	marshaledID, _ = me0_3.BoardID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board0_10 := &content.Board{}

	testCore(t0, bodyString, board0_10, t, isDebug)
	assert.Equal(board0_10.ID[common.AddressLength:], me0_3.ID[:common.AddressLength])
	assert.Equal(board0_10.ID, me0_3.BoardID)
	assert.Equal(board0_10.CreatorID, me0_3.ID)
	assert.Equal(types.StatusAlive, board0_10.Status)
	assert.Equal(pkgservice.EntityTypePersonal, board0_10.EntityType)

	// 11. masters from board
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterListFromCache", "params": ["%v"]}`, string(marshaledID))

	dataMasterList0_11 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11, t, isDebug)
	assert.Equal(1, len(dataMasterList0_11.Result))
	master0_11_0 := dataMasterList0_11.Result[0]

	// 11.1
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterList0_11_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11_1, t, isDebug)
	assert.Equal(1, len(dataMasterList0_11_1.Result))
	master0_11_1_0 := dataMasterList0_11_1.Result[0]

	assert.Equal(master0_11_0, master0_11_1_0)
	assert.Equal(types.StatusAlive, master0_11_0.Status)

	// 12. members from board
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberList0_12 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberList0_12, t, isDebug)
	assert.Equal(1, len(dataMemberList0_12.Result))

	// 13. upload image
	img0_13, _ := ioutil.ReadFile("./btn_confirm.png")
	marshaledStr = base64.StdEncoding.EncodeToString(img0_13)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_uploadImage", "params": ["%v", "image/png", "%v"]}`, string(marshaledID), marshaledStr)

	dataUploadImg0_13 := &content.BackendUploadImg{}
	testCore(t0, bodyString, dataUploadImg0_13, t, isDebug)
	assert.Equal(pkgservice.MediaTypeJPEG, dataUploadImg0_13.Type)

	// 14. get image
	marshaledID2, _ = dataUploadImg0_13.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getImage", "params": ["%v", "%v"]}`, string(marshaledID), string(marshaledID2))

	dataGetImage0_14 := &content.BackendGetImg{}
	testCore(t0, bodyString, dataGetImage0_14, t, isDebug)

	outImg, _ := base64.StdEncoding.DecodeString(`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`)

	assert.Equal(outImg, dataGetImage0_14.Buf)
	assert.Equal(pkgservice.MediaTypeJPEG, dataGetImage0_14.Type)
	assert.Equal(dataUploadImg0_13.ID, dataGetImage0_14.ID)
}
