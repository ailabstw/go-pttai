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

const (
	BoardOpTypeInvalid pkgservice.OpType = iota

	BoardOpTypeCreateBoard
	BoardOpTypeDeleteBoard
	BoardOpTypeMigrateBoard

	BoardOpTypeCreateTitle
	BoardOpTypeUpdateTitle

	BoardOpTypeCreateArticle
	BoardOpTypeUpdateArticle
	BoardOpTypeDeleteArticle

	BoardOpTypeCreateMedia
	BoardOpTypeDeleteMedia

	BoardOpTypeCreateComment
	BoardOpTypeDeleteComment

	BoardOpTypeCreateReply
	BoardOpTypeUpdateReply
	BoardOpTypeDeleteReply

	NBoardOpType
)

type BoardOpCreateBoard struct {
	Title []byte `json:"t"`
}

type BoardOpDeleteBoard struct {
}

type BoardOpMigrateBoard struct {
	ToID *types.PttID `json:"t"`
}

type BoardOpCreateTitle struct {
	TitleHash []byte `json:"TH"`
}

type BoardOpUpdateTitle struct {
	TitleHash []byte `json:"TH"`
}

type BoardOpCreateArticle struct {
	BlockInfoID *types.PttID `json:"BID"`
	Hashs       [][][]byte   `json:"H"`
	NBlock      int          `json:"NB"`

	MediaIDs []*types.PttID `json:"ms,omitempty"`

	TitleHash []byte `json:"th"`
}

type BoardOpUpdateArticle struct {
	BlockInfoID *types.PttID `json:"BID"`
	Hashs       [][][]byte   `json:"H"`
	NBlock      int          `json:"NB"`

	MediaIDs []*types.PttID `json:"ms,omitempty"`
}

type BoardOpDeleteArticle struct {
}

type BoardOpCreateComment struct {
	ArticleID *types.PttID `json:"AID"`

	BlockInfoID *types.PttID   `json:"BID"`
	Hashs       [][][]byte     `json:"H"`
	MediaIDs    []*types.PttID `json:"ms,omitempty"`
}

type BoardOpDeleteComment struct {
	ArticleID *types.PttID `json:"AID"`
}

type BoardOpCreateReply struct {
	ArticleID *types.PttID `json:"AID"`
	CommentID *types.PttID `json:"ACD"`

	BlockInfoID *types.PttID   `json:"BID"`
	Hashs       [][][]byte     `json:"H"`
	MediaIDs    []*types.PttID `json:"ms,omitempty"`
}

type BoardOpDeleteReply struct {
	ArticleID *types.PttID `json:"AID"`
	CommentID *types.PttID `json:"ACD"`
}

type BoardOpUpdateReply struct {
	BlockInfoID *types.PttID   `json:"BID"`
	Hashs       [][][]byte     `json:"H"`
	MediaIDs    []*types.PttID `json:"ms,omitempty"`
}
