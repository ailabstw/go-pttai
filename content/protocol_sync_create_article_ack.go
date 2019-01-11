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
	"encoding/json"

	pkgservice "github.com/ailabstw/go-pttai/service"
)

type SyncArticleAck struct {
	Objs []*Article `json:"o"`
}

func (pm *ProtocolManager) HandleSyncCreateArticleAck(dataBytes []byte, peer *pkgservice.PttPeer) error {

	data := &SyncArticleAck{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	origObj := NewEmptyArticle()
	pm.SetArticleDB(origObj)
	for _, obj := range data.Objs {
		pm.SetArticleDB(obj)

		pm.HandleSyncCreateObjectAck(
			obj,
			peer,
			origObj,

			pm.boardOplogMerkle,

			pm.SetBoardDB,
			pm.updateSyncCreateArticle,
			pm.postcreateArticle,
			pm.broadcastBoardOplogCore,
		)
	}

	return nil
}

func (pm *ProtocolManager) updateSyncCreateArticle(theToObj pkgservice.Object, theFromObj pkgservice.Object) error {
	toObj, ok := theToObj.(*Article)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	fromObj, ok := theFromObj.(*Article)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	toObj.BlockInfo = fromObj.BlockInfo
	toObj.Title = fromObj.Title

	return nil
}
