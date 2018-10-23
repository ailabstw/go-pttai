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
var TestnetBootnodes = []string{
	"pnode://6c537f268ca2c77d4e8783f110ae9404eb937b6b81106f053c95aca37eccf80b7e2d4a0032e5a1e921ea4a18e7feb1dc75f5b42d7337f954f3e268a4bc17d5fa@172.104.122.208:9487",
}
