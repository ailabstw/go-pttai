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
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type UpdateNameCard struct {
	Card []byte `json:"n"`
}

func (pm *ProtocolManager) UpdateNameCard(card []byte) (*NameCard, error) {
	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	data := &UpdateNameCard{Card: card}

	origObj := NewEmptyNameCard()
	pm.SetNameCardDB(origObj)

	opData := &UserOpUpdateNameCard{}

	err := pm.UpdateObject(
		myID,

		data,
		UserOpTypeUpdateNameCard,

		origObj,

		opData,

		pm.userOplogMerkle,

		pm.SetUserDB,

		pm.NewUserOplog,

		pm.inupdateNameCard,

		nil,

		pm.broadcastUserOplogCore,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return origObj, nil
}

func (pm *ProtocolManager) inupdateNameCard(obj pkgservice.Object, theData pkgservice.UpdateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	data, ok := theData.(*UpdateNameCard)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*UserOpUpdateNameCard)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// op-data
	opData.Hash = types.Hash(data.Card)

	// sync-info
	syncInfo := NewEmptySyncNameCardInfo()
	syncInfo.InitWithOplog(oplog.ToStatus(), oplog)

	syncInfo.Card = data.Card

	return syncInfo, nil
}
