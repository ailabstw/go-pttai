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

package friend

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

// optype
const (
	FriendOpTypeInvalid pkgservice.OpType = iota

	FriendOpTypeCreateFriend
	FriendOpTypeDeleteFriend

	FriendOpTypeCreateMessage

	FriendOpTypeCreateMedia

	NFriendOpType
)

type FriendOpCreateFriend struct {
	FriendID *types.PttID
}

type FriendOpDeleteFriend struct {
}

type FriendOpCreateMessage struct {
	BlockInfoID *types.PttID `json:"BID"`
	Hashs       [][][]byte   `json:"H"`
	NBlock      int          `json:"NB"`

	MediaIDs []*types.PttID `json:"ms,omitempty"`
}

type FriendOpCreateMedia struct {
	BlockInfoID *types.PttID `json:"BID"` // resized content-block-id
	Hashs       [][][]byte   `json:"H"`
	NBlock      int          `json:"NB"`
}
