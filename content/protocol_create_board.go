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

type CreateBoard struct {
	Title      []byte                `json:"T"`
	EntityType pkgservice.EntityType `json:"E"`
}

func (spm *ServiceProtocolManager) CreateBoard(title []byte, entityType pkgservice.EntityType) (*Board, error) {

	data := &CreateBoard{
		Title:      title,
		EntityType: entityType,
	}

	entity, err := spm.CreateEntity(data, BoardOpTypeCreateBoard, spm.NewBoard, spm.NewBoardOplogWithTS, nil, nil)
	if err != nil {
		return nil, err
	}

	board, ok := entity.(*Board)
	if !ok {
		return nil, pkgservice.ErrInvalidEntity
	}

	return board, nil
}

func (spm *ServiceProtocolManager) NewBoard(theData pkgservice.CreateData, ptt pkgservice.Ptt, service pkgservice.Service) (pkgservice.Entity, pkgservice.OpData, error) {

	data, ok := theData.(*CreateBoard)
	if !ok {
		return nil, nil, pkgservice.ErrInvalidData
	}

	myID := spm.Ptt().GetMyEntity().GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	board, err := NewBoard(myID, ts, ptt, service, spm, spm.GetDBLock())
	if err != nil {
		return nil, nil, err
	}
	board.EntityType = data.EntityType
	board.Title = data.Title

	return board, &BoardOpCreateBoard{Title: data.Title}, nil
}
