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
	"fmt"
	"reflect"
	"unicode"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/cmd/utils"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/me"
	"github.com/ailabstw/go-pttai/node"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/naoina/toml"
	cli "gopkg.in/urfave/cli.v1"
)

const ()

// config
var (
	DefaultConfig = Config{
		Node:    &node.DefaultConfig,
		Me:      &me.DefaultConfig,
		Content: &content.DefaultConfig,
		Account: &account.DefaultConfig,
		Friend:  &friend.DefaultConfig,
		Ptt:     &pkgservice.DefaultConfig,
		Utils:   &utils.DefaultConfig,
	}
)

// flags
var (
	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	// flags that configure me
	meFlags = []cli.Flag{
		utils.MyDataDirFlag,
		utils.MyKeyFileFlag,
		utils.MyKeyHexFlag,
		utils.ServerFlag,
	}

	// flags that configure content
	contentFlags = []cli.Flag{
		utils.ContentDataDirFlag,
		utils.ContentKeystoreDirFlag,
	}

	// flags that configure http-server
	httpFlags = []cli.Flag{
		utils.HTTPAddrFlag,
		utils.HTTPDirFlag,
		utils.ExternHTTPAddrFlag,
	}

	// flags that configure the node
	nodeFlags = []cli.Flag{
		configFileFlag,

		utils.DataDirFlag,
		utils.KeyStoreDirFlag,

		utils.TestnetFlag,
		utils.IdentityFlag,

		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheGCFlag,

		utils.PttStatsURLFlag,
		utils.MetricsEnabledFlag,
		utils.MetricsEnableInfluxDBFlag,
		utils.MetricsInfluxDBEndpointFlag,
		utils.MetricsInfluxDBDatabaseFlag,
		utils.MetricsInfluxDBUsernameFlag,
		utils.MetricsInfluxDBPasswordFlag,
		utils.MetricsInfluxDBHostTagFlag,
	}

	// flags that configure p2p-network
	networkFlags = []cli.Flag{
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.ListenPortFlag,

		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,

		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,

		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
	}

	// flags that configure rpc
	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.ExternRPCPortFlag,

		utils.RPCApiFlag,

		utils.IPCDisabledFlag,
		utils.IPCPathFlag,

		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
	}
)

// cmd
var (
	versionCommand = cli.Command{
		Action:    utils.MigrateFlags(version),
		Name:      "version",
		Usage:     "Print version numbers",
		ArgsUsage: " ",
		Category:  "MISCELLANEOUS COMMANDS",
		Description: `
The output of this command is supposed to be machine-readable.
`,
	}

	licenseCommand = cli.Command{
		Action:    utils.MigrateFlags(license),
		Name:      "license",
		Usage:     "Display license information",
		ArgsUsage: " ",
		Category:  "MISCELLANEOUS COMMANDS",
	}

	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(append(append(append(nodeFlags, meFlags...), contentFlags...), rpcFlags...), httpFlags...), networkFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
)

// toml-settings
var (
	tomlSettings = toml.Config{
		NormFieldName: func(rt reflect.Type, key string) string {
			return key
		},
		FieldToKey: func(rt reflect.Type, field string) string {
			return field
		},
		MissingField: func(rt reflect.Type, field string) error {
			link := ""
			if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
				link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
			}
			return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
		},
	}
)
