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
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) GetArticleBlockList(articleID *types.PttID, subContentID *types.PttID, contentType ContentType, blockID uint32, limit int, listOrder pttdb.ListOrder) ([]*ArticleBlock, error) {

	blocks := make([]*ArticleBlock, 0)
	if contentType == ContentTypeArticle {
		articleBlocks, nBlocks, err := pm.getArticleBlockListMainBlocks(articleID, subContentID, blockID, limit)
		if err != nil {
			return articleBlocks, err
		}
		limit -= nBlocks

		subContentID = nil
		contentType = ContentTypeComment
		blockID = 0

		blocks = articleBlocks
	}

	if limit <= 0 {
		return blocks, nil
	}

	commentAndReplyBlocks, _, err := pm.getArticleBlockListCommentAndReplyBlocks(articleID, subContentID, contentType, limit, listOrder)

	if err != nil && err != ErrNotFound {
		return nil, err
	}

	blocks = append(blocks, commentAndReplyBlocks...)

	return blocks, nil
}

func (pm *ProtocolManager) getArticleBlockListMainBlocks(articleID *types.PttID, subContentID *types.PttID, blockID uint32, limit int) ([]*ArticleBlock, int, error) {

	article := NewEmptyArticle()
	pm.SetArticleDB(article)
	article.SetID(articleID)

	err := article.GetByID(false)
	if err != nil {
		return nil, 0, err
	}

	blockInfo := article.GetBlockInfo()
	if blockInfo == nil {
		return nil, 0, pkgservice.ErrInvalidBlock
	}
	pm.SetBlockInfoDB(blockInfo, articleID)

	contentBlockList, err := pkgservice.GetContentBlockList(blockInfo, uint32(limit), false)
	log.Debug("getArticleBlockListMainBlocks: after GetBlockList", "err", err)
	if err != nil {
		return nil, 0, err
	}

	nBlock := len(contentBlockList)

	articleBlocks := make([]*ArticleBlock, nBlock)
	for i, contentBlock := range contentBlockList {
		articleBlocks[i] = contentBlockToArticleBlock(article, blockInfo, contentBlock, article.ID, ContentTypeArticle, CommentTypeNone)
	}

	return articleBlocks, nBlock, nil
}

func (pm *ProtocolManager) getArticleBlockListCommentAndReplyBlocks(articleID *types.PttID, subContentID *types.PttID, contentType ContentType, limit int, listOrder pttdb.ListOrder) ([]*ArticleBlock, int, error) {

	comment := NewEmptyComment()
	pm.SetCommentDB(comment)
	iter, err := comment.GetCrossObjIterWithObj(articleID[:], subContentID, listOrder, false)
	if err != nil {
		return nil, 0, err
	}
	defer iter.Release()

	iterFunc := pttdb.GetFuncIter(iter, listOrder)

	articleBlocks := make([]*ArticleBlock, 0)
	nBlock := 0
	var eachArticleBlock *ArticleBlock
	for iterFunc() {
		if limit > 0 && nBlock >= limit {
			break
		}

		v := iter.Value()
		eachComment := &Comment{}
		err = eachComment.Unmarshal(v)
		if err != nil {
			continue
		}

		eachArticleBlock, err = commentToArticleBlock(pm, eachComment)
		if err != nil {
			continue
		}
		articleBlocks = append(articleBlocks, eachArticleBlock)

		nBlock++

	}

	return articleBlocks, nBlock, nil
}
