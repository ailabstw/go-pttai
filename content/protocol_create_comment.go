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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type CreateComment struct {
	ArticleID   *types.PttID
	CommentType CommentType
	Comment     [][]byte
	MediaIDs    []*types.PttID
}

func (pm *ProtocolManager) CreateComment(articleID *types.PttID, commentType CommentType, commentBytes []byte, mediaID *types.PttID) (*Comment, error) {

	var mediaIDs []*types.PttID
	if mediaID != nil {
		mediaIDs = []*types.PttID{mediaID}
	}
	data := &CreateComment{
		ArticleID:   articleID,
		CommentType: commentType,
		Comment:     [][]byte{commentBytes},
		MediaIDs:    mediaIDs,
	}

	theComment, err := pm.CreateObject(data, BoardOpTypeCreateComment, pm.NewComment, pm.NewBoardOplogWithTS, pm.increateComment, pm.broadcastBoardOplogCore, pm.postcreateComment)
	if err != nil {
		return nil, err
	}

	comment, ok := theComment.(*Comment)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	return comment, nil
}

func (pm *ProtocolManager) NewComment(theData pkgservice.CreateData) (pkgservice.Object, pkgservice.OpData, error) {

	data, ok := theData.(*CreateComment)
	if !ok {
		return nil, nil, pkgservice.ErrInvalidData
	}

	myID := pm.Ptt().GetMyEntity().GetID()
	entityID := pm.Entity().GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	articleID := data.ArticleID

	article := NewEmptyArticle()
	pm.SetArticleDB(article)
	article.SetID(articleID)

	err = article.GetByID(false)
	if err != nil {
		return nil, nil, err
	}

	opData := &BoardOpCreateComment{}

	theComment, err := NewComment(ts, myID, entityID, nil, types.StatusInit, articleID, article.CreatorID, data.CommentType)
	if err != nil {
		return nil, nil, err
	}
	pm.SetCommentDB(theComment)

	return theComment, opData, nil
}

func (pm *ProtocolManager) increateComment(theObj pkgservice.Object, theData pkgservice.CreateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) error {

	obj, ok := theObj.(*Comment)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	data, ok := theData.(*CreateComment)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*BoardOpCreateComment)
	if !ok {
		return pkgservice.ErrInvalidData
	}

	// block-info
	blockID, blockHashs, err := pm.SplitContentBlocks(nil, obj.ID, data.Comment, NFirstLineInBlock)
	if err != nil {
		log.Error("increateComment: Unable to SplitContentBlocks", "e", err)
		return err
	}

	blockInfo, err := pkgservice.NewBlockInfo(blockID, blockHashs, data.MediaIDs, obj.CreatorID)
	if err != nil {
		return err
	}
	blockInfo.SetIsAllGood()

	theObj.SetBlockInfo(blockInfo)

	// op-data
	opData.ArticleID = obj.ArticleID
	opData.BlockInfoID = blockID
	opData.Hashs = blockHashs
	opData.MediaIDs = data.MediaIDs

	return nil
}

func (pm *ProtocolManager) postcreateComment(theObj pkgservice.Object, oplog *pkgservice.BaseOplog) error {

	log.Debug("postcreateComment: start")

	article := NewEmptyArticle()
	pm.SetArticleDB(article)
	article.SetID(oplog.ObjID)

	article.SaveCommentCreateTS(oplog.UpdateTS)
	article.SaveLastSeen(oplog.UpdateTS)

	return nil
}
