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
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) GetMessageList(startID *types.PttID, limit int, listOrder pttdb.ListOrder, isLocked bool) ([]*Message, error) {
	obj := NewEmptyMessage()
	pm.SetMessageDB(obj)

	objs, err := pkgservice.GetObjList(obj, startID, limit, listOrder, isLocked)
	if err != nil {
		return nil, err
	}
	typedObjs := ObjsToMessages(objs)

	return typedObjs, nil
}
