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

package service

import (
	"crypto/ecdsa"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

/**********
 * Me
 **********/

func (p *BasePtt) MyNodeID() *discover.NodeID {
	return p.myNodeID
}

func (p *BasePtt) MyRaftID() uint64 {
	return p.myRaftID
}

func (p *BasePtt) MyNodeType() NodeType {
	return p.myNodeType
}

func (p *BasePtt) MyNodeKey() *ecdsa.PrivateKey {
	return p.myNodeKey
}

func (p *BasePtt) SetMyEntity(myEntity PttMyEntity) error {
	p.myEntity = myEntity
	p.myService = myEntity.Service()

	return nil
}

func (p *BasePtt) GetMyEntity() MyEntity {
	return p.myEntity
}

func (p *BasePtt) GetMyEntityFromMe(myID *types.PttID) Entity {
	return nil
}

func (p *BasePtt) GetMyService() Service {
	return p.myService
}
