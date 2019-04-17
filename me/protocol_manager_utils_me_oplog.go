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

package me

import (
	"bytes"
	"reflect"
	"sort"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) InternalSignMyOplog(oplog *pkgservice.BaseOplog) (bool, error) {
	if oplog.MasterLogID != nil {
		return false, nil
	}

	myEntity := pm.Entity().(*MyInfo)
	myID := myEntity.NodeSignID

	masterSigns := oplog.MasterSigns
	lenMasterSigns := len(masterSigns)

	// already signs master
	if lenMasterSigns > 0 {
		idx := sort.Search(len(masterSigns), func(i int) bool {
			return bytes.Compare(masterSigns[i].ID[:], myID[:]) >= 0
		})
		if idx >= 0 && idx < lenMasterSigns && reflect.DeepEqual(masterSigns[idx].ID, myID) {
			return false, nil
		}
	}

	err := myEntity.MyMasterSign(oplog)
	if err != nil {
		return false, err
	}

	masterLogID, weight, isValid := pm.IsValidMyOplog(oplog.MasterSigns)
	if !isValid {
		return true, nil
	}

	// master-log-id
	err = oplog.SetMasterLogID(masterLogID, weight)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *ProtocolManager) ForceSignMyOplog(oplog *pkgservice.BaseOplog) error {
	err := pm.SignOplog(oplog)
	if err != nil {
		return err
	}

	if oplog.MasterLogID != nil {
		return nil
	}

	masterLogID := pm.GetNewestMasterLogID()
	err = oplog.SetMasterLogID(masterLogID, 0)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) IsValidMyOplog(signInfos []*pkgservice.SignInfo) (*types.PttID, uint32, bool) {

	return pm.IsValidInternalOplog(signInfos)
}

func (pm *ProtocolManager) ValidateIntegrateSignMyOplog(oplog *pkgservice.BaseOplog, isLocked bool) (err error) {

	if !isLocked {
		err = oplog.Lock()
		if err != nil {
			return
		}
		defer oplog.Unlock()
	}

	masterLogID, weight, isValid := pm.IsValidMyOplog(oplog.MasterSigns)
	if isValid {
		err = oplog.SetMasterLogID(masterLogID, weight)
		if err != nil {
			return
		}
	}

	return
}
