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

import "github.com/ailabstw/go-pttai/common/types"

/*
ToConfirmJoin puts the joinEntity into confirm-join-map and wait for confirming the join. (invitor)
*/
func (p *BasePtt) ToConfirmJoin(confirmKey []byte, entity Entity, joinEntity *JoinEntity, keyInfo *KeyInfo, peer *PttPeer, joinType JoinType) error {

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	confirmJoin := &ConfirmJoin{
		Entity:     entity,
		JoinEntity: joinEntity,
		KeyInfo:    keyInfo,
		Peer:       peer,
		UpdateTS:   ts,
		JoinType:   joinType,
	}

	confirmKeyStr := string(confirmKey)

	p.lockConfirmJoin.Lock()
	defer p.lockConfirmJoin.Unlock()

	_, ok := p.confirmJoins[confirmKeyStr]
	if ok {
		return types.ErrAlreadyExists
	}

	p.confirmJoins[confirmKeyStr] = confirmJoin

	return nil
}
