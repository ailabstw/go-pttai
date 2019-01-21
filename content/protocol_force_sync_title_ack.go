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

import (
	"encoding/json"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleForceSyncTitleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	data := &SyncTitleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyTitle()
	pm.SetTitleDB(origObj)

	for _, obj := range data.Objs {
		pm.SetTitleDB(obj)

		err = pm.HandleForceSyncObjectAck(
			obj,
			peer,

			origObj,

			pm.boardOplogMerkle,

			pm.SetBoardDB,
		)
		if err != nil {
			continue
		}
	}

	return nil
}
