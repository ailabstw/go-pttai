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

type UploadImage struct {
	FileType string
	Bytes    []byte
}

func (pm *ProtocolManager) UploadImage(fileType string, theBytes []byte) (*pkgservice.Media, error) {
	myID := pm.Ptt().GetMyEntity().GetID()

	if pm.Entity().GetEntityType() == pkgservice.EntityTypePersonal && !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	data := &UploadImage{
		FileType: fileType,
		Bytes:    theBytes,
	}

	theMedia, err := pm.CreateObject(data, BoardOpTypeCreateMedia, pm.NewMedia, pm.NewBoardOplogWithTS, pm.inuploadImage, pm.broadcastBoardOplogCore, nil)
	if err != nil {
		return nil, err
	}

	media, ok := theMedia.(*pkgservice.Media)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	return media, nil
}

func (pm *ProtocolManager) inuploadImage(theObj pkgservice.Object, theData pkgservice.CreateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) error {

	obj, ok := theObj.(*pkgservice.Media)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	data, ok := theData.(*UploadImage)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*pkgservice.OpCreateMedia)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// new image
	newMediaType, newData, newBytes, err := pkgservice.NormalizeImage(data.Bytes)
	if err != nil {
		return err
	}

	// media
	obj.MediaData = newData
	obj.MediaType = newMediaType

	// block-info
	blockID, blockHashs, err := pm.SplitMediaBlocks(obj.ID, newBytes)
	if err != nil {
		return err
	}

	blockInfo, err := pkgservice.NewBlockInfo(blockID, blockHashs, nil, obj.CreatorID)
	if err != nil {
		return err
	}
	blockInfo.SetIsAllGood()

	theObj.SetBlockInfo(blockInfo)

	// op-data
	opData.BlockInfoID = blockID
	opData.NBlock = blockInfo.NBlock
	opData.Hashs = blockHashs

	return nil
}
