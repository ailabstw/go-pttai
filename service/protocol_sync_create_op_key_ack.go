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

package service

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
)

type SyncCreateOpKeyAck struct {
	Objs []*KeyInfo `json:"o"`
}

func (pm *BaseProtocolManager) HandleSyncCreateOpKeyAck(dataBytes []byte, peer *PttPeer) error {
	data := &SyncCreateOpKeyAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyOpKey()
	pm.SetOpKeyObjDB(origObj)
	for _, obj := range data.Objs {
		pm.SetOpKeyObjDB(obj)

		pm.HandleSyncCreateObjectAck(
			obj,
			peer,

			origObj,

			nil,

			pm.SetOpKeyDB,
			pm.updateCreateOpKey,
			pm.postcreateOpKey,
			pm.broadcastOpKeyOplogCore,
		)
	}

	return nil
}

func (pm *BaseProtocolManager) HandleSyncCreateOpKeyAckObj(opKey *KeyInfo, peer *PttPeer) error {

	origObj := NewEmptyOpKey()
	pm.SetOpKeyObjDB(origObj)

	pm.SetOpKeyObjDB(opKey)

	return pm.HandleSyncCreateObjectAck(
		opKey,
		peer,

		origObj,
		nil,

		pm.SetOpKeyDB,
		pm.updateCreateOpKey,
		pm.postcreateOpKey,
		pm.broadcastOpKeyOplogCore,
	)
}

/***
 * syncCreateObject
 ***/

func (pm *BaseProtocolManager) updateCreateOpKey(theToObj Object, theFromObj Object) error {
	toObj, ok := theToObj.(*KeyInfo)
	if !ok {
		return ErrInvalidObject
	}

	fromObj, ok := theFromObj.(*KeyInfo)
	if !ok {
		return ErrInvalidObject
	}

	key, err := crypto.ToECDSA(fromObj.KeyBytes)
	if err != nil {
		return err
	}

	//toObj.BaseObject = fromObj.BaseObject
	//pm.SetOpKeyObjDB(toObj)

	toObj.Hash = fromObj.Hash
	toObj.Key = key
	toObj.KeyBytes = fromObj.KeyBytes
	toObj.PubKeyBytes = crypto.FromECDSAPub(&key.PublicKey)
	toObj.Extra = fromObj.Extra

	return nil
}
