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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ailabstw/go-pttai/cmd/utils"
	"github.com/ailabstw/go-pttai/internal/debug"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/p2p/discover"
	pkgservice "github.com/ailabstw/go-pttai/service"
	cli "gopkg.in/urfave/cli.v1"
)

func gptt(ctx *cli.Context) error {
	log.Info("Ptt.ai: Hello world!")

	// Load Config
	cfg, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	utils.SetNodeConfig(ctx, cfg.Node)

	utils.SetPttConfig(ctx, cfg.Ptt, cfg.Node, gitCommit)

	// Setup metrics
	utils.SetupMetrics(ctx)

	// new node
	n, err := node.New(cfg.Node)
	if err != nil {
		return err
	}

	// register ptt
	if err := registerPtt(n, cfg); err != nil {
		return err
	}

	// node start
	if err := n.Start(); err != nil {
		return err
	}

	// set-signal
	go setSignal(n)

	// wait-node
	if err := WaitNode(n); err != nil {
		return err
	}

	log.Info("Ptt.ai: see u later～")

	return nil
}

func registerPtt(n *node.Node, cfg *Config) error {
	return n.Register(func(ctx *pkgservice.ServiceContext) (pkgservice.PttService, error) {
		return registerServices(ctx, cfg)
	})
}

func registerServices(ctx *pkgservice.ServiceContext, cfg *Config) (pkgservice.PttService, error) {
	myNodeKey := cfg.Node.NodeKey()
	myNodeID := discover.PubkeyID(&myNodeKey.PublicKey)

	ptt, err := pkgservice.NewPtt(ctx, cfg.Ptt, &myNodeID)
	if err != nil {
		return nil, err
	}

	return ptt, nil
}

func setSignal(n *node.Node) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	<-sigc

	log.Debug("setSignal: received break-signal")
	go func() {
		n.Stop(false, false)
	}()

	for i := 10; i > 0; i-- {
		<-sigc
		if i > 1 {
			log.Warn("Already shutting down, interrupt more to panic.", "times", i-1)
		}
	}
	debug.Exit()
	debug.LoudPanic("boom")
}

func WaitNode(n *node.Node) error {
	log.Info("start Waiting...")

loop:
	for {
		err, ok := <-n.StopChan
		if ok && err != nil {
			log.Error("Wait", "e", err)
			return err
		}
		break loop
	}

	return nil
}
