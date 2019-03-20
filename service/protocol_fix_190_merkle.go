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
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pm *BaseProtocolManager) Fix190Merkle() error {
	isFixed, err := pm.isFix190Merkle()
	if err != nil {
		return err
	}
	if isFixed {
		return nil
	}

	err = pm.ForceReconstructMerkle()
	if err != nil {
		return err
	}

	err = pm.setFix190Merkle()
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) isFix190Merkle() (bool, error) {
	key, err := pm.marshalFix190MerkleKey()
	if err != nil {
		return false, err
	}
	_, err = dbMeta.Get(key)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pm *BaseProtocolManager) marshalFix190MerkleKey() ([]byte, error) {
	entityID := pm.Entity().GetID()
	return common.Concat([][]byte{DBFix190Prefix, entityID[:]})
}

func (pm *BaseProtocolManager) setFix190Merkle() error {
	key, err := pm.marshalFix190MerkleKey()
	if err != nil {
		return err
	}

	dbMeta.Put(key, pttdb.ValueTrue)

	return nil
}
