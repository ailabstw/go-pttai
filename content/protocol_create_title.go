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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type CreateTitle struct {
	Title []byte
}

func (pm *ProtocolManager) CreateTitle(title []byte) error {
	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return types.ErrInvalidID
	}

	data := &CreateTitle{Title: title}

	_, err := pm.CreateObject(
		data,
		BoardOpTypeCreateTitle,

		pm.boardOplogMerkle,

		pm.NewTitle,
		pm.NewBoardOplogWithTS,
		nil,

		pm.SetBoardDB,
		pm.broadcastBoardOplogsCore,
		pm.broadcastBoardOplogCore,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProtocolManager) NewTitle(theData pkgservice.CreateData) (pkgservice.Object, pkgservice.OpData, error) {

	data, ok := theData.(*CreateTitle)
	if !ok {
		return nil, nil, pkgservice.ErrInvalidData
	}

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	opData := &BoardOpCreateTitle{
		TitleHash: types.Hash(data.Title),
	}

	title, err := NewTitle(ts, myID, entityID, nil, types.StatusInit, nil)
	if err != nil {
		return nil, nil, err
	}
	pm.SetTitleDB(title)

	// set title
	title.Title = data.Title

	return title, opData, nil
}
