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
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) GetMessageBlockList(msgID *types.PttID, limit uint32) (*Message, []*pkgservice.ContentBlock, error) {

	msg := NewEmptyMessage()
	pm.SetMessageDB(msg)
	msg.SetID(msgID)

	err := msg.GetByID(false)
	if err != nil {
		return nil, nil, err
	}

	blockInfo := msg.GetBlockInfo()
	log.Debug("GetMessageBlockList: after GetBlockInfo", "msgID", msgID, "blockInfo", blockInfo)
	if blockInfo == nil {
		return nil, nil, pkgservice.ErrInvalidBlock
	}
	pm.SetBlockInfoDB(blockInfo, msgID)

	contentBlockList, err := pkgservice.GetContentBlockList(blockInfo, limit, false)
	log.Debug("GetMessageBlockList: after GetBlockList", "err", err)
	if err != nil {
		return nil, nil, err
	}

	return msg, contentBlockList, nil
}
