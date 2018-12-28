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

	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) HandleApproveJoinBoard(dataBytes []byte, joinRequest *pkgservice.JoinRequest, peer *pkgservice.PttPeer) error {

	theBoardData := content.NewEmptyApproveJoinBoard()
	approveJoin := &pkgservice.ApproveJoin{Data: theBoardData}
	err := json.Unmarshal(dataBytes, approveJoin)
	if err != nil {
		log.Error("HandleApproveJoinBoard: unable to unmarshal", "e", err)
		return err
	}

	boardData := theBoardData

	// board
	contentService := pm.Entity().Service().(*Backend).contentBackend
	contentSPM := contentService.SPM().(*content.ServiceProtocolManager)
	_, err = contentSPM.CreateJoinEntity(boardData, peer, nil, true, true, false, false, true)
	if err != nil {
		return err
	}

	// remove joinBoardRequest
	pm.lockJoinBoardRequest.Lock()
	defer pm.lockJoinBoardRequest.Unlock()
	delete(pm.joinBoardRequests, *joinRequest.Hash)

	return nil
}
