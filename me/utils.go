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

package me

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

func renewMe(c *Config, newKey *ecdsa.PrivateKey, newPostfixBytes []byte) error {
	err := c.DeleteKey()
	if err != nil {
		return err
	}

	newKeyHex := hex.EncodeToString(crypto.FromECDSA(newKey))

	log.Debug("renewMe: to SetMyKey")
	err = c.SetMyKey(newKeyHex, "", string(newPostfixBytes), true)
	log.Debug("renewMe: after SetMyKey", "e", err)
	if err != nil {
		return err
	}

	return nil
}

func setNodeSignID(nodeID *discover.NodeID, myID *types.PttID) (*types.PttID, error) {
	nodeIDPubkey, err := nodeID.Pubkey()
	if err != nil {
		return nil, err
	}

	nodeSignID, err := types.NewPttIDWithPubkeyAndRefID(nodeIDPubkey, myID)
	if err != nil {
		return nil, err
	}

	return nodeSignID, nil
}
