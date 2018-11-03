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

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

type InitMeInfoSync struct {
	KeyBytes     []byte            `json:"K"`
	PostfixBytes []byte            `json:"P"`
	UserName     *account.UserName `json:"U"`
	UserImg      *account.UserImg  `json:"I"`
}

func (pm *ProtocolManager) InitMeInfoSync(peer *pkgservice.PttPeer) error {
	log.Debug("InitMeInfoSync: start")
	var err error
	myInfo := pm.Entity().(*MyInfo)
	myID := myInfo.ID

	if myInfo.Status != types.StatusAlive {
		return nil
	}

	// private-key
	keyBytes := crypto.FromECDSA(myInfo.GetMyKey())

	// user-name
	userName := &account.UserName{}
	err = userName.Get(myID, true)
	if err == leveldb.ErrNotFound {
		err = nil
		userName = nil
	}
	if err != nil {
		return err
	}

	// user-img
	userImg := &account.UserImg{}
	err = userImg.Get(myID, true)
	if err == leveldb.ErrNotFound {
		err = nil
		userImg = nil
	}
	if err != nil {
		return err
	}

	// send-data-to-peer
	data := &InitMeInfoSync{
		KeyBytes:     keyBytes,
		PostfixBytes: myID[common.AddressLength:],
		UserName:     userName,
		UserImg:      userImg,
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

	log.Debug("HandleInitMeInfoSync: start")

	data := &InitMeInfoSync{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	// userName
	userName := data.UserName
	if userName != nil {
		err = userName.Save(true)
		if err != nil {
			return err
		}
	}

	// userImg
	userImg := data.UserImg
	if userImg != nil {
		err = userImg.Save(true)
		if err != nil {
			return err
		}
	}

	// renew-me
	cfg := pm.Entity().Service().(*Backend).Config
	newKey, err := crypto.ToECDSA(data.KeyBytes)
	err = renewMe(cfg, newKey, data.PostfixBytes)
	if err != nil {
		return err
	}

	myInfo.Status = types.StatusInternalSync
	myInfo.UpdateTS = ts
	err = myInfo.Save()
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
