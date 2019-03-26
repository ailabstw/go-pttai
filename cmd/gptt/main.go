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

// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// gptt is the official command-line client for Pttai.
package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	godebug "runtime/debug"
	"sort"
	"strconv"
	"time"

	"github.com/ailabstw/go-pttai/cmd/utils"
	"github.com/ailabstw/go-pttai/internal/debug"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/metrics"
	"github.com/elastic/gosigar"
	cli "gopkg.in/urfave/cli.v1"
)

const (
	clientIdentifier = "gptt" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit  = ""
	theVersion = ""
)

func InitApp() *cli.App {
	println("InitApp: start", "theVersion", theVersion, "gitCommit", gitCommit)
	// The app that holds all commands and flags.
	app := utils.NewApp(gitCommit, theVersion, "the go-pttai command line interface")

	// Initialize the CLI app and start Geth
	app.Action = gptt
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2018-2018 The go-pttai Authors"
	app.Commands = []cli.Command{
		versionCommand,
		licenseCommand,
		dumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, meFlags...)
	app.Flags = append(app.Flags, contentFlags...)
	app.Flags = append(app.Flags, friendFlags...)
	app.Flags = append(app.Flags, serviceFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, httpFlags...)
	app.Flags = append(app.Flags, networkFlags...)
	app.Flags = append(app.Flags, debug.Flags...)

	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		if err := debug.Setup(ctx); err != nil {
			return err
		}
		// Cap the cache allowance and tune the garbage colelctor
		var mem gosigar.Mem
		if err := mem.Get(); err == nil {
			allowance := int(mem.Total / 1024 / 1024 / 3)
			if cache := ctx.GlobalInt(utils.CacheFlag.Name); cache > allowance {
				log.Warn("Sanitizing cache to Go's GC limits", "provided", cache, "updated", allowance)
				ctx.GlobalSet(utils.CacheFlag.Name, strconv.Itoa(allowance))
			}
		}
		// Ensure Go's GC ignores the database cache for trigger percentage
		cache := ctx.GlobalInt(utils.CacheFlag.Name)
		gogc := math.Max(20, math.Min(100, 100/(float64(cache)/1024)))

		log.Debug("Sanitizing Go's GC trigger", "percent", int(gogc))
		godebug.SetGCPercent(int(gogc))

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}

	return app
}

func main() {
    app := InitApp()
    err := app.Run(os.Args)

    if err != nil {
        fmt.Fprintf(os.Stderr, "something went wrong with app: cmd: %v e: %v\n", os.Args, err)
        os.Exit(1)
    }
}
