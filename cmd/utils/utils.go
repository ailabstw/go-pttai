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

package utils

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	golog "log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/fdlimit"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/internal/debug"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	"github.com/ailabstw/go-pttai/metrics"
	"github.com/ailabstw/go-pttai/metrics/influxdb"
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/p2p/nat"
	"github.com/ailabstw/go-pttai/p2p/netutil"
	"github.com/ailabstw/go-pttai/params"
	"github.com/ailabstw/go-pttai/raft"
	pkgservice "github.com/ailabstw/go-pttai/service"
	logging "github.com/whyrusleeping/go-logging"
	cli "gopkg.in/urfave/cli.v1"
)

func SetLogging(ctx *cli.Context) {
	logging.SetLevel(logging.DEBUG, "swarm2")

	if ctx.GlobalIsSet(debug.VerbosityFlag.Name) {
		log.LogLevel = log.Lvl(ctx.GlobalInt(debug.VerbosityFlag.Name))
	}

	if ctx.GlobalIsSet(LogFilenameFlag.Name) {
		log.LogFilename = ctx.GlobalString(LogFilenameFlag.Name)
		logging.SetLevel(logging.CRITICAL, "swarm2")
		raft.SetLogger(&raft.DefaultLogger{Logger: golog.New(ioutil.Discard, "", 0)})
	}

}

// SetContentConfig applies node-related command line flags to the config.
func SetUtilsConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(HTTPDirFlag.Name):
		cfg.HTTPDir = ctx.GlobalString(HTTPDirFlag.Name)
	}

	switch {
	case ctx.GlobalIsSet(HTTPAddrFlag.Name):
		cfg.HTTPAddr = ctx.GlobalString(HTTPAddrFlag.Name)
	}

	switch {
	case ctx.GlobalIsSet(ExternHTTPAddrFlag.Name):
		cfg.ExternHTTPAddr = ctx.GlobalString(ExternHTTPAddrFlag.Name)
	default:
		cfg.ExternHTTPAddr = "http://" + cfg.HTTPAddr
	}

}

// SetNodeConfig applies node-related command line flags to the config.
func SetNodeConfig(ctx *cli.Context, cfg *node.Config) {
	log.Debug("SetNodeConfig: start")
	setP2PConfig(ctx, &cfg.P2P)
	setIPC(ctx, cfg)
	setHTTP(ctx, cfg)
	setWS(ctx, cfg)

	// data-dir
	switch {
	case ctx.GlobalIsSet(DataDirFlag.Name):
		cfg.DataDir = ctx.GlobalString(DataDirFlag.Name)
	}

	if ctx.GlobalIsSet(KeyStoreDirFlag.Name) {
		cfg.KeyStoreDir = ctx.GlobalString(KeyStoreDirFlag.Name)
	}
}

// SetMyConfig applies node-related command line flags to the config.
func SetMeConfig(ctx *cli.Context, cfg *me.Config, cfgNode *node.Config) {
	// data-dir
	log.Debug("SetMeConfig: to set DataDir", "cfgNode.DataDIR", cfgNode.DataDir)
	cfg.DataDir = filepath.Join(cfgNode.DataDir, "me")

	// key/id/postfix
	setMyKey(ctx, cfg)
}

// SetMyKey creates a node key from set command line flags, either loading it
// from a file or as a specified hex value. If neither flags were provided, this
// method returns nil and an emphemeral key is to be generated.
func setMyKey(ctx *cli.Context, cfg *me.Config) error {
	var (
		hex     = ctx.GlobalString(MyKeyHexFlag.Name)
		file    = ctx.GlobalString(MyKeyFileFlag.Name)
		postfix = ctx.GlobalString(MyPostfixFlag.Name)
	)

	err := cfg.SetMyKey(hex, file, postfix, false)
	if err != nil {
		return err
	}

	return nil
}

