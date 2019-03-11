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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

type IdentifyPeer struct {
	Challenge *types.Salt `json:"C"`
}

/*
IdentifyPeer identifies the peer by providing the op-key of the pm (requester)
	1. generate salt
	2. initialize info in peer
	3. send data to peer
*/
func (pm *BaseProtocolManager) IdentifyPeer(peer *PttPeer) {
	log.Debug("IdentifyPeer: start", "entity", pm.Entity().Name(), "service", pm.Entity().Service().Name(), "nodeID", peer.ID(), "userID", peer.UserID)

	if peer.UserID != nil {
		return
	}

	ptt := pm.Ptt()
	data, err := ptt.IdentifyPeer(pm.Entity().GetID(), pm.QuitSync(), peer)
	if err != nil {
		log.Warn("IdentifyPeer: unable to ptt.IdentifyPeer", "e", err, "p", peer, "userID", peer.UserID)
		return
	}

	pm.SendDataToPeerWithCode(CodeTypeIdentifyPeer, IdentifyPeerMsg, data, peer)
}

/*
HandleIdentifyPeer handles IdentifyPeer (acker)
*/
func (pm *BaseProtocolManager) HandleIdentifyPeer(dataBytes []byte, peer *PttPeer) error {
	log.Debug("HandleIdentifyPeer: start", "peer", peer.ID(), "userID", peer.UserID)
	data := &IdentifyPeer{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	if pm.Entity().GetStatus() > types.StatusAlive {
		return types.ErrAlreadyDeleted
	}

	return pm.IdentifyPeerAck(data, peer)
}

/**********
 * Ptt
 **********/

func (p *BasePtt) IdentifyPeer(entityID *types.PttID, quitSync chan struct{}, peer *PttPeer) (*IdentifyPeer, error) {

	// 2. init info in peer
	salt, err := peer.InitID(entityID, quitSync)
	if err != nil {
		return nil, err
	}

	// 3. send data to peer
	data := &IdentifyPeer{
		Challenge: salt,
	}

	return data, nil
}
