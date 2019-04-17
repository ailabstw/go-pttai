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

package account

import (
	"encoding/json"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleForceSyncUserNameAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	data := &SyncUserNameAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyUserName()
	pm.SetUserNameDB(origObj)

	for _, obj := range data.Objs {
		pm.SetUserNameDB(obj)

		err = pm.HandleForceSyncObjectAck(
			obj,
			peer,

			origObj,

			pm.userOplogMerkle,

			pm.SetUserDB,
		)
		if err != nil {
			continue
		}
	}

	return nil
}
