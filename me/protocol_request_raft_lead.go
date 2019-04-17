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
	"time"
)

type RequestRaftLead struct{}

func (pm *ProtocolManager) EnsureRaftLead() error {
	myRaftID := pm.myPtt.MyRaftID()

	err := ErrUnableToBeLead
	var raftLead uint64
	for i := 0; i < NRequestRaftLead; i++ {
		raftLead = pm.GetRaftLead(false)

		if raftLead == myRaftID {
			err = nil
			break
		}

		if raftLead != 0 {
			pm.ProposeRaftRequestLead()
		}

		time.Sleep(3 * time.Second)
	}
	return err
}
