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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type LocaleInfo struct {
	DefaultName  string
	DefaultTitle func(myID *types.PttID, creatorID *types.PttID, myName string) []byte
}

var (
	localeInfos []*LocaleInfo
)

func InitLocaleInfo() {
	localeInfos = make([]*LocaleInfo, pkgservice.NLocale)

	localeInfos[pkgservice.LocaleTW] = &LocaleInfo{
		DefaultName:  "沒有人",
		DefaultTitle: DefaultTitleTW,
	}

	localeInfos[pkgservice.LocaleHK] = &LocaleInfo{
		DefaultName:  "沒有人",
		DefaultTitle: DefaultTitleHK,
	}

	localeInfos[pkgservice.LocaleCN] = &LocaleInfo{
		DefaultName:  "沒有人",
		DefaultTitle: DefaultTitleCN,
	}

	localeInfos[pkgservice.LocaleEN] = &LocaleInfo{
		DefaultName:  "No One",
		DefaultTitle: DefaultTitleEN,
	}
}

func DefaultTitleCN(myID *types.PttID, creatorID *types.PttID, myName string) []byte {
	var title []byte = nil

	if myName == "" {
		myName = localeInfos[pkgservice.LocaleCN].DefaultName
	}

	if reflect.DeepEqual(myID, creatorID) {
		title = []byte("厉害了 我的板")
	} else {
		title = []byte("厉害了 " + myName + "的板")
	}

	return title
}

func DefaultTitleTW(myID *types.PttID, creatorID *types.PttID, myName string) []byte {
	var title []byte = nil

	if myName == "" {
		myName = localeInfos[pkgservice.LocaleTW].DefaultName
	}

	if reflect.DeepEqual(myID, creatorID) {
		title = []byte("厲害了 我的板")
	} else {
		title = []byte("厲害了 " + myName + "的板")
	}

	return title
}

func DefaultTitleHK(myID *types.PttID, creatorID *types.PttID, myName string) []byte {
	var title []byte = nil

	if myName == "" {
		myName = localeInfos[pkgservice.LocaleHK].DefaultName
	}

	if reflect.DeepEqual(myID, creatorID) {
		title = []byte("猴腮嘞 我嘅板")
	} else {
		title = []byte("猴腮嘞 " + myName + "嘅板")
	}

	return title
}

func DefaultTitleEN(myID *types.PttID, creatorID *types.PttID, myName string) []byte {
	var title []byte = nil

	if myName == "" {
		myName = localeInfos[pkgservice.LocaleEN].DefaultName
	}

	if reflect.DeepEqual(myID, creatorID) {
		title = []byte("Amazing My Board")
	} else {
		title = []byte("Amazing " + myName + "'s Board")
	}

	return title
}
