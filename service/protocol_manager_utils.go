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
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
)

/*
StartPM starts the pm
	1. go PMSync
	2. go PMSyncOpKeyLoop
	3. pm.Start
*/
func StartPM(pm ProtocolManager) error {
	log.Info("StartPM: start", "entity", pm.Entity().Name())

	// 1. PMSync
	pm.SyncWG().Add(1)
	go func() {
		defer pm.SyncWG().Done()

		PMSync(pm)
	}()

	// 3. pm.Start
	err := pm.Start()
	if err != nil {
		return err
	}

	return nil
}

func StopPM(pm ProtocolManager) error {
	log.Info("Stop PM: to stop", "entity", pm.Entity().Name())

	err := pm.Stop()
	if err != nil {
		log.Warn("Stop PM: unable to stop", "entity", pm.Entity().Name(), "e", err)
		return err
	}

	log.Info("Stop PM: done", "entity", pm.Entity().Name())

	return nil
}

func PMSync(pm ProtocolManager) error {
	var err error
	forceSyncTicker := time.NewTicker(pm.ForceSyncCycle())

looping:
	for {
		select {
		case peer, ok := <-pm.NewPeerCh():
			if !ok {
				break looping
			}

			err = pm.Sync(peer)
			if err != nil {
				log.Error("unable to Sync after newPeer", "e", err)
			}
		case <-forceSyncTicker.C:
			forceSyncTicker.Stop()
			forceSyncTicker = time.NewTicker(pm.ForceSyncCycle())

			err = pm.Sync(nil)
			if err != nil {
				log.Error("unable to Sync after forceSync", "e", err)
			}
		case <-pm.QuitSync():
			return p2p.DiscQuitting
		}
	}
	forceSyncTicker.Stop()

	return nil
}

func PMHandleMessageWrapper(pm ProtocolManager, hash *common.Address, encData []byte, peer *PttPeer) error {
	opKeyInfo, err := pm.GetOpKeyInfoFromHash(hash)

	if err != nil {
		return err
	}

	op, dataBytes, err := pm.Ptt().DecryptData(encData, opKeyInfo)
	//log.Debug("PMHandleMessageWrapper: after DecryptData", "e", err, "op", op)
	if err != nil {
		return err
	}

	switch op {
	case IdentifyPeerMsg:
		return pm.HandleIdentifyPeer(dataBytes, peer)
	case IdentifyPeerAckMsg:
		return pm.HandleIdentifyPeerAck(dataBytes, peer)
	}

	fitPeerType := pm.GetPeerType(peer)

	if fitPeerType < PeerTypeMember {
		return ErrInvalidEntity
	}

	return pm.HandleMessage(op, dataBytes, peer)
}

/*
Send Data to Peers using op-key
*/
func (pm *BaseProtocolManager) SendDataToPeers(op OpType, data interface{}, peerList []*PttPeer) error {

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	opKeyInfo, err := pm.GetOldestOpKey(false)
	if err != nil {
		return err
	}

	ptt := pm.Ptt()
	encData, err := ptt.EncryptData(op, dataBytes, opKeyInfo)
	if err != nil {
		return err
	}

	pttData, err := ptt.MarshalData(CodeTypeOp, opKeyInfo.Hash, encData)
	if err != nil {
		return err
	}

	okCount := 0
	for _, peer := range peerList {
		pttData.Node = peer.GetID()[:]
		err := peer.SendData(pttData)
		if err == nil {
			okCount++
		} else {
			log.Warn("PMSendDataToPeers: unable to SendData", "peer", peer, "e", err)
		}
	}
	if okCount == 0 {
		return ErrNotSent
	}

	return nil
}

func (pm *BaseProtocolManager) SendDataToPeer(op OpType, data interface{}, peer *PttPeer) error {
	return pm.SendDataToPeerWithCode(CodeTypeOp, op, data, peer)
}

/*
Send Data to Peers using op-key
*/
func (pm *BaseProtocolManager) SendDataToPeerWithCode(code CodeType, op OpType, data interface{}, peer *PttPeer) error {

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	opKeyInfo, err := pm.GetOldestOpKey(false)
	if err != nil {
		return err
	}

	ptt := pm.Ptt()
	encData, err := ptt.EncryptData(op, dataBytes, opKeyInfo)
	if err != nil {
		return err
	}

	pttData, err := ptt.MarshalData(code, opKeyInfo.Hash, encData)
	if err != nil {
		return err
	}

	pttData.Node = peer.GetID()[:]

	err = peer.SendData(pttData)
	if err != nil {
		return ErrNotSent
	}

	return nil
}
