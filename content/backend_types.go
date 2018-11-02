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

import "github.com/ailabstw/go-pttai/common/types"

type BackendCreateBoard struct {
}

type BackendCreateArticle struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	ContentBlockID *types.PttID `json:"cID"`
	NBlock         int          `json:"NB"`
}

type BackendCreateComment struct {
	BoardID        *types.PttID `json:"BID"`
	ArticleID      *types.PttID `json:"AID"`
	CommentID      *types.PttID `json:"CID"`
	ContentBlockID *types.PttID `json:"cID"`
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
	CreatorID       *types.PttID `json:"C"`
	BoardType       BoardType    `json:"BT"`
}

func boardToBackendGetBoard(b *Board, myName string) *BackendGetBoard {
	title := b.Title

	myID := b.Ptt().GetMyEntity().GetID()
	if len(title) == 0 && b.BoardType == BoardTypePersonal {
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
		BoardType:       b.BoardType,
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
		BoardID:         a.BoardID,
		ContentBlockID:  a.ContentBlockID,
		NBlock:          a.NBlock,
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
	BoardID *types.PttID `json:"BID"`
	Type    MediaType    `json:"T"`
}

func imgToBackendUploadImg(img *Media) *BackendUploadImg {
	return &BackendUploadImg{
		ID:      img.ID,
		BoardID: img.BoardID,
		Type:    img.MediaType,
	}
}

type BackendGetImg struct {
	ID      *types.PttID
	BoardID *types.PttID `json:"BID"`
	Type    MediaType    `json:"T"`
	Buf     []byte       `json:"B"`
}

func imgToBackendGetImg(img *Media) *BackendGetImg {
	return &BackendGetImg{
		ID:      img.ID,
		BoardID: img.BoardID,
		Type:    img.MediaType,
		Buf:     img.Buf,
	}
}

type BackendUploadFile struct {
	ID      *types.PttID
	BoardID *types.PttID `json:"BID"`
}

func mediaToBackendUploadFile(media *Media) *BackendUploadFile {
	return &BackendUploadFile{
		ID:      media.ID,
		BoardID: media.BoardID,
	}
}

type BackendGetFile struct {
	ID      *types.PttID
	BoardID *types.PttID `json:"BID"`
	Buf     []byte       `json:"B"`
}

func mediaToBackendGetFile(media *Media) *BackendGetFile {
	return &BackendGetFile{
		ID:      media.ID,
		BoardID: media.BoardID,
		Buf:     media.Buf,
	}
}

type BackendArticleSummaryParams struct {
	ArticleID      string `json:"A"`
	ContentBlockID string `json:"B"`
}
