// Copyright 2019 The go-pttai Authors
// This file is part of go-pttai.
//
// go-pttai is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-pttai is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-pttai. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/cmd/utils"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/internal/debug"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/ptthttp"
	pkgservice "github.com/ailabstw/go-pttai/service"
	cli "gopkg.in/urfave/cli.v1"
)

func gptt(ctx *cli.Context) error {
	utils.SetLogging(ctx)

	log.Info("PTT.ai: Hello world!")

	// Load Config
	cfg, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	utils.SetUtilsConfig(ctx, cfg.Utils)

	// we need NodeConfig be the 1st. The DataDir in other configs are referring to the DataDir in NodeConfig.
	utils.SetNodeConfig(ctx, cfg.Node)

	utils.SetMeConfig(ctx, cfg.Me, cfg.Node)

	utils.SetAccountConfig(ctx, cfg.Account, cfg.Node)

	utils.SetContentConfig(ctx, cfg.Content, cfg.Node)

	utils.SetFriendConfig(ctx, cfg.Friend, cfg.Node)

	utils.SetPttConfig(ctx, cfg.Ptt, cfg.Node, gitCommit, theVersion)

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

	// http-server

	httpServer, err := ptthttp.NewServer(cfg.Utils.HTTPDir, cfg.Utils.HTTPAddr, cfg.Node.HTTPPort, cfg.Utils.ExternHTTPAddr, cfg.Node.ExternHTTPAddr, n)
	if err != nil {
		return err
	}

	httpServer.Start()

	// open-browser
	if !ctx.GlobalIsSet(utils.ServerFlag.Name) && !ctx.GlobalIsSet(utils.ExternRPCAddrFlag.Name) && !ctx.GlobalIsSet(utils.ExternHTTPAddrFlag.Name) {
		go func() {
			time.Sleep(TimeSleepBrowser * time.Second)
			utils.OpenBrowser(cfg.Utils.HTTPAddr)
		}()
	}

	// set-signal
	go setSignal(n, httpServer)

	// wait-node
	if err := WaitNode(n, httpServer); err != nil {
		return err
	}

	log.Info("PTT.ai: see u later～")

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

	ptt, err := pkgservice.NewPtt(ctx, cfg.Ptt, &myNodeID, myNodeKey)
	if err != nil {
		return nil, err
	}

	accountBackend, err := account.NewBackend(ctx, cfg.Account, ptt)
	if err != nil {
		return nil, err
	}
	err = ptt.RegisterService(accountBackend)
	if err != nil {
		return nil, err
	}

	// content
	contentBackend, err := content.NewBackend(ctx, cfg.Content, cfg.Me.ID, ptt, accountBackend)
	if err != nil {
		return nil, err
	}
	err = ptt.RegisterService(contentBackend)
	if err != nil {
		return nil, err
	}

	// friend
	friendBackend, err := friend.NewBackend(ctx, cfg.Friend, cfg.Me.ID, ptt, accountBackend, contentBackend)
	if err != nil {
		return nil, err
	}
	err = ptt.RegisterService(friendBackend)
	if err != nil {
		return nil, err
	}

	// me
	meBackend, err := me.NewBackend(ctx, cfg.Me, ptt, accountBackend, contentBackend, friendBackend)
	if err != nil {
		return nil, err
	}

	err = ptt.RegisterService(meBackend)
	if err != nil {
		return nil, err
	}

	err = ptt.Prestart()
	if err != nil {
		log.Error("unable to do Prestart", "e", err)
		return nil, err
	}

	return ptt, nil
}

func setSignal(n *node.Node, server *ptthttp.Server) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	<-sigc

	log.Debug("setSignal: received break-signal")
	go func() {
		server.Stop()
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

func WaitNode(n *node.Node, server *ptthttp.Server) error {
	log.Info("start Waiting...")

	ptt := n.Services()[reflect.TypeOf(&pkgservice.BasePtt{})].(*pkgservice.BasePtt)

loop:
	for {
		select {
		case _, ok := <-ptt.NotifyNodeRestart().GetChan():
			log.Debug("WaitNode: NotifyNodeRestart: start")
			if !ok {
				break loop
			}
			server.Stop()
			err := n.Restart(false, true)
			if err != nil {
				return err
			}
			server.SetRPCServer(n)
			server.Start()
			ptt = n.Services()[reflect.TypeOf(&pkgservice.BasePtt{})].(*pkgservice.BasePtt)
			log.Debug("WaitNode: NotifyNodeRestart: done")
		case _, ok := <-ptt.NotifyNodeStop().GetChan():
			log.Debug("WaitNode: NotifyNodeStop: start")
			if !ok {
				break loop
			}
			server.Stop()
			n.Stop(false, false)
			log.Debug("WaitNode: NotifyNodeStop: done")
			break loop
		case err, ok := <-ptt.ErrChan().GetChan():
			if !ok {
				break loop
			}
			log.Error("Received err from ptt", "e", err)
			break loop
		case err, ok := <-n.StopChan:
			log.Debug("WaitNode: StopChan: start")
			if ok && err != nil {
				log.Error("Wait", "e", err)
				return err
			}
			log.Debug("WaitNode: StopChan: done")
			break loop
		}
	}

	return nil
}
