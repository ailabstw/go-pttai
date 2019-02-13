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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *ProtocolManager) Fix141MasterOplog() error {
	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return nil
	}

	// oplog
	oplogs, err := pm.GetMasterOplogList(nil, 0, pttdb.ListOrderNext, types.StatusAlive)
	if err != nil {
		return err
	}

	lenOplogs := len(oplogs)

	if lenOplogs == 2 {
		return nil
	}

	f := pm.Entity().(*Friend)

	if !pm.IsMember(f.FriendID, false) {
		return nil
	}

	pm.AddMaster(f.FriendID, false)

	return nil
}
