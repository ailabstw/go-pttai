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
	"fmt"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestContentTitle(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaledID []byte
	var marshaled []byte
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

	// 3. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_3 := &me.MyInfo{}
	testCore(t0, bodyString, me0_3, t, isDebug)
	assert.Equal(types.StatusAlive, me0_3.Status)
	assert.Equal(me0_1.ID, me0_3.ID)
	assert.Equal(1, len(me0_3.OwnerIDs))
	assert.Equal(me0_3.ID, me0_3.OwnerIDs[0])
	assert.Equal(true, me0_3.IsOwner(me0_3.ID))

	// 13. get board list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoardList", "params": ["", 0, 2]}`)

	dataBoardList0_13 := &struct {
		Result []*content.BackendGetBoard `json:"result"`
	}{}

	testListCore(t0, bodyString, dataBoardList0_13, t, isDebug)
	assert.Equal(1, len(dataBoardList0_13.Result))
	board0_13_0 := dataBoardList0_13.Result[0]
	assert.Equal(me0_3.BoardID, board0_13_0.ID)
	assert.Equal(types.StatusAlive, board0_13_0.Status)

	defaultTitle0_13_0 := content.DefaultTitleTW(me0_1.ID, me0_1.ID, "")
	assert.Equal(defaultTitle0_13_0, board0_13_0.Title)

	// 14. set title
	marshaled, _ = board0_13_0.ID.MarshalText()

	title0_14 := []byte("個版標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_14)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_setTitle", "params": ["%v", "%v"]}`, string(marshaled), marshaledStr)

	dataSetTitle0_14 := &content.BackendGetBoard{}
	testCore(t0, bodyString, dataSetTitle0_14, t, isDebug)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 15. get title
	marshaled, _ = board0_13_0.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawTitle", "params": ["%v"]}`, string(marshaled))

	dataGetTitle0_15 := &content.Title{}
	testCore(t0, bodyString, dataGetTitle0_15, t, isDebug)

	assert.Equal(title0_14, dataGetTitle0_15.Title)

	dataGetTitle1_15 := &content.Title{}
	testCore(t0, bodyString, dataGetTitle1_15, t, isDebug)

	assert.Equal(title0_14, dataGetTitle1_15.Title)

	// 13. create-board
	title := []byte("非個板標題1")
	marshaledStr = base64.StdEncoding.EncodeToString(title)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createBoard", "params": ["%v", true]}`, marshaledStr)

	dataCreateBoard0_13 := &content.BackendCreateBoard{}

	testCore(t0, bodyString, dataCreateBoard0_13, t, isDebug)
	assert.Equal(pkgservice.EntityTypePrivate, dataCreateBoard0_13.BoardType)
	assert.Equal(title, dataCreateBoard0_13.Title)
	assert.Equal(types.StatusAlive, dataCreateBoard0_13.Status)
	assert.Equal(me0_1.ID, dataCreateBoard0_13.CreatorID)
	assert.Equal(me0_1.ID, dataCreateBoard0_13.UpdaterID)

	// 13.1. get-raw-board
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board0_13_1 := content.NewEmptyBoard()
	testCore(t0, bodyString, board0_13_1, t, isDebug)
	assert.Equal(title, board0_13_1.Title)

	// 13.2. get-board
	marshaledID, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getBoard", "params": ["%v"]}`, string(marshaledID))

	board0_13_2 := &content.BackendGetBoard{}
	testCore(t0, bodyString, board0_13_2, t, isDebug)
	assert.Equal(title, board0_13_2.Title)

	// 16. set title
	marshaled, _ = dataCreateBoard0_13.ID.MarshalText()

	title0_16 := []byte("非個板標題2")
	marshaledStr = base64.StdEncoding.EncodeToString(title0_16)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_setTitle", "params": ["%v", "%v"]}`, string(marshaled), marshaledStr)

	dataSetTitle0_16 := &content.BackendGetBoard{}
	testCore(t0, bodyString, dataSetTitle0_16, t, isDebug)

	//assert.Equal(defaultTitle, data0_11.Title)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 17. get title
	marshaled, _ = dataCreateBoard0_13.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawTitle", "params": ["%v"]}`, string(marshaled))

	dataGetTitle0_17 := &content.Title{}
	testCore(t0, bodyString, dataGetTitle0_17, t, isDebug)

	assert.Equal(title0_16, dataGetTitle0_17.Title)

	dataGetTitle1_17 := &content.Title{}
	testCore(t0, bodyString, dataGetTitle1_17, t, isDebug)

	assert.Equal(title0_16, dataGetTitle1_17.Title)
}
