// Copyright 2018 The go-pttai Authors
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

// Copyright 2015 The go-ethereum Authors
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

package utils

import (
	"strings"

	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/me"
	"github.com/ailabstw/go-pttai/metrics"
	"github.com/ailabstw/go-pttai/node"
	"gopkg.in/urfave/cli.v1"
)

// These are all the command line flags we support.
// If you add to this list, please remember to include the
// flag in the appropriate command definition.
//
// The flags are defined here so their names and help texts
// are the same for all commands.

var (
	// General settings
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: DirectoryString{node.DefaultDataDir()},
	}
	KeyStoreDirFlag = DirectoryFlag{
		Name:  "keystore",
		Usage: "Directory for the keystore (default = inside the datadir)",
	}
	TestP2PFlag = cli.BoolFlag{
		Name:  "testp2p",
		Usage: "Test p2p-network: pre-configured PTT.ai test p2p-network",
	}
	IPFSP2PFlag = cli.BoolFlag{
		Name:  "ipfsp2p",
		Usage: "IPFS network: pre-configured PTT.ai ipfs p2p-network",
	}
	DevP2PFlag = cli.BoolFlag{
		Name:  "devp2p",
		Usage: "Dev p2p-network: pre-configured PTT.ai dev p2p-network",
	}

	E2EFlag = cli.BoolFlag{
		Name:  "e2e",
		Usage: "e2e environment",
	}

	PrivateAsPublicFlag = cli.BoolFlag{
		Name:  "private-as-public",
		Usage: "Private api as public api",
	}

	OffsetSecondFlag = cli.Int64Flag{
		Name:  "offset-second",
		Usage: "offset second",
	}

	IdentityFlag = cli.StringFlag{
		Name:  "username",
		Usage: "Custom user name",
	}

	LogFilenameFlag = cli.StringFlag{
		Name:  "log",
		Usage: "log filename",
	}

	// My settings
	MyDataDirFlag = DirectoryFlag{
		Name:  "mydatadir",
		Usage: "Data directory for my info",
		Value: DirectoryString{me.DefaultConfig.DataDir},
	}
	MyKeyFileFlag = cli.StringFlag{
		Name:  "mykey",
		Usage: "my key file",
	}
	MyKeyHexFlag = cli.StringFlag{
		Name:  "mykeyhex",
		Usage: "my key as hex (for testing)",
	}

	MyPostfixFlag = cli.StringFlag{
		Name:  "mypostfix",
		Usage: "my postfix (20 bytes)",
	}

	ServerFlag = cli.BoolFlag{
		Name:  "server",
		Usage: "set as server mode",
	}

	// service settings
	ServiceExpireOplogSecondsFlag = cli.IntFlag{
		Name:  "serviceexpireoplog",
		Usage: "expire oplog seconds",
	}

	// Content settings
	ContentDataDirFlag = DirectoryFlag{
		Name:  "contentdatadir",
		Usage: "Data directory for content",
		Value: DirectoryString{content.DefaultConfig.DataDir},
	}

	ContentKeystoreDirFlag = DirectoryFlag{
		Name:  "contentkeystoredir",
		Usage: "Keystore directory for content",
		Value: DirectoryString{content.DefaultConfig.KeystoreDir},
	}

	// Friend settings
	FriendMaxSyncRandomSecondsFlag = cli.IntFlag{
		Name:  "friendmaxsync",
		Usage: "max sync random seconds",
	}

	FriendMinSyncRandomSecondsFlag = cli.IntFlag{
		Name:  "friendminsync",
		Usage: "min sync random seconds",
	}

	// Performance tuning settings
	CacheFlag = cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching",
		Value: 1024,
	}
	CacheDatabaseFlag = cli.IntFlag{
		Name:  "cache.database",
		Usage: "Percentage of cache memory allowance to use for database io",
		Value: 75,
	}
	CacheGCFlag = cli.IntFlag{
		Name:  "cache.gc",
		Usage: "Percentage of cache memory allowance to use for trie pruning",
		Value: 25,
	}

	// Logging and debug settings
	PttStatsURLFlag = cli.StringFlag{
		Name:  "pttstats",
		Usage: "Reporting URL of a pttstats service (nodename:secret@host:port)",
	}
	MetricsEnabledFlag = cli.BoolFlag{
		Name:  metrics.MetricsEnabledFlag,
		Usage: "Enable metrics collection and reporting",
	}
	MetricsEnableInfluxDBFlag = cli.BoolFlag{
		Name:  "metrics.influxdb",
		Usage: "Enable metrics export/push to an external InfluxDB database",
	}
	MetricsInfluxDBEndpointFlag = cli.StringFlag{
		Name:  "metrics.influxdb.endpoint",
		Usage: "InfluxDB API endpoint to report metrics to",
		Value: "http://localhost:8086",
	}
	MetricsInfluxDBDatabaseFlag = cli.StringFlag{
		Name:  "metrics.influxdb.database",
		Usage: "InfluxDB database name to push reported metrics to",
		Value: "gptt",
	}
	MetricsInfluxDBUsernameFlag = cli.StringFlag{
		Name:  "metrics.influxdb.username",
		Usage: "Username to authorize access to the database",
		Value: "test",
	}
	MetricsInfluxDBPasswordFlag = cli.StringFlag{
		Name:  "metrics.influxdb.password",
		Usage: "Password to authorize access to the database",
		Value: "test",
	}
	// The `host` tag is part of every measurement sent to InfluxDB. Queries on tags are faster in InfluxDB.
	// It is used so that we can group all nodes and average a measurement across all of them, but also so
	// that we can select a specific node and inspect its measurements.
	// https://docs.influxdata.com/influxdb/v1.4/concepts/key_concepts/#tag-key
	MetricsInfluxDBHostTagFlag = cli.StringFlag{
		Name:  "metrics.influxdb.host.tag",
		Usage: "InfluxDB `host` tag attached to all measurements",
		Value: "localhost",
	}

	// HTTP server
	HTTPAddrFlag = cli.StringFlag{
		Name:  "httpaddr",
		Usage: "HTTP server listening addr",
	}
	HTTPDirFlag = cli.StringFlag{
		Name:  "httpdir",
		Usage: "HTTP server serving file-dir",
	}
	ExternHTTPAddrFlag = cli.StringFlag{
		Name:  "exthttpaddr",
		Usage: "External HTTP server listening addr",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolTFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: node.DefaultHTTPHost,
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: node.DefaultHTTPPort,
	}
	ExternRPCAddrFlag = cli.StringFlag{
		Name:  "extrpcaddr",
		Usage: "External HTTP-RPC server listening addr",
	}
	RPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced)",
		Value: "",
	}
	RPCVirtualHostsFlag = cli.StringFlag{
		Name:  "rpcvhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}
	RPCApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface",
		Value: "",
	}
	IPCDisabledFlag = cli.BoolFlag{
		Name:  "ipcdisable",
		Usage: "Disable the IPC-RPC server",
	}
	IPCPathFlag = DirectoryFlag{
		Name:  "ipcpath",
		Usage: "Filename for IPC socket/pipe within the datadir (explicit paths escape it)",
	}
	WSEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable the WS-RPC server",
	}
	WSListenAddrFlag = cli.StringFlag{
		Name:  "wsaddr",
		Usage: "WS-RPC server listening interface",
		Value: node.DefaultWSHost,
	}
	WSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port",
		Value: node.DefaultWSPort,
	}
	WSApiFlag = cli.StringFlag{
		Name:  "wsapi",
		Usage: "API's offered over the WS-RPC interface",
		Value: "",
	}
	WSAllowedOriginsFlag = cli.StringFlag{
		Name:  "wsorigins",
		Usage: "Origins from which to accept websockets requests",
		Value: "",
	}
	ExecFlag = cli.StringFlag{
		Name:  "exec",
		Usage: "Execute JavaScript statement",
	}
	PreloadJSFlag = cli.StringFlag{
		Name:  "preload",
		Usage: "Comma separated list of JavaScript files to preload into the console",
	}

	// Network Settings
	MaxPeersFlag = cli.IntFlag{
		Name:  "maxpeers",
		Usage: "Maximum number of network peers (network disabled if set to 0)",
		Value: 25,
	}
	MaxPendingPeersFlag = cli.IntFlag{
		Name:  "maxpendpeers",
		Usage: "Maximum number of pending connection attempts (defaults used if set to 0)",
		Value: 0,
	}
	ListenPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Network listening port",
		Value: 29487,
	}
	BootnodesFlag = cli.StringFlag{
		Name:  "bootnodes",
		Usage: "Comma separated enode URLs for P2P discovery bootstrap (set v4+v5 instead for light servers)",
		Value: "",
	}
	BootnodesV4Flag = cli.StringFlag{
		Name:  "bootnodesv4",
		Usage: "Comma separated enode URLs for P2P v4 discovery bootstrap (light server, full nodes)",
		Value: "",
	}
	BootnodesV5Flag = cli.StringFlag{
		Name:  "bootnodesv5",
		Usage: "Comma separated enode URLs for P2P v5 discovery bootstrap (light server, light nodes)",
		Value: "",
	}
	P2PListenPortFlag = cli.IntFlag{
		Name:  "p2pport",
		Usage: "Network listening port",
		Value: 9487,
	}
	P2PBootnodesFlag = cli.StringFlag{
		Name:  "p2pbootnodes",
		Usage: "Comma separated enode URLs for libp2p bootstrap",
		Value: "",
	}
	NodeKeyFileFlag = cli.StringFlag{
		Name:  "nodekey",
		Usage: "P2P node key file",
	}
	NodeKeyHexFlag = cli.StringFlag{
		Name:  "nodekeyhex",
		Usage: "P2P node key as hex (for testing)",
	}
	NATFlag = cli.StringFlag{
		Name:  "nat",
		Usage: "NAT port mapping mechanism (any|none|upnp|pmp|extip:<IP>)",
		Value: "any",
	}
	NoDiscoverFlag = cli.BoolFlag{
		Name:  "nodiscover",
		Usage: "Disables the peer discovery mechanism (manual peer addition)",
	}
	DiscoveryV5Flag = cli.BoolFlag{
		Name:  "v5disc",
		Usage: "Enables the experimental RLPx V5 (Topic Discovery) mechanism",
	}
	NetrestrictFlag = cli.StringFlag{
		Name:  "netrestrict",
		Usage: "Restricts network communication to the given IP networks (CIDR masks)",
	}
)
