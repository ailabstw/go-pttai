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

import "github.com/ailabstw/go-pttai/common/types"

type BackendUserName struct {
	ID   *types.PttID
	Name []byte `json:"N"`
}

func userNameToBackendUserName(u *UserName) *BackendUserName {
	return &BackendUserName{
		ID:   u.ID,
		Name: u.Name,
	}
}

type BackendUserImg struct {
	ID     *types.PttID
	Type   ImgType `json:"T"`
	Img    string  `json:"I"`
	Width  uint16  `json:"W"`
	Height uint16  `json:"H"`
}

func userImgToBackendUserImg(u *UserImg) *BackendUserImg {
	return &BackendUserImg{
		ID:     u.ID,
		Type:   u.ImgType,
		Img:    u.Str, //XXX TODO: ensure the img
		Width:  u.Width,
		Height: u.Height,
	}
}
