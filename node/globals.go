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

package node

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/nat"
)

const (
	_ uint32 = iota
	Mainnet
	Testnet
	Devnet
)

const (
	DataDirPrivateKey      = "nodekey"            // Path within the datadir to the node's private key
	DataDirDefaultKeyStore = "keystore"           // Path within the datadir to the keystore
	DataDirStaticNodes     = "static-nodes.json"  // Path within the datadir to the static node list
	DataDirTrustedNodes    = "trusted-nodes.json" // Path within the datadir to the trusted node list
	DataDirNodeDatabase    = "nodes"              // Path within the datadir to store the node infos

	DefaultHTTPHost = ""    // Default host interface for the HTTP RPC server
	DefaultHTTPPort = 14779 // Default TCP port for the HTTP RPC server
	DefaultWSHost   = ""    // Default host interface for the websocket RPC server
	DefaultWSPort   = 15779 // Default TCP port for the websocket RPC server

	DefaultNetworkID = Devnet
)

var (
	DefaultConfig = Config{
		DataDir:          DefaultDataDir(),
		IPCPath:          "gptt.ipc",
		HTTPHost:         DefaultHTTPHost,
		HTTPPort:         DefaultHTTPPort,
		HTTPCors:         []string{"localhost"},
		HTTPVirtualHosts: []string{"localhost"},
		HTTPModules:      []string{"debug", "net", "admin", "ptt", "account", "content", "me", "friend"},
		WSPort:           DefaultWSPort,
		P2P: p2p.Config{
			ListenAddr: ":9487",
			MaxPeers:   350,
			NAT:        nat.Any(),
		},
		NetworkID: DefaultNetworkID,
	}
)

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
// used by cmd.utils
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".pttai")
			//return filepath.Join(home, "Library", "Pttai")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "Pttai")
		} else {
			return filepath.Join(home, ".pttai")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