// SetContentConfig applies node-related command line flags to the config.
func SetAccountConfig(ctx *cli.Context, cfg *account.Config, cfgNode *node.Config) {
	// datadir
	log.Debug("SetAccountConfig: to set DataDir", "cfgNode.DataDIR", cfgNode.DataDir)
	cfg.DataDir = filepath.Join(cfgNode.DataDir, "account")
}

// SetContentConfig applies node-related command line flags to the config.
func SetContentConfig(ctx *cli.Context, cfg *content.Config, cfgNode *node.Config) {
	// datadir
	log.Debug("SetContentConfig: to set DataDir", "cfgNode.DataDIR", cfgNode.DataDir)
	cfg.DataDir = filepath.Join(cfgNode.DataDir, "content")

	switch {
	case ctx.GlobalIsSet(ContentKeystoreDirFlag.Name):
		cfg.KeystoreDir = ctx.GlobalString(ContentKeystoreDirFlag.Name)
	default:
		cfg.KeystoreDir = filepath.Join(cfgNode.DataDir, ".keystore")
	}
}

// SetContentConfig applies node-related command line flags to the config.
func SetFriendConfig(ctx *cli.Context, cfg *friend.Config, cfgNode *node.Config) {
	// datadir
	log.Debug("SetFriendConfig: to set DataDir", "cfgNode.DataDIR", cfgNode.DataDir)
	cfg.DataDir = filepath.Join(cfgNode.DataDir, "friend")

	if ctx.GlobalIsSet(FriendMaxSyncRandomSecondsFlag.Name) {
		cfg.MaxSyncRandomSeconds = ctx.GlobalInt(FriendMaxSyncRandomSecondsFlag.Name)
	}

	if ctx.GlobalIsSet(FriendMinSyncRandomSecondsFlag.Name) {
		cfg.MinSyncRandomSeconds = ctx.GlobalInt(FriendMinSyncRandomSecondsFlag.Name)
	}

	friend.MaxSyncRandomSeconds = cfg.MaxSyncRandomSeconds
	friend.MinSyncRandomSeconds = cfg.MinSyncRandomSeconds
}

// SetPttConfig applies ptt-related command line flags to the config.
func SetPttConfig(ctx *cli.Context, cfg *pkgservice.Config, cfgNode *node.Config, gitCommit string, version string) {
	log.Debug("SetPttConfig: start", "cfg", cfg, "cfgNode", cfgNode, "cfgNode.DataDir", cfgNode.DataDir, "params.Version", params.Version)

	// data-dir
	cfg.DataDir = filepath.Join(cfgNode.DataDir, "ptt")

	cfg.Version = version
	cfg.GitCommit = gitCommit

	// node-type
	switch {
	case ctx.GlobalBool(ServerFlag.Name):
		cfg.NodeType = pkgservice.NodeTypeServer
	default:
		cfg.NodeType = pkgservice.NodeTypeDesktop
	}

	// expire oplog seconds
	if ctx.GlobalIsSet(ServiceExpireOplogSecondsFlag.Name) {
		cfg.ExpireOplogSeconds = ctx.GlobalInt(ServiceExpireOplogSecondsFlag.Name)
	}

	pkgservice.ExpireOplogSeconds = cfg.ExpireOplogSeconds

	log.Debug("SetPttConfig: to return", "ExpireOplogSeconds", pkgservice.ExpireOplogSeconds)

}

// MakeDataDir retrieves the currently requested data directory, terminating
// if none (or the empty string) is specified. If the node is starting a testnet,
// the a subdirectory of the specified datadir will be used.
func MakeDataDir(ctx *cli.Context) string {
	if path := ctx.GlobalString(DataDirFlag.Name); path != "" {
		return path
	}
	Fatalf("Cannot determine default data directory, please set manually (--datadir)")
	return ""
}

