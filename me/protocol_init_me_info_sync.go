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
	"encoding/json"
	"reflect"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type InitMeInfoSync struct {
	KeyBytes     []byte                        `json:"K"`
	PostfixBytes []byte                        `json:"P"`
	Oplog0       *pkgservice.BaseOplog         `json:"O"`
	ProfileData  *account.ApproveJoinEntity    `json:"p"`
	BoardData    *pkgservice.ApproveJoinEntity `json:"b"`
}

func (pm *ProtocolManager) InitMeInfoSync(peer *pkgservice.PttPeer) error {
	log.Debug("InitMeInfoSync: start")
	var err error
	myInfo := pm.Entity().(*MyInfo)
	myID := myInfo.ID

	err = myInfo.Lock()
	if err != nil {
		return err
	}
	defer myInfo.Unlock()

	if myInfo.Status != types.StatusAlive {
		return nil
	}

	// oplog0
	oplog0 := pm.GetOplog0()

	// private-key
	keyBytes := crypto.FromECDSA(myInfo.GetMyKey())

	// profile
	myEntity := pm.Entity().(*MyInfo)
	profile := myEntity.Profile
	profilePM := profile.PM()

	joinEntity := &pkgservice.JoinEntity{ID: myID}
	_, theProfileData, err := profilePM.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return err
	}
	profileData, ok := theProfileData.(*account.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// board
	board := myEntity.Board
	boardPM := board.PM()

	_, theBoardData, err := boardPM.ApproveJoin(joinEntity, nil, peer)
	if err != nil {
		return err
	}
	boardData, ok := theBoardData.(*pkgservice.ApproveJoinEntity)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// send-data-to-peer
	data := &InitMeInfoSync{
		KeyBytes:     keyBytes,
		PostfixBytes: myID[common.AddressLength:],
		Oplog0:       oplog0,
		ProfileData:  profileData,
		BoardData:    boardData,
	}

	err = pm.SendDataToPeer(InitMeInfoSyncMsg, data, peer)
	if err != nil {
		return err
	}

	log.Debug("InitMeInfoSync: done")

	return nil
}

func (pm *ProtocolManager) HandleInitMeInfoSync(dataBytes []byte, peer *pkgservice.PttPeer) error {
	myInfo := pm.Entity().(*MyInfo)

	// validate
	if !reflect.DeepEqual(myInfo.ID, peer.UserID) {
		return types.ErrInvalidID
	}

	// lock
	log.Debug("HandleInitMeInfoSync: start")

	err := myInfo.Lock()
	if err != nil {
		return err
	}
	defer myInfo.Unlock()

	theProfileData := account.NewEmptyApproveJoinProfile()
	theBoardData := content.NewEmptyApproveJoinBoard()
	data := &InitMeInfoSync{ProfileData: theProfileData, BoardData: theBoardData}
	err = json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}
	profileData, boardData := data.ProfileData, data.BoardData

	// migrate origin-me
	origMe := pm.Ptt().GetMyEntity().(*MyInfo)
	err = origMe.PM().(*ProtocolManager).MigrateMe(myInfo)
	log.Debug("HandleInitMeInfoSync: after MigrateMe", "e", err)
	if err != nil {
		return err
	}

	// set new info
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	// oplog0
	oplog0 := data.Oplog0
	log.Debug("HandleInitMeInfoSync: oplog0", "oplog0", oplog0)
	pm.SetMeDB(oplog0)
	oplog0.Save(false, pm.meOplogMerkle)

	// profile
	profileSPM := pm.Entity().Service().(*Backend).accountBackend.SPM().(*account.ServiceProtocolManager)
	_, err = profileSPM.CreateJoinEntity(profileData, peer, nil, false, true, true, false, false)
	if err != nil {
		return err
	}

	// board
	contentSPM := pm.Entity().Service().(*Backend).contentBackend.SPM().(*content.ServiceProtocolManager)
	_, err = contentSPM.CreateJoinEntity(boardData, peer, nil, false, true, true, false, false)
	if err != nil {
		return err
	}

	// renew-me
	cfg := pm.Entity().Service().(*Backend).Config
	newKey, err := crypto.ToECDSA(data.KeyBytes)
	err = renewMe(cfg, newKey, data.PostfixBytes)
	log.Debug("HandleInitMeInfoSync: after renewMe", "e", err)
	if err != nil {
		return err
	}
	myInfo.LogID = oplog0.ID
	myInfo.Status = types.StatusSync
	myInfo.UpdateTS = ts
	err = myInfo.Save(true)
	if err != nil {
		return err
	}

	// notify the other that my status changed.
	pm.SendDataToPeer(InitMeInfoAckMsg, &InitMeInfoAck{Status: myInfo.Status}, peer)

	log.Debug("HandleInitMeInfoSync: to restart")

	// restart
	pm.myPtt.NotifyNodeRestart().PassChan(struct{}{})

	return nil
}
