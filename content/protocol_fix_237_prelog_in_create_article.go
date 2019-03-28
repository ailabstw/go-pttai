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
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pm *ProtocolManager) Fix237PrelogInCreateArticle() error {
	isFixed, err := pm.isFix237PrelogInCreateArticle()
	if err != nil {
		return err
	}
	if isFixed {
		return nil
	}

	err = pm.fix237PrelogInCreateArticleCore()
	if err != nil {
		return err
	}

	err = pm.setFix237PrelogInCreateArticle()
	if err != nil {
		return err
	}

	return nil

}

func (pm *ProtocolManager) isFix237PrelogInCreateArticle() (bool, error) {
	key, err := pm.marshalFix237PrelogInCreateArticleKey()
	if err != nil {
		return false, err
	}
	_, err = dbMeta.Get(key)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pm *ProtocolManager) marshalFix237PrelogInCreateArticleKey() ([]byte, error) {
	entityID := pm.Entity().GetID()
	return common.Concat([][]byte{DBFix237Prefix, entityID[:]})
}

func (pm *ProtocolManager) setFix237PrelogInCreateArticle() error {
	key, err := pm.marshalFix237PrelogInCreateArticleKey()
	if err != nil {
		return err
	}

	dbMeta.Put(key, pttdb.ValueTrue)

	return nil
}

func (pm *ProtocolManager) fix237PrelogInCreateArticleCore() error {
	oplogs, err := pm.GetBoardOplogList(nil, 0, pttdb.ListOrderNext, types.StatusAlive)
	if err != nil {
		return err
	}

	for _, oplog := range oplogs {
		pm.SetBoardDB(oplog.BaseOplog)
		if !fix237PreLogInCreateArticleCoreIsValidOp(oplog) {
			continue
		}
		if oplog.PreLogID == nil {
			continue
		}

		log.Warn("fix237PrelogInCreateArticleCore: PreLogID in CreateArticle (alive)", "ID", oplog.ID, "objID", oplog.ObjID, "entity", pm.Entity().IDString())

		oplog.PreLogID = nil

		oplog.Save(false, pm.boardOplogMerkle)
	}

	oplogs, err = pm.GetBoardOplogList(nil, 0, pttdb.ListOrderNext, types.StatusPending)
	if err != nil {
		return err
	}

	for _, oplog := range oplogs {
		if !fix237PreLogInCreateArticleCoreIsValidOp(oplog) {
			continue
		}
		if oplog.PreLogID == nil {
			continue
		}

		log.Warn("fix237PrelogInCreateArticleCore: PreLogID in CreateArticle (pending)", "ID", oplog.ID, "objID", oplog.ObjID, "entity", pm.Entity().IDString())

		oplog.PreLogID = nil

		oplog.Save(false, pm.boardOplogMerkle)
	}

	oplogs, err = pm.GetBoardOplogList(nil, 0, pttdb.ListOrderNext, types.StatusInternalPending)
	if err != nil {
		return err
	}

	for _, oplog := range oplogs {
		if !fix237PreLogInCreateArticleCoreIsValidOp(oplog) {
			continue
		}
		if oplog.PreLogID == nil {
			continue
		}

		log.Warn("fix237PrelogInCreateArticleCore: PreLogID in CreateArticle (internal)", "ID", oplog.ID, "objID", oplog.ObjID, "entity", pm.Entity().IDString())

		oplog.PreLogID = nil

		oplog.Save(false, pm.boardOplogMerkle)
	}

	return nil
}

func fix237PreLogInCreateArticleCoreIsValidOp(oplog *BoardOplog) bool {
	switch oplog.Op {
	case BoardOpTypeCreateArticle:
		return true
	case BoardOpTypeCreateComment:
		return true
	default:
		return false
	}
}
