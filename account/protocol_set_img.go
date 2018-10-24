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

import (
	"encoding/base64"
	"strings"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func (spm *ServiceProtocolManager) SetImg(ts types.Timestamp, userID *types.PttID, imgStr string, boardID *types.PttID, oplogID *types.PttID, status types.Status) (*UserImg, error) {

	newImgType, newImgWidth, newImgHeight, newImgStr, err := spm.normalizeImg(imgStr)
	if err != nil {
		return nil, err
	}

	u := &UserImg{ID: userID}

	err = u.Get(userID, true)
	if err == leveldb.ErrNotFound {
		err = nil
		u, err = NewUserImg(userID, ts)
		u.BoardID = boardID
		u.Status = types.StatusInit
	}
	if err != nil {
		return nil, err
	}

	if status == types.StatusAlive {
		u.ImgType = newImgType
		u.Width = newImgWidth
		u.Height = newImgHeight
		u.Str = newImgStr
		u.UpdateTS = ts
		u.BoardID = boardID
		u.LogID = oplogID
		u.Status = status
		u.SyncImgInfo = nil
	} else {
		u.IntegrateSyncImgInfo(&SyncImgInfo{
			LogID:    oplogID,
			ImgType:  newImgType,
			Width:    newImgWidth,
			Height:   newImgHeight,
			BoardID:  boardID,
			Str:      newImgStr,
			UpdateTS: ts,
			Status:   status,
		})
	}

	err = u.Save(true)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (spm *ServiceProtocolManager) normalizeImg(str string) (ImgType, uint16, uint16, string, error) {
	imgStrs := strings.SplitN(str, ";", 2) // data:image/png;
	if len(imgStrs) < 2 {
		return ImgTypeJPEG, 0, 0, "", ErrInvalidImg
	}
	imgStr := imgStrs[1]
	imgStrs = strings.SplitN(imgStr, ",", 2) // base64,
	if len(imgStr) < 2 {
		return ImgTypeJPEG, 0, 0, "", ErrInvalidImg
	}

	imgStr = strings.TrimSpace(imgStrs[1])
	imgBuf, err := base64.StdEncoding.DecodeString(imgStr)
	if err != nil {
		return ImgTypeJPEG, 0, 0, "", ErrInvalidImg
	}

	newImgType, newImgWidth, newImgHeight, newBuf, err := NormalizeImage(imgBuf, MaxProfileImgWidth, MaxProfileImgHeight)
	if err != nil {
		return ImgTypeJPEG, 0, 0, "", ErrInvalidImg
	}

	if len(newBuf) > MaxProfileImgSize {
		return ImgTypeJPEG, 0, 0, "", ErrInvalidImg
	}

	newBufStr := base64.StdEncoding.EncodeToString(newBuf)
	newFormatString := ""
	switch newImgType {
	case ImgTypeJPEG:
		newFormatString = "image/jpg"
	case ImgTypeGIF:
		newFormatString = "image/gif"
	case ImgTypePNG:
		newFormatString = "image/png"
	}

	newStr := "data:" + newFormatString + ";base64," + newBufStr

	return newImgType, newImgWidth, newImgHeight, newStr, nil
}
