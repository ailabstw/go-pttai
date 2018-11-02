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

package me

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type InitMeInfo struct {
}

func (pm *ProtocolManager) InitMeInfo(peer *pkgservice.PttPeer) error {
	data := &InitMeInfo{}
	err := pm.SendDataToPeer(InitMeInfoMsg, data, peer)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) HandleInitMeInfo(dataBytes []byte, peer *pkgservice.PttPeer) error {
	data := &InitMeInfo{}

	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return pm.InitMeInfoAck(data, peer)
}

func (pm *ProtocolManager) InitMeInfoLoop() {
	tick := time.NewTicker(InitMeInfoTickTime)
	defer tick.Stop()

loop:
	for {
		select {
		case <-tick.C:
			pm.initMeInfoLoopCore()
		case <-pm.QuitSync():
			break loop
		}
	}
}

func (pm *ProtocolManager) initMeInfoLoopCore() {
	myInfo := pm.Entity().(*MyInfo)
	if myInfo.Status != types.StatusAlive {
		return
	}

	peerList := pm.Peers().PeerList(false)

	myNodeID := pm.myPtt.MyNodeID()
	for _, peer := range peerList {
		peerID := peer.GetID()
		if reflect.DeepEqual(peerID, myNodeID) {
			continue
		}

		raftID, err := peerID.ToRaftID()
		if err != nil {
			continue
		}

		myNode := pm.MyNodes[raftID]
		if myNode == nil {
			continue
		}

		if myNode.Status == types.StatusInternalPending {
			pm.InitMeInfo(peer)
		}
	}
}
