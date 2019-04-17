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

package service

import (
	"crypto/ecdsa"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func joinKeyToKeyInfo(key *ecdsa.PrivateKey) *KeyInfo {
	return &KeyInfo{
		Key:         key,
		KeyBytes:    crypto.FromECDSA(key),
		PubKeyBytes: crypto.FromECDSAPub(&key.PublicKey),
	}
}

func (p *BasePtt) AddJoinKey(hash *common.Address, entityID *types.PttID, isLocked bool) error {
	if !isLocked {
		p.LockJoins()
		defer p.UnlockJoins()
	}

	log.Debug("AddJoinKey: start", "hash", hash, "entityID", entityID)

	p.joins[*hash] = entityID

	return nil
}

func (p *BasePtt) RemoveJoinKey(hash *common.Address, entityID *types.PttID, isLocked bool) error {
	if !isLocked {
		p.LockJoins()
		defer p.UnlockJoins()
	}

	log.Debug("RemoveJoinKey: start", "hash", hash, "entityID", entityID)

	delete(p.joins, *hash)

	return nil
}

func (p *BasePtt) LockJoins() {
	p.lockJoins.Lock()
}

func (p *BasePtt) UnlockJoins() {
	p.lockJoins.Unlock()
}
