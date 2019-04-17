// Copyright 2019 The go-pttai Authors
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

	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) CleanObject() error {
	article := NewEmptyArticle()
	pm.SetArticleDB(article)

	comment := NewEmptyComment()
	pm.SetCommentDB(comment)

	media := pkgservice.NewEmptyMedia()
	pm.SetMediaDB(media)

	// article
	iter, err := article.GetObjIterWithObj(nil, pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	var val []byte
	for iter.Next() {
		val = iter.Value()

		err = json.Unmarshal(val, article)
		if err != nil {
			continue
		}
		pm.SetArticleDB(article)

		article.DeleteAll(comment, false)
	}

	// comment
	iter, err = comment.GetObjIterWithObj(nil, pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	for iter.Next() {
		val = iter.Value()

		err = json.Unmarshal(val, comment)
		if err != nil {
			continue
		}
		pm.SetCommentDB(comment)

		comment.DeleteAll(false)
	}

	// media
	iter, err = media.GetObjIterWithObj(nil, pttdb.ListOrderNext, false)
	if err != nil {
		return err
	}
	defer iter.Release()

	for iter.Next() {
		val = iter.Value()

		err = json.Unmarshal(val, media)
		if err != nil {
			continue
		}
		pm.SetMediaDB(media)

		media.DeleteAll(false)
	}

	return nil
}
