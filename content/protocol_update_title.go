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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type UpdateTitle struct {
	Title []byte `json:"t"`
}

func (pm *ProtocolManager) UpdateTitle(title []byte) error {
	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	data := &UpdateTitle{Title: title}

	origObj := NewEmptyTitle()
	pm.SetTitleDB(origObj)

	opData := &BoardOpUpdateTitle{}

	entityID := pm.Entity().GetID()
	log.Debug("UpdateTitle: to UpdateObject")

	err := pm.UpdateObject(
		entityID,
		data,
		BoardOpTypeUpdateTitle,
		origObj,
		opData,

		pm.boardOplogMerkle,

		pm.SetBoardDB,
		pm.NewBoardOplog,
		pm.inupdateTitle,
		nil,
		pm.broadcastBoardOplogCore,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) inupdateTitle(obj pkgservice.Object, theData pkgservice.UpdateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	data, ok := theData.(*UpdateTitle)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*BoardOpUpdateTitle)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// op-data
	opData.TitleHash = types.Hash(data.Title)

	// sync-info
	syncInfo := NewEmptySyncTitleInfo()
	syncInfo.InitWithOplog(oplog.ToStatus(), oplog)

	syncInfo.Title = data.Title

	return syncInfo, nil
}
