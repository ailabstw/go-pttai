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

type SyncCommentInfo struct {
	*pkgservice.BaseSyncInfo `json:"b"`

	CommentType CommentType `json:"t"`
}

type Comment struct {
	*pkgservice.BaseObject `json:"b"`

	UpdateTS types.Timestamp `json:"UT"`

	ArticleID        *types.PttID `json:"AID"`
	ArticleCreatorID *types.PttID `json:"aID"`

	CommentType CommentType `json:"t"`
}

func NewComment() (*Comment, error) {
	return &Comment{}, nil
}
