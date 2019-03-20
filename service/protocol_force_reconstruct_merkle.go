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
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *BaseProtocolManager) ForceReconstructMerkle() error {
	err := pm.forceReconstructMerkleCore(pm.masterMerkle, pm.SetMasterDB)
	if err != nil {
		return err
	}

	err = pm.forceReconstructMerkleCore(pm.memberMerkle, pm.SetMemberDB)
	if err != nil {
		return err
	}

	err = pm.forceReconstructMerkleCore(pm.log0Merkle, pm.SetLog0DB)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) forceReconstructMerkleCore(
	merkle *Merkle,

	setLogDB func(oplog *BaseOplog),
) error {

	if merkle == nil {
		return nil
	}

	merkle.Clean()

	oplog := &BaseOplog{}
	setLogDB(oplog)

	iter, err := GetOplogIterWithOplog(oplog, nil, pttdb.ListOrderNext, types.StatusAlive, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	var val []byte

	for iter.Next() {
		val = iter.Value()
		err = oplog.Unmarshal(val)
		if err != nil {
			continue
		}

		if !oplog.IsSync {
			continue
		}

		oplog.Save(false, merkle)
	}

	return nil
}
