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
	"encoding/binary"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
)

func (p *BasePtt) CreateMasterOplog(raftIdx uint64, ts types.Timestamp, op OpType, data interface{}) (*MasterOplog, error) {
	key := p.SignKey()
	myID := p.myEntity.GetID()
	nodeSignID := p.myEntity.GetNodeSignID()

	oplog, err := NewMasterOplog(myID, ts, nodeSignID, op, data)
	if err != nil {
		return nil, err
	}

	// oplog.ID
	binary.BigEndian.PutUint64(oplog.ID[OffsetMasterOplogRaftIdx:], raftIdx)
	copy(oplog.ID[common.AddressLength:], myID[:common.AddressLength])

	err = oplog.Sign(key)
	if err != nil {
		return nil, err
	}

	p.myEntity.PM().SetNewestMasterLogID(oplog.ID)
	oplog.MasterLogID = oplog.ID

	return oplog, nil
}
