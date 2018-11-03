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

package service

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

const (
	_ OpType = iota
	PttOpTypeCreateMe
	PttOpTypeCreateArticle
	PttOpTypeCreateComment
	PttOpTypeCreateReply
)

type PttOpCreateMe struct {
	NodeID *discover.NodeID `json:"ID"`
}

type PttOpCreateArticle struct {
	BoardID *types.PttID `json:"bID"`
	Title   []byte       `json:"T"`
}

type PttOpCreateComment struct {
	BoardID   *types.PttID `json:"bID"`
	ArticleID *types.PttID `json:"aID"`
}

type PttOpCreateReply struct {
	BoardID   *types.PttID `json:"bID"`
	ArticleID *types.PttID `json:"aID"`
	CommentID *types.PttID `json:"cID"`
}
