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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type InitMeInfo struct {
	Status types.Status `json:"S"`
}

func (pm *ProtocolManager) InitMeInfo(peer *pkgservice.PttPeer) error {
	log.Debug("InitMeInfo: start")

	myInfo := pm.Entity().(*MyInfo)

	data := &InitMeInfo{Status: myInfo.Status}
	err := pm.SendDataToPeer(InitMeInfoMsg, data, peer)
	if err != nil {
		return err
	}

	log.Debug("InitMeInfo: done")

	return nil
}

func (pm *ProtocolManager) HandleInitMeInfo(dataBytes []byte, peer *pkgservice.PttPeer) error {
	data := &InitMeInfo{}

	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	pm.handleInitMeInfoCore(data.Status, peer)

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
			log.Debug("InitMeInfoLoop: QuitSync", "entity", pm.Entity().GetID())
			break loop
		}
	}
}

func (pm *ProtocolManager) initMeInfoLoopCore() {
	myInfo := pm.Entity().(*MyInfo)
	if myInfo.Status != types.StatusAlive {
		return
	}

	pendingPeerList := pm.Peers().PendingPeerList(false)
	peerList := pm.Peers().PeerList(false)
	peerList = append(peerList, pendingPeerList...)

	log.Debug("initMeInfoLoopCore", "peerList", peerList, "me", myInfo.ID)

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
			log.Warn("initMeInfoLoopCore: myNode as nil", "peer", peer)
			continue
		}

		log.Debug("initMeInfoLoopCore: to check status", "peer", peer, "status", myNode.Status)

		if myNode.Status != types.StatusAlive {
			pm.InitMeInfo(peer)
		}
	}
}
