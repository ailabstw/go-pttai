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

package friend

import (
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type CreateMessage struct {
	Msg      [][]byte
	MediaIDs []*types.PttID
}

func (pm *ProtocolManager) CreateMessage(msg [][]byte, mediaIDs []*types.PttID) (*Message, error) {

	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	data := &CreateMessage{
		Msg:      msg,
		MediaIDs: mediaIDs,
	}

	theMessage, err := pm.CreateObject(
		data,
		FriendOpTypeCreateMessage,

		pm.friendOplogMerkle,

		pm.NewMessage,
		pm.NewFriendOplogWithTS,
		pm.increateMessage,

		pm.SetFriendDB,
		pm.broadcastFriendOplogsCore,
		pm.broadcastFriendOplogCore,

		pm.postcreateMessage,
	)
	if err != nil {
		return nil, err
	}

	message, ok := theMessage.(*Message)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	return message, nil
}

func (pm *ProtocolManager) NewMessage(theData pkgservice.CreateData) (pkgservice.Object, pkgservice.OpData, error) {

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	opData := &FriendOpCreateMessage{}

	userName, err := NewMessage(ts, myID, entityID, nil, types.StatusInit)
	if err != nil {
		return nil, nil, err
	}
	pm.SetMessageDB(userName)

	return userName, opData, nil
}

func (pm *ProtocolManager) increateMessage(theObj pkgservice.Object, theData pkgservice.CreateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) error {

	obj, ok := theObj.(*Message)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	data, ok := theData.(*CreateMessage)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*FriendOpCreateMessage)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// block-info
	blockID, blockHashs, err := pm.SplitContentBlocks(nil, obj.ID, data.Msg, NFirstLineInBlock)
	log.Debug("increateMessage: after SplitContentBlocks", "obj", obj.ID, "blockID", blockID, "e", err)
	if err != nil {
		log.Error("increateMessage: Unable to SplitContentBlocks", "e", err)
		return err
	}

	blockInfo, err := pkgservice.NewBlockInfo(blockID, blockHashs, data.MediaIDs, obj.CreatorID)
	if err != nil {
		return err
	}
	blockInfo.SetIsAllGood()

	theObj.SetBlockInfo(blockInfo)

	// op-data
	opData.BlockInfoID = blockID
	opData.NBlock = blockInfo.NBlock
	opData.Hashs = blockHashs
	opData.MediaIDs = data.MediaIDs

	return nil
}

func (pm *ProtocolManager) postcreateMessage(theObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {

	log.Debug("postcreateMessage: start")

	entity := pm.Entity().(*Friend)
	entity.SaveMessageCreateTS(oplog.UpdateTS)

	myID := pm.Ptt().GetMyEntity().GetID()
	creatorID := theObj.GetCreatorID()

	if reflect.DeepEqual(myID, creatorID) {
		pm.SaveLastSeen(oplog.UpdateTS)
	}

	return nil
}
