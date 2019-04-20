// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main Ethereum network.
var MainnetBootnodes = []string{}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{}

var MainP2PBootnodes = []string{}

var TestP2PBootnodes = []string{
	"pnode://16Uiu2HAm2iXrcfxL5mG3EKQ2Hn7gFEPRiAG98FEjpcqHW6GZAJEn@172.104.122.208:9487",
}

var IPFSBootnodes = []string{
	"pnode://QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ@104.131.131.82:4001",
	"pnode://QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM@104.236.179.241:4001",
	"pnode://QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64@104.236.76.40:4001",
	"pnode://QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu@128.199.219.111:4001",
	"pnode://QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd@178.62.158.247:4001",
}

var DevP2PBootnodes = []string{
	"pnode://16Uiu2HAmJjpTxFgUu4WT57D1jK1HpQbtrNzZwJVHEewpv2wTCUqJ@10.1.1.27:9487",
}

var MainSignalServerURL = ""
var TestSignalServerAddr = "testnet-signal.ptt.ai:9488"
