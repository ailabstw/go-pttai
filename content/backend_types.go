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

type BackendCreateBoard struct {
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	UpdateTS  types.Timestamp `json:"UT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`

	Status types.Status `json:"S"`

	Title []byte `json:"T"`

	BoardType pkgservice.EntityType `json:"BT"`
}

func boardToBackendCreateBoard(board *Board) *BackendCreateBoard {
	return &BackendCreateBoard{
		ID:        board.ID,
		CreateTS:  board.CreateTS,
		UpdateTS:  board.UpdateTS,
		CreatorID: board.CreatorID,
		UpdaterID: board.UpdaterID,
		Status:    board.Status,
		Title:     board.Title,
		BoardType: board.EntityType,
	}
}

type BackendCreateArticle struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	ContentBlockID *types.PttID `json:"cID"`
	NBlock         int          `json:"NB"`
}

func articleToBackendCreateArticle(a *Article) *BackendCreateArticle {
	blockInfo := a.GetBlockInfo()
	var blockInfoID *types.PttID
	var nBlock int

	if blockInfo != nil {
		blockInfoID = blockInfo.ID
		nBlock = blockInfo.NBlock
	}

	return &BackendCreateArticle{
		BoardID:        a.EntityID,
		ArticleID:      a.ID,
		ContentBlockID: blockInfoID,
		NBlock:         nBlock,
	}
}

type BackendCreateComment struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	CommentID      *types.PttID `json:"CID"`
	ContentBlockID *types.PttID `json:"cID"`
}

func commentToBackendCreateComment(c *Comment) *BackendCreateComment {
	blockInfo := c.GetBlockInfo()
	var blockInfoID *types.PttID
	if blockInfo != nil {
		blockInfoID = blockInfo.ID
	}

	return &BackendCreateComment{
		BoardID:        c.EntityID,
		ArticleID:      c.ArticleID,
		CommentID:      c.ID,
		ContentBlockID: blockInfoID,
	}
}

type BackendCreateReply struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	CommentID      *types.PttID `json:"CID"`
	ReplyID        *types.PttID `json:"RID"`
	ContentBlockID *types.PttID `json:"cID"`
}

type BackendUpdateArticle struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	ContentBlockID *types.PttID `json:"cID"`
	NBlock         int          `json:"NB"`
}

func articleToBackendUpdateArticle(a *Article) *BackendUpdateArticle {
	syncInfo := a.GetSyncInfo()
	blockInfo := a.GetBlockInfo()
	if syncInfo != nil && syncInfo.GetStatus() <= types.StatusAlive {
		blockInfo = syncInfo.GetBlockInfo()
	}

	return &BackendUpdateArticle{
		BoardID:        a.EntityID,
		ArticleID:      a.ID,
		ContentBlockID: blockInfo.ID,
		NBlock:         blockInfo.NBlock,
	}
}

type BackendUpdateReply struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	CommentID      *types.PttID `json:"CID"`
	ReplyID        *types.PttID `json:"RID"`
	ContentBlockID *types.PttID `json:"cID"`
}

type BackendDeleteBoard struct {
}

type BackendDeleteArticle struct {
}

type BackendDeleteComment struct {
}

type BackendDeleteReply struct {
}

type BackendJoinBoard struct {
}

type BackendLeaveBoard struct {
}

type BackendInviteMaster struct {
}

type BackendRevokeMaster struct {
}

type BackendGetBoard struct {
	ID              *types.PttID
	Title           []byte
	Status          types.Status `json:"S"`
	UpdateTS        types.Timestamp
	ArticleCreateTS types.Timestamp
	LastSeen        types.Timestamp
	CreatorID       *types.PttID          `json:"C"`
	BoardType       pkgservice.EntityType `json:"BT"`
}

func boardToBackendGetBoard(b *Board, myName string, theTitle *Title, myID *types.PttID) *BackendGetBoard {
	title := b.Title
	if theTitle != nil {
		title = theTitle.Title
	}

	if len(title) == 0 && b.EntityType == pkgservice.EntityTypePersonal {
		title = DefaultTitle(myID, b.CreatorID, myName)
	}

	articleCreateTS := b.ArticleCreateTS
	if articleCreateTS.IsLess(b.CreateTS) {
		articleCreateTS = b.CreateTS
	}

	return &BackendGetBoard{
		ID:              b.ID,
		Title:           title,
		Status:          b.Status,
		UpdateTS:        b.UpdateTS,
		ArticleCreateTS: b.ArticleCreateTS,
		LastSeen:        b.LastSeen,
		CreatorID:       b.CreatorID,
		BoardType:       b.EntityType,
	}
}

