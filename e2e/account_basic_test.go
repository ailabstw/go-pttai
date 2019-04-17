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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestAccountBasic(t *testing.T) {
	NNodes = 1
	isDebug := true

	var bodyString string
	var marshaledID []byte
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
	assert.Equal(me0_3.ProfileID[common.AddressLength:], me0_3.ID[:common.AddressLength])

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

	// 10. profile
	marshaledID, _ = me0_3.ProfileID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaledID))

	profile0_10 := &account.Profile{}

	testCore(t0, bodyString, profile0_10, t, isDebug)
	assert.Equal(profile0_10.ID[common.AddressLength:], me0_3.ID[:common.AddressLength])
	assert.Equal(profile0_10.ID, me0_3.ProfileID)
	assert.Equal(profile0_10.MyID, me0_3.ID)
	assert.Equal(types.StatusAlive, profile0_10.Status)

	// 11. masters from profile
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterListFromCache", "params": ["%v"]}`, string(marshaledID))

	dataMasterList0_11 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11, t, isDebug)
	assert.Equal(1, len(dataMasterList0_11.Result))
	master0_11_0 := dataMasterList0_11.Result[0]

	// 11.1
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMasterList0_11_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11_1, t, isDebug)
	assert.Equal(1, len(dataMasterList0_11_1.Result))
	master0_11_1_0 := dataMasterList0_11_1.Result[0]

	assert.Equal(master0_11_0, master0_11_1_0)
	assert.Equal(types.StatusAlive, master0_11_0.Status)

	// 12. members from profile
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaledID))

	dataMemberList0_12 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberList0_12, t, isDebug)
	assert.Equal(1, len(dataMemberList0_12.Result))

	// 13. op-key from profile

	// 14. set user name
	myName := base64.StdEncoding.EncodeToString([]byte("老蕭"))
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_setMyName", "params": ["%v"]}`, myName)

	dataSetMyName0_14 := &account.UserName{}
	testCore(t0, bodyString, dataSetMyName0_14, t, isDebug)
	assert.Equal(me0_1.ID, dataSetMyName0_14.ID)

	// 15. get user name
	marshaledID, _ = me0_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawUserName", "params": ["%v"]}`, string(marshaledID))

	dataGetUserName0_15 := &account.UserName{}
	testCore(t0, bodyString, dataGetUserName0_15, t, isDebug)
	assert.Equal(me0_1.ID, dataGetUserName0_15.ID)
	assert.Equal([]byte("老蕭"), dataGetUserName0_15.Name)

	// 16. my_setMyImg
	normalizedImgStr := "data:image/png;base64,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"

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_setMyImage", "params": ["%v"]}`, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==")
	t.Logf("setMyImage: bodyString: %v", bodyString)
	dataSetMyImg0_16 := &account.UserImg{}
	testCore(t0, bodyString, dataSetMyImg0_16, t, false)
	assert.Equal(me0_1.ID, dataSetMyImg0_16.ID)

	// 17. get user img
	marshaledID, _ = me0_1.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawUserImg", "params": ["%v"]}`, string(marshaledID))

	dataGetUserImg0_17 := &account.UserImg{}
	testCore(t0, bodyString, dataGetUserImg0_17, t, isDebug)
	assert.Equal(me0_1.ID, dataGetUserImg0_17.ID)
	assert.Equal(account.ImgTypePNG, dataGetUserImg0_17.ImgType)
	assert.Equal(uint16(account.MaxProfileImgWidth), dataGetUserImg0_17.Width)
	assert.Equal(uint16(account.MaxProfileImgHeight), dataGetUserImg0_17.Height)
	assert.Equal(normalizedImgStr, dataGetUserImg0_17.Str)
	assert.Equal(types.StatusAlive, dataGetUserImg0_17.Status)

	// 18. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 2, 2]}`)

	dataBoardList0_18 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_18, t, isDebug)
	assert.Equal(1, len(dataBoardList0_18.Result))
	board0_18_0 := dataBoardList0_18.Result[0]
	assert.Equal(me0_3.BoardID, board0_18_0.ID)
	assert.Equal(pkgservice.EntityTypePersonal, board0_18_0.BoardType)
	assert.Equal(types.StatusAlive, board0_18_0.Status)
}
