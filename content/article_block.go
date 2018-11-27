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

type ArticleBlock struct {
	V           types.Version
	BlockInfoID *types.PttID `json:"bID"`
	ArticleID   *types.PttID `json:"AID"`
	RefID       *types.PttID `json:"ID"`
	ContentType ContentType  `json:"ct"`
	CommentType CommentType  `json:"mt"`
	BlockID     uint32       `json:"BID"`

	Status types.Status `json:"S"`

	CreateTS types.Timestamp `json:"CT"`
	UpdateTS types.Timestamp `json:"UT"`

	CreatorID *types.PttID `json:"CID"`
	UpdaterID *types.PttID `json:"UID"`

	Buf [][]byte `json:"B"`
}

func NewArticleBlock() (*ArticleBlock, error) {
	return &ArticleBlock{}, nil
}

func contentBlockToArticleBlock(obj pkgservice.Object, blockInfo *pkgservice.BlockInfo, contentBlock *pkgservice.ContentBlock, articleID *types.PttID, contentType ContentType, commentType CommentType) *ArticleBlock {
	return &ArticleBlock{
		V:           types.CurrentVersion,
		BlockInfoID: blockInfo.ID,
		ArticleID:   articleID,
		RefID:       obj.GetID(),
		ContentType: contentType,
		CommentType: commentType,
		BlockID:     contentBlock.BlockID,

		Status: obj.GetStatus(),

		CreateTS: obj.GetCreateTS(),
		UpdateTS: obj.GetUpdateTS(),

		CreatorID: obj.GetCreatorID(),
		UpdaterID: obj.GetUpdaterID(),

		Buf: contentBlock.Buf,
	}
}

func commentToArticleBlock(pm *ProtocolManager, comment *Comment) (*ArticleBlock, error) {
	articleBlock := &ArticleBlock{
		V:           types.CurrentVersion,
		BlockInfoID: nil,
		ArticleID:   comment.ArticleID,
		RefID:       comment.ID,
		ContentType: ContentTypeComment,
		CommentType: comment.CommentType,
		BlockID:     0,

		Status: comment.Status,

		CreateTS:  comment.CreateTS,
		UpdateTS:  comment.UpdateTS,
		CreatorID: comment.CreatorID,
		UpdaterID: comment.UpdaterID,
	}

	if comment.Status > types.StatusAlive {
		articleBlock.Buf = DefaultDeletedComment
		return articleBlock, nil
	}

	blockInfo := comment.GetBlockInfo()
	if blockInfo == nil {
		return nil, pkgservice.ErrInvalidBlock
	}
	pm.SetBlockInfoDB(blockInfo, comment.ID)

	contentBlockList, err := pkgservice.GetContentBlockList(blockInfo, uint32(1), false)
	log.Debug("commentToArticleBlock: after GetContentBlockList", "err", err)
	if err != nil {
		return nil, err
	}

	if len(contentBlockList) != 1 {
		return nil, ErrInvalidBlock
	}

	contentBlock := contentBlockList[0]

	articleBlock.BlockInfoID = blockInfo.ID
	articleBlock.Buf = contentBlock.Buf

	return articleBlock, nil
}