// setNodeKey creates a node key from set command line flags, either loading it
// from a file or as a specified hex value. If neither flags were provided, this
// method returns nil and an emphemeral key is to be generated.
func setNodeKey(ctx *cli.Context, cfg *p2p.Config) {
	var (
		hex  = ctx.GlobalString(NodeKeyHexFlag.Name)
		file = ctx.GlobalString(NodeKeyFileFlag.Name)
		key  *ecdsa.PrivateKey
		err  error
	)
	switch {
	case file != "" && hex != "":
		Fatalf("Options %q and %q are mutually exclusive", NodeKeyFileFlag.Name, NodeKeyHexFlag.Name)
	case file != "":
		if key, err = crypto.LoadECDSA(file); err != nil {
			Fatalf("Option %q: %v", NodeKeyFileFlag.Name, err)
		}
		cfg.PrivateKey = key
	case hex != "":
		if key, err = crypto.HexToECDSA(hex); err != nil {
			Fatalf("Option %q: %v", NodeKeyHexFlag.Name, err)
		}
		cfg.PrivateKey = key
	}
}

// setBootstrapNodes creates a list of bootstrap nodes from the command line
// flags, reverting to pre-configured ones if none have been specified.
func setBootstrapNodes(ctx *cli.Context, cfg *p2p.Config) {
	var urls []string
	switch {
	case ctx.GlobalIsSet(BootnodesFlag.Name) || ctx.GlobalIsSet(BootnodesV4Flag.Name):
		if ctx.GlobalIsSet(BootnodesV4Flag.Name) {
			urls = strings.Split(ctx.GlobalString(BootnodesV4Flag.Name), ",")
		} else {
			urls = strings.Split(ctx.GlobalString(BootnodesFlag.Name), ",")
		}
	case cfg.BootstrapNodes != nil:
		return // already set, don't apply defaults.
	default:
		urls = params.MainnetBootnodes
	}

	cfg.BootstrapNodes = make([]*discover.Node, 0, len(urls))
	for _, url := range urls {
		node, err := discover.ParseNode(url)
		if err != nil {
			log.Error("Bootstrap URL invalid", "enode", url, "err", err)
			continue
		}
		cfg.BootstrapNodes = append(cfg.BootstrapNodes, node)
	}
}

func setP2PBootnodes(ctx *cli.Context, cfg *p2p.Config) {
	var urls []string
	switch {
	case ctx.GlobalIsSet(P2PBootnodesFlag.Name):
		urls = strings.Split(ctx.GlobalString(P2PBootnodesFlag.Name), ",")
	case ctx.GlobalBool(TestP2PFlag.Name):
		urls = params.TestP2PBootnodes
	case ctx.GlobalBool(DevP2PFlag.Name):
		urls = params.DevP2PBootnodes
	case ctx.GlobalBool(IPFSP2PFlag.Name):
		urls = params.IPFSBootnodes
	case cfg.P2PBootnodes != nil:
		return // already set, don't apply defaults.
	default:
		urls = params.MainP2PBootnodes
	}

	cfg.P2PBootnodes = make([]*discover.Node, 0, len(urls))
	for _, url := range urls {
		node, err := discover.ParseP2PNode(url)
		if err != nil {
			log.Error("Bootstrap P2P URL invalid", "pnode", url, "e", err)
			continue
		}
		cfg.P2PBootnodes = append(cfg.P2PBootnodes, node)
	}
	log.Info("setP2PBootnodes: done", "P2PBootnodes", len(cfg.P2PBootnodes))
}

// setListenAddress creates a TCP listening address string from set command
// line flags.
func setListenAddress(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(ListenPortFlag.Name) {
		cfg.ListenAddr = fmt.Sprintf(":%d", ctx.GlobalInt(ListenPortFlag.Name))
	}
}

// setListenAddress creates a TCP listening address string from set command
// line flags.
func setP2PListenAddress(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(P2PListenPortFlag.Name) {
		cfg.P2PListenAddr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", ctx.GlobalInt(P2PListenPortFlag.Name))
	}
}

