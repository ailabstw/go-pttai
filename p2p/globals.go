// Copyright 2019 The go-pttai Authors
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

package p2p

import (
	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// p2p
const (
	SleepTimeSecondAnnounceP2P = 120
	TimeoutSecondAnnounceP2P   = 10

	TimeoutSecondConnectP2P = 5
	TimeoutSecondResolveP2P = 10

	PTTAI_STREAM_PATH = "/pttai/0.3.0"

	SizePadSpace = 300
)

var (
	v1b                = cid.V1Builder{Codec: cid.Raw, MhType: mh.SHA2_256}
	rendezvousString   = "PTTAI_RENDEZVOUS"
	RendezvousPoint, _ = v1b.Sum([]byte(rendezvousString))
)
