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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

const ()

var (
	ctx       context.Context    = nil
	cancel    context.CancelFunc = nil
	tBootnode *exec.Cmd          = nil

	ctxs    []context.Context    = nil
	cancels []context.CancelFunc = nil
	nodes   []*exec.Cmd          = nil
	stderrs []io.ReadCloser      = nil

	NNodes         = 3
	TimeoutSeconds = 120 * time.Second
)

type RBody struct {
	Body []byte
}

func GetResponseBody(r *RBody) func(res *http.Response, req *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		body, err := readBody(res)
		if err != nil {
			return err
		}
		r.Body = body
		return nil
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
	os.MkdirAll("./test.out", 0755)

	ctx, cancel = context.WithTimeout(context.Background(), TimeoutSeconds)

	tBootnode = exec.CommandContext(ctx, "../build/bin/bootnode", "--nodekeyhex", "03f509202abd40be562951247c7fe05294bb71ccad54f4853f2d75e3bf94affd", "--addr", "127.0.0.1:9488")
	err := tBootnode.Start()
	if err != nil {
		t.Errorf("unable to start tBootnode, e: %v", err)
	}

	ctxs = make([]context.Context, NNodes)
	cancels = make([]context.CancelFunc, NNodes)
	nodes = make([]*exec.Cmd, NNodes)
	stderrs = make([]io.ReadCloser, NNodes)

	for i := 0; i < NNodes; i++ {
		dir := fmt.Sprintf("./test.out/.test%d", i)
		rpcport := fmt.Sprintf("%d", 9450+i)
		port := fmt.Sprintf("%d", 9500+i)
		httpaddr := fmt.Sprintf("127.0.0.1:%d", 9600+i)

		ctxs[i], cancels[i] = context.WithTimeout(context.Background(), TimeoutSeconds)
		nodes[i] = exec.CommandContext(ctxs[i], "../build/bin/gptt", "--datadir", dir, "--rpcaddr", "127.0.0.1", "--rpcport", rpcport, "--port", port, "--httpaddr", httpaddr, "--bootnodes", "pnode://03f509202abd40be562951247c7fe05294bb71ccad54f4853f2d75e3bf94affd@127.0.0.1:9488", "--ipcdisable")
		stderrs[i], _ = nodes[i].StderrPipe()
		err := nodes[i].Start()
		if err != nil {
			t.Errorf("unable to start node: i: %v e: %v", i, err)
		}
	}

	t.Logf("wait 3 seconds for node starting")
	time.Sleep(3 * time.Second)
}

func teardownTest(t *testing.T) {
	cancel()
	for i := 0; i < NNodes; i++ {
		cancels[i]()

		/*
		   err, _ := ioutil.ReadAll(stderrs[i])
		   t.Logf("after teardownTest: (%v/%v) e: %v", i, NNodes, string(err))
		*/
	}

	os.RemoveAll("./test.out")

}
