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
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"

	"github.com/ailabstw/go-pttai/cmd/utils"
	"github.com/ailabstw/go-pttai/key"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/crypto"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	logging "github.com/whyrusleeping/go-logging"
)

func main() {
	logging.SetLevel(logging.DEBUG, "swarm2")
	logging.SetLevel(logging.DEBUG, "relay")
	logging.SetLevel(logging.DEBUG, "discovery")
	logging.SetLevel(logging.DEBUG, "transport")
	logging.SetLevel(logging.DEBUG, "autonat")
	logging.SetLevel(logging.DEBUG, "dht")
	logging.SetLevel(logging.DEBUG, "nat")

	var (
		listenAddr  = flag.String("addr", "/ip4/0.0.0.0/tcp/9487", "listen address")
		genKey      = flag.String("genkey", "", "generate a node key")
		writeAddr   = flag.Bool("writeaddress", false, "write out the node's pubkey hash and quit")
		nodeKeyFile = flag.String("nodekey", "", "private key filename")
		nodeKeyHex  = flag.String("nodekeyhex", "", "private key as hex (for testing)")
		//netrestrict = flag.String("netrestrict", "", "restrict network communication to the given IP networks (CIDR masks)")
		verbosity = flag.Int("verbosity", int(log.LvlInfo), "log verbosity (0-9)")
		vmodule   = flag.String("vmodule", "", "log verbosity pattern")

		nodeKey *ecdsa.PrivateKey
		err     error
	)
	flag.Parse()

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(*verbosity))
	glogger.Vmodule(*vmodule)
	log.Root().SetHandler(glogger)

	// key
	switch {
	case *genKey != "":
		nodeKey, err = key.GenerateKey()
		if err != nil {
			utils.Fatalf("could not generate key: %v", err)
		}
		if err = crypto.SaveECDSA(*genKey, nodeKey); err != nil {
			utils.Fatalf("%v", err)
		}
		return
	case *nodeKeyFile == "" && *nodeKeyHex == "":
		utils.Fatalf("Use -nodekey or -nodekeyhex to specify a private key")
	case *nodeKeyFile != "" && *nodeKeyHex != "":
		utils.Fatalf("Options -nodekey and -nodekeyhex are mutually exclusive")
	case *nodeKeyFile != "":
		if nodeKey, err = crypto.LoadECDSA(*nodeKeyFile); err != nil {
			utils.Fatalf("-nodekey: %v", err)
		}
	case *nodeKeyHex != "":
		if nodeKey, err = crypto.HexToECDSA(*nodeKeyHex); err != nil {
			utils.Fatalf("-nodekeyhex: %v", err)
		}
	}

	// we got key
	privKey, err := key.PrivateKeyToP2PPrivKey(nodeKey)

	if err != nil {
		log.Error("P2PBootnode: unable to get privKey", "e", err)
		return
	}

	// id
	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		log.Error("P2PBootnode: unable to get id", "e", err)
		return

	}

	// addr
	addr, err := ma.NewMultiaddr(*listenAddr)
	if err != nil {
		log.Error("P2PBootnode: unable to get addr", "e", err)
		return
	}

	peerInfo := &pstore.PeerInfo{
		ID:    id,
		Addrs: []ma.Multiaddr{addr},
	}

	addrWithPeerInfos, err := pstore.InfoToP2pAddrs(peerInfo)
	if err != nil {
		log.Error("P2PBootnode: invalid peerInfo", "e", err)
	}
	if len(addrWithPeerInfos) != 1 {
		log.Error("P2PBootnode: invalid peerInfo", "addrWithPeerInfos", addrWithPeerInfos)
	}
	addrWithPeerInfo := addrWithPeerInfos[0]
	if *writeAddr {
		fmt.Printf("%v\n", addrWithPeerInfo)
		os.Exit(0)
	}

	// new host
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addrStr := addr.String()

	h, err := libp2p.New(
		ctx,
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(addrStr),
	)
	if err != nil {
		log.Error("P2PBootnode: unable to new host", "e", err)
	}

	log.Info("P2PBootnode: start", "host", h.ID(), "addr", h.Addrs())

	log.Info("P2PBootnode: Listening addr", "addr", addrWithPeerInfo)

	// init dht
	_, err = dht.New(ctx, h)
	if err != nil {
		return
	}

	done := make(chan struct{})

	<-done

}
