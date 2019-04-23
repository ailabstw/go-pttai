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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
	signalserver "github.com/ailabstw/pttai-signal-server"
	"github.com/gorilla/mux"
	baloo "gopkg.in/h2non/baloo.v3"
)

const ()

var (
	ctx    context.Context    = nil
	cancel context.CancelFunc = nil

	Ctxs    []context.Context    = nil
	Cancels []context.CancelFunc = nil
	Nodes   []*exec.Cmd          = nil
	stderrs []*os.File           = nil

	NNodes         = 5
	TimeoutSeconds = 240 * time.Second

	origHandler log.Handler

	nilPttID           *types.PttID
	nilSyncInfo        *pkgservice.BaseSyncInfo
	nilSyncArticleInfo *content.SyncArticleInfo
	nilProfile         *account.Profile

	TimeSleepRestart = 30 * time.Second
	TimeSleepDefault = 15 * time.Second

	ServiceExpireOplog = "300"
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

func GetResponseBody(r *RBody, t *testing.T) func(res *http.Response, req *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		body, err := readBody(res, t)
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
		t.Logf("unable to parse: b: %v e: %v", b, err)
	}
}

func readBody(res *http.Response, t *testing.T) ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Logf("[ERROR] Unable to read body: e: %v", err)
		return []byte{}, err
	}
	// Re-fill body reader stream after reading it
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, err
}

func startNode(t *testing.T, idx int, offsetSecond int64, isNewLog bool) {
	if Ctxs[idx] != nil && Cancels[idx] != nil {
		Cancels[idx]()
		Ctxs[idx] = nil
		Cancels[idx] = nil
	}

	dir := fmt.Sprintf("./test.out/.test%d", idx)
	rpcport := fmt.Sprintf("%d", 9450+idx)
	p2pport := fmt.Sprintf("%d", 9500+idx)
	port := fmt.Sprintf("%d", 9600+idx)
	httpaddr := fmt.Sprintf("127.0.0.1:%d", 9700+idx)

	Ctxs[idx], Cancels[idx] = context.WithTimeout(context.Background(), TimeoutSeconds)
	Nodes[idx] = exec.CommandContext(
		Ctxs[idx],
		"../build/bin/gptt",
		"--exthttpaddr", "http://localhost:9776",
		"--verbosity", "4",
		"--datadir", dir,
		"--rpcaddr", "127.0.0.1",
		"--httpaddr", httpaddr,
		"--rpcport", rpcport,
		"--port", port,
		"--p2pport", p2pport,
		"--webrtcsignalserver", "127.0.0.1:9489",
		"--ipcdisable",
		"--friendmaxsync", "7",
		"--friendminsync", "5",
		"--serviceexpireoplog", ServiceExpireOplog,
		"--offset-second", strconv.FormatInt(offsetSecond, 10),
		"--e2e",
	)
	filename := fmt.Sprintf("./test.out/log.err.%d.txt", idx)
	var err error
	if stderrs[idx] != nil {
		stderrs[idx].Close()
	}
	if isNewLog {
		stderrs[idx], _ = os.Create(filename)
	} else {
		stderrs[idx], err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			t.Errorf("unable to open log-filename: filename: %v e: %v", filename, err)
		}
	}
	Nodes[idx].Stderr = stderrs[idx]
	err = Nodes[idx].Start()
	if err != nil {
		t.Errorf("unable to start node: i: %v e: %v", idx, err)
	}

	time.Sleep(5 * time.Second)

}

func setupTest(t *testing.T) {
	content.InitLocaleInfo()

	os.RemoveAll("./test.out")

	os.MkdirAll("./test.out", 0755)

	ctx, cancel = context.WithTimeout(context.Background(), TimeoutSeconds)

	addr := "127.0.0.1:9489"
	go func() {
		server := signalserver.NewServer()

		srv := &http.Server{Addr: addr}
		r := mux.NewRouter()
		r.HandleFunc("/signal", server.SignalHandler)
		srv.Handler = r

		srv.ListenAndServe()
	}()

	origHandler = log.Root().GetHandler()
	log.Root().SetHandler(log.Must.FileHandler("./test.out/log.tmp.txt", log.TerminalFormat(true)))

	Ctxs = make([]context.Context, NNodes)
	Cancels = make([]context.CancelFunc, NNodes)
	Nodes = make([]*exec.Cmd, NNodes)
	stderrs = make([]*os.File, NNodes)

	for i := 0; i < NNodes; i++ {
		startNode(t, i, 0, true)
	}

	seconds := 0
	switch {
	case NNodes <= 3:
		seconds = 5
	case NNodes == 4:
		seconds = 5
	case NNodes == 5:
		seconds = 5
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
		AssertFunc(GetResponseBody(rbody, t)).
		Done()

	ParseBody(rbody.Body, t, data, true)

	if isDebug {
		t.Logf("after Parse: length: %v header: %v body: %v data: %v", rbody.ContentLength, rbody.Header, string(rbody.Body), data)
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
		AssertFunc(GetResponseBody(rbody, t)).
		Done()

	var dataWrapper *DataWrapper
	err := &MyError{}
	if data != nil {
		dataWrapper = &DataWrapper{Result: data, Error: err}
		ParseBody(rbody.Body, t, dataWrapper, false)
	}

	if isDebug {
		if data != nil {
			t.Logf("after Parse: body: %v data: %v", string(rbody.Body), dataWrapper.Result)
		} else {
			t.Logf("after Parse: body: %v", string(rbody.Body))

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
		AssertFunc(GetResponseBody(rbody, t)).
		Done()

	ParseBody(rbody.Body, t, dataWrapper, false)
	if isDebug {
		t.Logf("after Parse: length: %v header: %v body: %v data: %v", rbody.ContentLength, rbody.Header, string(rbody.Body), dataWrapper.Result)
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
		AssertFunc(GetResponseBody(rbody, t)).
		Done()

	ParseBody(rbody.Body, t, dataWrapper, false)
	if isDebug {
		t.Logf("after Parse: length: %v header: %v body: %v data: %v", rbody.ContentLength, rbody.Header, string(rbody.Body), dataWrapper.Result)
	}

	return dataWrapper.Result, rbody.Body
}

func testBodyEqualCore(c *baloo.Client, bodyString string, resultString string, t *testing.T) []byte {

	rbody := &RBody{}

	c.Post("/").
		BodyString(bodyString).
		SetHeader("Content-Type", "application/json").
		Expect(t).
		AssertFunc(GetResponseBody(rbody, t)).
		BodyEquals(resultString).
		Done()

	return rbody.Body
}
