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

type ArticleBlock struct {
	V           types.Version
	BlockInfoID *types.PttID `json:"ID"`
	ArticleID   *types.PttID `json:"AID"`
	RefID       *types.PttID `json:"RID"`
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
