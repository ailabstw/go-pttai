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

package account

import "github.com/ailabstw/go-pttai/pttdb"

type PrivateAPI struct {
	b *Backend
}

func NewPrivateAPI(b *Backend) *PrivateAPI {
	return &PrivateAPI{b}
}

func (api *PrivateAPI) GetRawUserName(idStr string) (*UserName, error) {
	return api.b.GetRawUserName([]byte(idStr))
}

func (api *PrivateAPI) GetRawUserImg(idStr string) (*UserImg, error) {
	return api.b.GetRawUserImg([]byte(idStr))
}

type PublicAPI struct {
	b *Backend
}

func NewPublicAPI(b *Backend) *PublicAPI {
	return &PublicAPI{b}
}

func (api *PublicAPI) GetUserName(idStr string) (*BackendUserName, error) {
	return api.b.GetUserName([]byte(idStr))
}

func (api *PublicAPI) GetUserNameList(idStr string, limit int, listOrder pttdb.ListOrder) ([]*BackendUserName, error) {
	return api.b.GetUserNameList([]byte(idStr), limit, listOrder)
}

func (api *PublicAPI) GetUserNameByIDs(idStrs []string) (map[string]*BackendUserName, error) {
	idByteList := make([][]byte, len(idStrs))
	for i, idStr := range idStrs {
		idByteList[i] = []byte(idStr)
	}
	return api.b.GetUserNameByIDs(idByteList)
}

func (api *PublicAPI) GetUserImg(idStr string) (*BackendUserImg, error) {
	return api.b.GetUserImg([]byte(idStr))
}

func (api *PublicAPI) GetUserImgByIDs(idStrs []string) (map[string]*BackendUserImg, error) {
	idByteList := make([][]byte, len(idStrs))
	for i, idStr := range idStrs {
		idByteList[i] = []byte(idStr)
	}
	return api.b.GetUserImgByIDs(idByteList)
}
