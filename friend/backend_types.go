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
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
)

/*
BackendGetFriend provides just UserName and nothing else for saving the bandwidth
*/
type BackendGetFriend struct {
	ID       *types.PttID
	FriendID *types.PttID `json:"FID"`
	Name     []byte       `json:"N"`

	BoardID         *types.PttID    `json:"BID"`
	Status          types.Status    `json:"S"`
	ArticleCreateTS types.Timestamp //`json:"ACT"`
	LastSeen        types.Timestamp `json:"LT"`
}

func friendToBackendGetFriend(f *Friend, userName *account.UserName) *BackendGetFriend {
	return &BackendGetFriend{
		ID:              f.ID,
		FriendID:        f.FriendID,
		Name:            userName.Name,
		Status:          f.Status,
		BoardID:         f.BoardID,
		ArticleCreateTS: f.ArticleCreateTS,
		LastSeen:        f.LastSeen,
	}
}

type BackendCreateMessage struct {
	FriendID       *types.PttID `json:"FID"`
	ArticleID      *types.PttID `json:"AID"`
	ContentBlockID *types.PttID `json:"cID"`
	NBlock         int          `json:"NB"`
}

type BackendGetMessage struct {
	ID             *types.PttID
	CreateTS       types.Timestamp //`json:"CT"`
	UpdateTS       types.Timestamp //`json:"UT"`
	CreatorID      *types.PttID    //`json:"CID"`
	FriendID       *types.PttID    //`json:"FID"`
	ContentBlockID *types.PttID    //`json:"cID"`
	NBlock         int             //`json:"N"`
	Status         types.Status    `json:"S"`
}

func articleToBackendGetMessage(a *Message) *BackendGetMessage {
	return &BackendGetMessage{
		ID:             a.ID,
		CreateTS:       a.CreateTS,
		UpdateTS:       a.UpdateTS,
		CreatorID:      a.CreatorID,
		ContentBlockID: a.BlockInfo.ID,
		NBlock:         a.BlockInfo.NBlock,
		Status:         a.Status,
	}
}
