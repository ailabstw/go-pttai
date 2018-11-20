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

package content

import "github.com/syndtr/goleveldb/leveldb"

func (pm *ProtocolManager) SetTitle(title []byte) error {

	isExists, err := pm.setTitleCheckIsExists()
	if err != nil {
		return err
	}

	if !isExists {
		return pm.CreateTitle(title)
	}

	return pm.UpdateTitle(title)
}

func (pm *ProtocolManager) setTitleCheckIsExists() (bool, error) {
	entityID := pm.Entity().GetID()

	title := NewEmptyTitle()
	pm.SetTitleDB(title)
	title.SetID(entityID)

	// lock
	err := title.RLock()
	if err != nil {
		return false, err
	}
	defer title.RUnlock()

	// get
	err = title.GetByID(true)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
