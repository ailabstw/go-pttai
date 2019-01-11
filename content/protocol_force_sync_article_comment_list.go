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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ForceSyncArticleCommentList struct {
	IDs []*types.PttID
}

func (pm *ProtocolManager) ForceSyncArticleCommentList(
	ids []*types.PttID,
	peer *pkgservice.PttPeer,
) error {

	if len(ids) == 0 {
		return nil
	}

	pIDs := ids
	var eachIDs []*types.PttID
	lenEachIDs := 0
	var data *ForceSyncArticleCommentList
	for len(pIDs) > 0 {
		lenEachIDs = pkgservice.MaxSyncObjectAck
		if lenEachIDs > len(pIDs) {
			lenEachIDs = len(pIDs)
		}

		eachIDs, pIDs = pIDs[:lenEachIDs], pIDs[lenEachIDs:]

		data = &ForceSyncArticleCommentList{
			IDs: eachIDs,
		}

		err := pm.SendDataToPeer(ForceSyncArticleCommentMsg, data, peer)
		if err != nil {
			return err
		}

	}
	return nil
}

func (pm *ProtocolManager) HandleForceSyncAricleCommentList(
	dataBytes []byte,
	peer *pkgservice.PttPeer,

) error {

	data := &ForceSyncArticleCommentList{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	comments := make([]pkgservice.Object, 0, pkgservice.MaxSyncObjectAck)
	var eachComments []pkgservice.Object
	for _, id := range data.IDs {
		eachComments, err = pm.handleForceSyncArticleCommentListGetComments(id)
		if err != nil {
			continue
		}

		if len(comments)+len(eachComments) > pkgservice.MaxSyncObjectAck {
			pm.SyncObjectAck(comments, ForceSyncCommentAckMsg, peer)
			comments = make([]pkgservice.Object, 0, pkgservice.MaxSyncObjectAck)
		}
		comments = append(comments, eachComments...)
	}

	if len(comments) != 0 {
		pm.SyncObjectAck(comments, ForceSyncCommentAckMsg, peer)
	}

	return nil
}

func (pm *ProtocolManager) handleForceSyncArticleCommentListGetComments(articleID *types.PttID) ([]pkgservice.Object, error) {

	comment := NewEmptyComment()
	pm.SetCommentDB(comment)
	iter, err := comment.GetCrossObjIterWithObj(articleID[:], nil, pttdb.ListOrderNext, false)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	iterFunc := pttdb.GetFuncIter(iter, pttdb.ListOrderNext)

	comments := make([]pkgservice.Object, 0)
	var eachComment *Comment
	for iterFunc() {

		v := iter.Value()
		eachComment = &Comment{}
		err = eachComment.Unmarshal(v)
		if err != nil {
			continue
		}

		comments = append(comments, eachComment)
	}

	return comments, nil
}
