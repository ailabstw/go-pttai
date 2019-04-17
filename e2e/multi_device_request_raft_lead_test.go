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
	"fmt"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceRequestRaftLead(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
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

	// 7. join-me
	log.Debug("7. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey0_4)

	dataJoinMe0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_7.NodeID)

	// wait 15
	t.Logf("wait 15 seconds for hand-shaking")
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

	// 9. getRawMeByID
	marshaled, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaled))

	me0_9 := &me.MyInfo{}
	testCore(t0, bodyString, me0_9, t, isDebug)
	assert.Equal(types.StatusMigrated, me0_9.Status)
	assert.Equal(2, len(me0_9.OwnerIDs))
	assert.Equal(true, me0_9.IsOwner(me1_3.ID))
	assert.Equal(true, me0_9.IsOwner(me0_3.ID))

	// 10. raft-satus
	marshaled, _ = me1_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRaftStatus", "params": ["%v"]}`, string(marshaled))

	raftStatus0_10 := &me.RaftStatus{}
	testCore(t0, bodyString, raftStatus0_10, t, isDebug)

	assert.Equal(2, len(raftStatus0_10.ConfState.Nodes))

	raftStatus1_10 := &me.RaftStatus{}
	testCore(t1, bodyString, raftStatus1_10, t, isDebug)

	assert.Equal(2, len(raftStatus1_10.ConfState.Nodes))
	assert.Equal(raftStatus0_10.Lead, raftStatus1_10.Lead)

	// 11. request-raft
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_requestRaftLead", "params": []}`)

	requestRaftLead1_10 := false
	testCore(t1, bodyString, &requestRaftLead1_10, t, isDebug)

	time.Sleep(5 * time.Second)

	// 12. raft-satus
	marshaled, _ = me1_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRaftStatus", "params": ["%v"]}`, string(marshaled))

	raftStatus0_12 := &me.RaftStatus{}
	testCore(t0, bodyString, raftStatus0_12, t, isDebug)

	assert.Equal(2, len(raftStatus0_12.ConfState.Nodes))

	raftStatus1_12 := &me.RaftStatus{}
	testCore(t1, bodyString, raftStatus1_12, t, isDebug)

	assert.Equal(2, len(raftStatus1_12.ConfState.Nodes))
	assert.Equal(raftStatus0_12.Lead, raftStatus1_12.Lead)
	assert.Equal(raftStatus0_12.Lead, me1_1.RaftID)

	// 13. request-raft
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_requestRaftLead", "params": []}`)

	requestRaftLead0_10 := false
	testCore(t0, bodyString, &requestRaftLead0_10, t, isDebug)

	time.Sleep(5 * time.Second)

	// 14. raft-satus
	marshaled, _ = me1_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRaftStatus", "params": ["%v"]}`, string(marshaled))

	raftStatus0_14 := &me.RaftStatus{}
	testCore(t0, bodyString, raftStatus0_14, t, isDebug)

	assert.Equal(2, len(raftStatus0_14.ConfState.Nodes))

	raftStatus1_14 := &me.RaftStatus{}
	testCore(t1, bodyString, raftStatus1_14, t, isDebug)

	assert.Equal(2, len(raftStatus1_14.ConfState.Nodes))
	assert.Equal(raftStatus0_14.Lead, raftStatus1_14.Lead)
	assert.Equal(raftStatus0_14.Lead, me0_1.RaftID)

	// 15. request-raft
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_requestRaftLead", "params": []}`)

	requestRaftLead1_15 := false
	testCore(t1, bodyString, &requestRaftLead1_15, t, isDebug)

	time.Sleep(5 * time.Second)

	// 16. raft-satus
	marshaled, _ = me1_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRaftStatus", "params": ["%v"]}`, string(marshaled))

	raftStatus0_16 := &me.RaftStatus{}
	testCore(t0, bodyString, raftStatus0_16, t, isDebug)

	assert.Equal(2, len(raftStatus0_16.ConfState.Nodes))

	raftStatus1_16 := &me.RaftStatus{}
	testCore(t1, bodyString, raftStatus1_16, t, isDebug)

	assert.Equal(2, len(raftStatus1_16.ConfState.Nodes))
	assert.Equal(raftStatus0_16.Lead, raftStatus1_16.Lead)
	assert.Equal(raftStatus0_16.Lead, me1_1.RaftID)

	// 17. request-raft
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_requestRaftLead", "params": []}`)

	requestRaftLead0_17 := false
	testCore(t0, bodyString, &requestRaftLead0_17, t, isDebug)

	time.Sleep(5 * time.Second)

	// 18. raft-satus
	marshaled, _ = me1_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRaftStatus", "params": ["%v"]}`, string(marshaled))

	raftStatus0_18 := &me.RaftStatus{}
	testCore(t0, bodyString, raftStatus0_18, t, isDebug)

	assert.Equal(2, len(raftStatus0_18.ConfState.Nodes))

	raftStatus1_18 := &me.RaftStatus{}
	testCore(t1, bodyString, raftStatus1_18, t, isDebug)

	assert.Equal(2, len(raftStatus1_18.ConfState.Nodes))
	assert.Equal(raftStatus0_18.Lead, raftStatus1_18.Lead)
	assert.Equal(raftStatus0_18.Lead, me0_1.RaftID)

}