type BackendGetArticle struct {
	ID              *types.PttID
	CreateTS        types.Timestamp //`json:"CT"`
	UpdateTS        types.Timestamp //`json:"UT"`
	CreatorID       *types.PttID    //`json:"CID"`
	BoardID         *types.PttID    //`json:"BID"`
	ContentBlockID  *types.PttID    //`json:"cID"`
	NBlock          int             //`json:"N"`
	NPush           int             `json:"NP"`
	NBoo            int             `json:"NB"`
	Title           []byte          //`json:"T"`
	CommentCreateTS types.Timestamp `json:"c"`
	LastSeen        types.Timestamp `json:"L"`
	Status          types.Status    `json:"S"`
}

func articleToBackendGetArticle(a *Article) *BackendGetArticle {
	nPush := uint64(0)
	if a.NPush != nil {
		nPush = a.NPush.Count()
	}

	nBoo := uint64(0)
	if a.NBoo != nil {
		nBoo = a.NBoo.Count()
	}

	commentCreateTS := a.CommentCreateTS
	if commentCreateTS.IsLess(a.UpdateTS) {
		commentCreateTS = a.UpdateTS
	}

	return &BackendGetArticle{
		ID:              a.ID,
		CreateTS:        a.CreateTS,
		UpdateTS:        a.UpdateTS,
		CreatorID:       a.CreatorID,
		BoardID:         a.EntityID,
		ContentBlockID:  a.BlockInfo.ID,
		NBlock:          a.BlockInfo.NBlock,
		NPush:           int(nPush),
		NBoo:            int(nBoo),
		Title:           a.Title,
		CommentCreateTS: commentCreateTS,
		LastSeen:        a.LastSeen,
		Status:          a.Status,
	}
}

type BackendGetArticleBlock struct {
}

type BackendShowBoardURL struct {
	ID        string
	CreatorID string `json:"C"`
	Title     string `json:"T"`
	Pnode     string `json:"Pn"`
	URL       string
}

type BackendShowArticleURL struct {
	ID  *types.PttID
	BID *types.PttID
	URL string
}

type BackendUploadImg struct {
	ID      *types.PttID
	BoardID *types.PttID         `json:"BID"`
	Type    pkgservice.MediaType `json:"T"`
}

func mediaToBackendUploadImg(img *pkgservice.Media) *BackendUploadImg {
	return &BackendUploadImg{
		ID:      img.ID,
		BoardID: img.EntityID,
		Type:    img.MediaType,
	}
}

type BackendGetImg struct {
	ID      *types.PttID
	BoardID *types.PttID         `json:"BID"`
	Type    pkgservice.MediaType `json:"T"`
	Buf     []byte               `json:"B"`
}

func mediaToBackendGetImg(img *pkgservice.Media) *BackendGetImg {
	return &BackendGetImg{
		ID:      img.ID,
		BoardID: img.EntityID,
		Type:    img.MediaType,
		Buf:     img.Buf,
	}
}

type BackendUploadFile struct {
	ID      *types.PttID
	BoardID *types.PttID `json:"BID"`
}

func mediaToBackendUploadFile(media *pkgservice.Media) *BackendUploadFile {
	return &BackendUploadFile{
		ID:      media.ID,
		BoardID: media.EntityID,
	}
}

type BackendGetFile struct {
	ID        *types.PttID
	BoardID   *types.PttID         `json:"BID"`
	MediaType pkgservice.MediaType `json:"M"`
	Buf       []byte               `json:"B"`
}

func mediaToBackendGetFile(media *pkgservice.Media) *BackendGetFile {
	return &BackendGetFile{
		ID:        media.ID,
		BoardID:   media.EntityID,
		Buf:       media.Buf,
		MediaType: media.MediaType,
	}
}

type BackendArticleSummaryParams struct {
	ArticleID      string `json:"A"`
	ContentBlockID string `json:"B"`
}