// setNAT creates a port mapper from command line flags.
func setNAT(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(NATFlag.Name) {
		natif, err := nat.Parse(ctx.GlobalString(NATFlag.Name))
		if err != nil {
			Fatalf("Option %s: %v", NATFlag.Name, err)
		}
		cfg.NAT = natif
	}
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setHTTP(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(RPCEnabledFlag.Name) && cfg.HTTPHost == "" {
		cfg.HTTPHost = "127.0.0.1"
		if ctx.GlobalIsSet(RPCListenAddrFlag.Name) {
			cfg.HTTPHost = ctx.GlobalString(RPCListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(RPCPortFlag.Name) {
		cfg.HTTPPort = ctx.GlobalInt(RPCPortFlag.Name)
	}
	if ctx.GlobalIsSet(RPCCORSDomainFlag.Name) {
		cfg.HTTPCors = splitAndTrim(ctx.GlobalString(RPCCORSDomainFlag.Name))
	}
	if ctx.GlobalIsSet(RPCApiFlag.Name) {
		cfg.HTTPModules = splitAndTrim(ctx.GlobalString(RPCApiFlag.Name))
	}
	if ctx.GlobalIsSet(RPCVirtualHostsFlag.Name) {
		cfg.HTTPVirtualHosts = splitAndTrim(ctx.GlobalString(RPCVirtualHostsFlag.Name))
	}

	if ctx.GlobalIsSet(ExternRPCAddrFlag.Name) {
		cfg.ExternHTTPAddr = ctx.GlobalString(ExternRPCAddrFlag.Name)
	} else {
		cfg.ExternHTTPAddr = "http://" + cfg.HTTPHost + ":" + strconv.Itoa(cfg.HTTPPort)
	}
}

// setWS creates the WebSocket RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setWS(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(WSEnabledFlag.Name) && cfg.WSHost == "" {
		cfg.WSHost = "127.0.0.1"
		if ctx.GlobalIsSet(WSListenAddrFlag.Name) {
			cfg.WSHost = ctx.GlobalString(WSListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(WSPortFlag.Name) {
		cfg.WSPort = ctx.GlobalInt(WSPortFlag.Name)
	}
	if ctx.GlobalIsSet(WSAllowedOriginsFlag.Name) {
		cfg.WSOrigins = splitAndTrim(ctx.GlobalString(WSAllowedOriginsFlag.Name))
	}
	if ctx.GlobalIsSet(WSApiFlag.Name) {
		cfg.WSModules = splitAndTrim(ctx.GlobalString(WSApiFlag.Name))
	}
}

// setIPC creates an IPC path configuration from the set command line flags,
// returning an empty string if IPC was explicitly disabled, or the set path.
func setIPC(ctx *cli.Context, cfg *node.Config) {
	checkExclusive(ctx, IPCDisabledFlag, IPCPathFlag)
	switch {
	case ctx.GlobalBool(IPCDisabledFlag.Name):
		cfg.IPCPath = ""
	case ctx.GlobalIsSet(IPCPathFlag.Name):
		cfg.IPCPath = ctx.GlobalString(IPCPathFlag.Name)
	}
}

// makeDatabaseHandles raises out the number of allowed file handles per process
// for Geth and returns half of the allowance to assign to the database.
func makeDatabaseHandles() int {
	limit, err := fdlimit.Current()
	if err != nil {
		Fatalf("Failed to retrieve file descriptor allowance: %v", err)
	}
	if limit < 2048 {
		if err := fdlimit.Raise(2048); err != nil {
			Fatalf("Failed to raise file descriptor allowance: %v", err)
		}
	}
	if limit > 2048 { // cap database file descriptors even if more is available
		limit = 2048
	}
	return limit / 2 // Leave half for networking and other stuff
}

func setP2PConfig(ctx *cli.Context, cfg *p2p.Config) {
	setNodeKey(ctx, cfg)
	setNAT(ctx, cfg)
	setListenAddress(ctx, cfg)
	setBootstrapNodes(ctx, cfg)
	setP2PListenAddress(ctx, cfg)
	setP2PBootnodes(ctx, cfg)

	if ctx.GlobalIsSet(MaxPeersFlag.Name) {
		cfg.MaxPeers = ctx.GlobalInt(MaxPeersFlag.Name)
	}

	pttPeers := cfg.MaxPeers

	log.Info("Maximum peer count", "PTT", pttPeers, "total", cfg.MaxPeers)

	if ctx.GlobalIsSet(MaxPendingPeersFlag.Name) {
		cfg.MaxPendingPeers = ctx.GlobalInt(MaxPendingPeersFlag.Name)
	}
	if ctx.GlobalIsSet(NoDiscoverFlag.Name) {
		cfg.NoDiscovery = true
	}

	// if we're running a light client or server, force enable the v5 peer discovery
	// unless it is explicitly disabled with --nodiscover note that explicitly specifying
	// --v5disc overrides --nodiscover, in which case the later only disables v4 discovery
	// forceV5Discovery := (lightClient || lightServer) && !ctx.GlobalBool(NoDiscoverFlag.Name)

	forceV5Discovery := false
	if ctx.GlobalIsSet(DiscoveryV5Flag.Name) {
		cfg.DiscoveryV5 = ctx.GlobalBool(DiscoveryV5Flag.Name)
	} else if forceV5Discovery {
		cfg.DiscoveryV5 = true
	}

	if netrestrict := ctx.GlobalString(NetrestrictFlag.Name); netrestrict != "" {
		list, err := netutil.ParseNetlist(netrestrict)
		if err != nil {
			Fatalf("Option %q: %v", NetrestrictFlag.Name, err)
		}
		cfg.NetRestrict = list
	}

}

// checkExclusive verifies that only a single isntance of the provided flags was
// set by the user. Each flag might optionally be followed by a string type to
// specialize it further.
func checkExclusive(ctx *cli.Context, args ...interface{}) {
	set := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		// Make sure the next argument is a flag and skip if not set
		flag, ok := args[i].(cli.Flag)
		if !ok {
			panic(fmt.Sprintf("invalid argument, not cli.Flag type: %T", args[i]))
		}
		// Check if next arg extends current and expand its name if so
		name := flag.GetName()

		if i+1 < len(args) {
			switch option := args[i+1].(type) {
			case string:
				// Extended flag, expand the name and shift the arguments
				if ctx.GlobalString(flag.GetName()) == option {
					name += "=" + option
				}
				i++

			case cli.Flag:
			default:
				panic(fmt.Sprintf("invalid argument, not cli.Flag or string extension: %T", args[i+1]))
			}
		}
		// Mark the flag if it's set
		if ctx.GlobalIsSet(flag.GetName()) {
			set = append(set, "--"+name)
		}
	}
	if len(set) > 1 {
		Fatalf("Flags %v can't be used at the same time", strings.Join(set, ", "))
	}
}

// MigrateFlags sets the global flag from a local flag when it's set.
// This is a temporary function used for migrating old command/flags to the
// new format.
//
// e.g. geth account new --keystore /tmp/mykeystore --lightkdf
//
// is equivalent after calling this method with:
//
// geth --keystore /tmp/mykeystore --lightkdf account new
//
// This allows the use of the existing configuration functionality.
// When all flags are migrated this function can be removed and the existing
// configuration functionality must be changed that is uses local flags
func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}

func SetupMetrics(ctx *cli.Context) {
	if metrics.Enabled {
		log.Info("Enabling metrics collection")
		var (
			enableExport = ctx.GlobalBool(MetricsEnableInfluxDBFlag.Name)
			endpoint     = ctx.GlobalString(MetricsInfluxDBEndpointFlag.Name)
			database     = ctx.GlobalString(MetricsInfluxDBDatabaseFlag.Name)
			username     = ctx.GlobalString(MetricsInfluxDBUsernameFlag.Name)
			password     = ctx.GlobalString(MetricsInfluxDBPasswordFlag.Name)
			hosttag      = ctx.GlobalString(MetricsInfluxDBHostTagFlag.Name)
		)

		if enableExport {
			log.Info("Enabling metrics export to InfluxDB")
			go influxdb.InfluxDBWithTags(metrics.DefaultRegistry, 10*time.Second, endpoint, database, username, password, "gptt.", map[string]string{
				"host": hosttag,
			})
		}
	}
}
