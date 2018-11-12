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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	baloo "gopkg.in/h2non/baloo.v3"
)

const ()

var (
	ctx       context.Context    = nil
	cancel    context.CancelFunc = nil
	tBootnode *exec.Cmd          = nil

	Ctxs    []context.Context    = nil
	Cancels []context.CancelFunc = nil
	Nodes   []*exec.Cmd          = nil
	stderrs []*os.File           = nil

	NNodes         = 5
	TimeoutSeconds = 120 * time.Second

	origHandler log.Handler

	nilPttID *types.PttID
)

type RBody struct {
	Header        map[string][]string
	Body          []byte
	ContentLength int64
}

type DataWrapper struct {
	Result interface{} `json:"result"`
	Error  *MyError    `json:"error"`
}

type MyError struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

func GetResponseBody(r *RBody) func(res *http.Response, req *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		body, err := readBody(res)
		if err != nil {
			return err
		}
		r.Body = body
		r.ContentLength = res.ContentLength
		r.Header = res.Header
		return nil
	}
}

func ParseBody(b []byte, t *testing.T, data interface{}, isList bool) {
	err := json.Unmarshal(b, data)
	if err != nil && !isList {
		t.Logf("unable to parse: e: %v", err)
	}
}

func readBody(res *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	// Re-fill body reader stream after reading it
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, err
}

func setupTest(t *testing.T) {
	os.RemoveAll("./test.out")

	os.MkdirAll("./test.out", 0755)

	ctx, cancel = context.WithTimeout(context.Background(), TimeoutSeconds)

	tBootnode = exec.CommandContext(ctx, "../build/bin/bootnode", "--nodekeyhex", "03f509202abd40be562951247c7fe05294bb71ccad54f4853f2d75e3bf94affd", "--addr", "127.0.0.1:9489")
	err := tBootnode.Start()
	if err != nil {
		t.Errorf("unable to start tBootnode, e: %v", err)
	}

	origHandler = log.Root().GetHandler()
	log.Root().SetHandler(log.Must.FileHandler("./test.out/log.tmp.txt", log.TerminalFormat(true)))

	Ctxs = make([]context.Context, NNodes)
	Cancels = make([]context.CancelFunc, NNodes)
	Nodes = make([]*exec.Cmd, NNodes)
	stderrs = make([]*os.File, NNodes)

	for i := 0; i < NNodes; i++ {
		dir := fmt.Sprintf("./test.out/.test%d", i)
		rpcport := fmt.Sprintf("%d", 9450+i)
		port := fmt.Sprintf("%d", 9500+i)

		Ctxs[i], Cancels[i] = context.WithTimeout(context.Background(), TimeoutSeconds)
		Nodes[i] = exec.CommandContext(Ctxs[i], "../build/bin/gptt", "--verbosity", "4", "--datadir", dir, "--rpcaddr", "127.0.0.1", "--rpcport", rpcport, "--port", port, "--bootnodes", "pnode://847e1b261cd827f83a62c6fa6d335179054cecb5651d47b4b152cef67e4b45d7f872e07a2e222771124e0354e58b6b3b1fc8908bb63ec30744abd9784ced31e8@127.0.0.1:9489", "--ipcdisable")
		filename := fmt.Sprintf("./test.out/log.err.%d.txt", i)
		stderrs[i], _ = os.Create(filename)
		Nodes[i].Stderr = stderrs[i]
		err := Nodes[i].Start()
		if err != nil {
			t.Errorf("unable to start node: i: %v e: %v", i, err)
		}
	}

	seconds := 0
	switch {
	case NNodes <= 3:
		seconds = 7
	case NNodes == 4:
		seconds = 8
	case NNodes == 5:
		seconds = 9
	}

	log.Debug("wait for node starting", "seconds", seconds)
	t.Logf("wait %v seconds for node starting", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
}

func teardownTest(t *testing.T) {
	log.Debug("teardownTest: start")

	log.Root().SetHandler(origHandler)

	for i := 0; i < NNodes; i++ {
		Cancels[i]()

		stderrs[i].Close()
	}
	cancel()

	t.Logf("wait 3 seconds for node shutdown")
	time.Sleep(3 * time.Second)

}

func testListCore(c *baloo.Client, bodyString string, data interface{}, t *testing.T, isDebug bool) []byte {
	rbody := &RBody{}

	c.Post("/").
		BodyString(bodyString).
		SetHeader("Content-Type", "application/json").
		Expect(t).
		AssertFunc(GetResponseBody(rbody)).
		Done()

	ParseBody(rbody.Body, t, data, true)

	if isDebug {
		t.Logf("after Parse: length: %v header: %v body: %v data: %v", rbody.ContentLength, rbody.Header, rbody.Body, data)
	}

	return rbody.Body
}

func testError(url string) error {
	bodyString := `{"id": "testID", "method": "ptt_getVersion", "params": []}`
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(bodyString)))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err := client.Do(req)

	return err
}

func testCore(c *baloo.Client, bodyString string, data interface{}, t *testing.T, isDebug bool) ([]byte, *MyError) {
	rbody := &RBody{}

	c.Post("/").
		BodyString(bodyString).
		SetHeader("Content-Type", "application/json").
		Expect(t).
		AssertFunc(GetResponseBody(rbody)).
		Done()

	var dataWrapper *DataWrapper
	err := &MyError{}
	if data != nil {
		dataWrapper = &DataWrapper{Result: data, Error: err}
		ParseBody(rbody.Body, t, dataWrapper, false)
	}

	if isDebug {
		if data != nil {
			t.Logf("after Parse: body: %v data: %v", rbody.Body, dataWrapper.Result)
		} else {
			t.Logf("after Parse: body: %v", rbody.Body)

		}
	}

	return rbody.Body, err
}

func testStringCore(c *baloo.Client, bodyString string, t *testing.T, isDebug bool) (string, []byte) {
	rbody := &RBody{}

	dataWrapper := &struct {
		Result string `json:"result"`
	}{}

	c.Post("/").
		BodyString(bodyString).
		SetHeader("Content-Type", "application/json").
		Expect(t).
		AssertFunc(GetResponseBody(rbody)).
		Done()

	ParseBody(rbody.Body, t, dataWrapper, false)
	if isDebug {
		t.Logf("after Parse: length: %v header: %v body: %v data: %v", rbody.ContentLength, rbody.Header, rbody.Body, dataWrapper.Result)
	}

	return dataWrapper.Result, rbody.Body
}

func testIntCore(c *baloo.Client, bodyString string, t *testing.T, isDebug bool) (int, []byte) {
	rbody := &RBody{}

	dataWrapper := &struct {
		Result int `json:"result"`
	}{}

	c.Post("/").
		BodyString(bodyString).
		SetHeader("Content-Type", "application/json").
		Expect(t).
		AssertFunc(GetResponseBody(rbody)).
		Done()

	ParseBody(rbody.Body, t, dataWrapper, false)
	if isDebug {
		t.Logf("after Parse: length: %v header: %v body: %v data: %v", rbody.ContentLength, rbody.Header, rbody.Body, dataWrapper.Result)
	}

	return dataWrapper.Result, rbody.Body
}

func testBodyEqualCore(c *baloo.Client, bodyString string, resultString string, t *testing.T) []byte {

	rbody := &RBody{}

	c.Post("/").
		BodyString(bodyString).
		SetHeader("Content-Type", "application/json").
		Expect(t).
		AssertFunc(GetResponseBody(rbody)).
		BodyEquals(resultString).
		Done()

	return rbody.Body
}
