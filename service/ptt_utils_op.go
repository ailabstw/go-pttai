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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"
)

func (p *BasePtt) AddOpKey(hash *common.Address, entityID *types.PttID, isLocked bool) error {
	if !isLocked {
		p.LockOps()
		defer p.UnlockOps()
	}

	log.Debug("AddOpkey: to add key", "hash", hash, "entityID", entityID)

	p.ops[*hash] = entityID

	return nil
}

func (p *BasePtt) RemoveOpKey(hash *common.Address, entityID *types.PttID, isLocked bool) error {
	if !isLocked {
		p.LockOps()
		defer p.UnlockOps()
	}

	log.Debug("RemoveOpKey: to remove key", "hash", hash, "entityID", entityID)

	delete(p.ops, *hash)

	return nil
}

func (p *BasePtt) LockOps() {
	p.lockOps.Lock()
}

func (p *BasePtt) UnlockOps() {
	p.lockOps.Unlock()
}

func (p *BasePtt) RemoveOpHash(hash *common.Address) error {
	entityID, ok := p.ops[*hash]
	if !ok {
		return nil
	}

	entity, ok := p.entities[*entityID]
	if !ok {
		return p.RemoveOpKey(hash, entityID, false)
	}

	return entity.PM().RemoveOpKeyFromHash(hash, false, true, true)
}
