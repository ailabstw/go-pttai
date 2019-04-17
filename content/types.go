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

package content

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

//content type
type ContentType uint8

const (
	ContentTypeArticle ContentType = iota
	ContentTypeComment
	ContentTypeReply
)

// comment type
type CommentType int

const (
	CommentTypePush CommentType = iota
	CommentTypeBoo
	CommentTypeNone
)

func (c *CommentType) Marshal() []byte {
	theBytes := [1]byte{}
	theBytes[0] = uint8(*c)

	return theBytes[:]
}

type ReplyInfo struct {
	Op pkgservice.OpType

	ArticleID      *types.PttID `json:"AID"`
	CommentID      *types.PttID `json:"CID"`
	ContentBlockID *types.PttID `json:"cID"`
	MediaID        *types.PttID

	Log *BoardOplog
}

type ContentBlockInfo struct {
	ArticleID *types.PttID `json:"AID"`
	RefID     *types.PttID `json:"RID"`
	ID        *types.PttID
	NBlock    int `json:"N"`
}

type ArticleInfo struct {
	Op pkgservice.OpType

	ContentBlockID *types.PttID
	NBlock         int
	MediaIDs       []*types.PttID

	Log *BoardOplog
}

type MediaInfo struct {
	Op pkgservice.OpType

	ContentBlockID     *types.PttID
	NBlock             int
	OrigContentBlockID *types.PttID
	OrigNBlock         int

	Log *BoardOplog
}

type MetaInfo struct {
	Op pkgservice.OpType

	Log *BoardOplog
}

type CommentInfo struct {
	Op pkgservice.OpType

	ArticleID      *types.PttID
	ContentBlockID *types.PttID
	MediaID        *types.PttID

	Log *BoardOplog
}

type SyncID struct {
	ID    *types.PttID
	LogID *types.PttID `json:"l"`
}

type SyncReplyID struct {
	ArticleID *types.PttID `json:"a"`
	CommentID *types.PttID `json:"c"`
	LogID     *types.PttID `json:"l"`
}
