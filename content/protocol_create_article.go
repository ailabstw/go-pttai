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
	"reflect"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type CreateArticle struct {
	Title    []byte
	Article  [][]byte
	MediaIDs []*types.PttID
}

func (pm *ProtocolManager) CreateArticle(title []byte, articleBytes [][]byte, mediaIDs []*types.PttID) (*Article, error) {

	myID := pm.Ptt().GetMyEntity().GetID()

	if pm.Entity().GetEntityType() == pkgservice.EntityTypePersonal && !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	data := &CreateArticle{
		Title:    title,
		Article:  articleBytes,
		MediaIDs: mediaIDs,
	}

	theArticle, err := pm.CreateObject(data, BoardOpTypeCreateArticle, pm.NewArticle, pm.NewBoardOplogWithTS, pm.increateArticle, pm.broadcastBoardOplogCore, pm.postcreateArticle)
	if err != nil {
		return nil, err
	}

	article, ok := theArticle.(*Article)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	return article, nil
}

func (pm *ProtocolManager) NewArticle(theData pkgservice.CreateData) (pkgservice.Object, pkgservice.OpData, error) {

	data, ok := theData.(*CreateArticle)
	if !ok {
		return nil, nil, pkgservice.ErrInvalidData
	}

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	opData := &BoardOpCreateArticle{}

	theArticle, err := NewArticle(ts, myID, entityID, nil, types.StatusInit, data.Title)
	if err != nil {
		return nil, nil, err
	}
	pm.SetArticleDB(theArticle)

	return theArticle, opData, nil
}

func (pm *ProtocolManager) increateArticle(theObj pkgservice.Object, theData pkgservice.CreateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) error {

	obj, ok := theObj.(*Article)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	data, ok := theData.(*CreateArticle)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*BoardOpCreateArticle)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// block-info
	blockID, blockHashs, err := pm.SplitContentBlocks(nil, obj.ID, data.Article, NFirstLineInBlock)
	log.Debug("increateArticle: after SplitContentBlocks", "obj", obj.ID, "blockID", blockID, "e", err)
	if err != nil {
		log.Error("increateArticle: Unable to SplitContentBlocks", "e", err)
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

	opData.TitleHash = types.Hash(obj.Title)

	return nil
}

func (pm *ProtocolManager) postcreateArticle(theObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {

	log.Debug("postcreateArticle: start")

	article, ok := theObj.(*Article)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	myID := pm.Ptt().GetMyEntity().GetID()

	entity := pm.Entity().(*Board)
	entity.SaveArticleCreateTS(oplog.UpdateTS)
	pm.SaveLastSeen(oplog.UpdateTS)

	// ptt-oplog
	if reflect.DeepEqual(article.CreatorID, myID) {
		return nil
	}

	// I can get only mine and my friends' user name
	accountSPM := pm.Entity().Service().(*Backend).accountBackend.SPM().(*account.ServiceProtocolManager)
	_, err := accountSPM.GetUserNameByID(article.CreatorID)
	if err != nil {
		return nil
	}

	opData := &pkgservice.PttOpCreateArticle{
		BoardID: entity.GetID(),
		Title:   article.Title,
	}

	pttOplog, err := pkgservice.NewPttOplog(article.ID, article.UpdateTS, oplog.CreatorID, pkgservice.PttOpTypeCreateArticle, opData, myID)
	if err != nil {
		return err
	}

	err = pttOplog.Save(false)
	if err != nil {
		return err
	}

	return nil
}
