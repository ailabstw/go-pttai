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
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type UpdateUserImg struct {
	ImgType ImgType `json:"T"`
	Width   uint16  `json:"W"`
	Height  uint16  `json:"H"`
	Str     string  `json:"I"`
}

func (pm *ProtocolManager) UpdateUserImg(imgStr string) (*UserImg, error) {

	newImgType, newImgWidth, newImgHeight, newImgStr, err := normalizeUserImg(imgStr)
	if err != nil {
		return nil, err
	}

	myID := pm.Ptt().GetMyEntity().GetID()

	if !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	data := &UpdateUserImg{
		ImgType: newImgType,
		Width:   newImgWidth,
		Height:  newImgHeight,
		Str:     newImgStr,
	}

	origObj := NewEmptyUserImg()
	pm.SetUserImgDB(origObj)

	opData := &UserOpUpdateUserImg{}

	err = pm.UpdateObject(
		myID,

		data,
		UserOpTypeUpdateUserImg,

		origObj,

		opData,

		pm.userOplogMerkle,

		pm.SetUserDB,
		pm.NewUserOplog,
		pm.inupdateUserImg,

		nil,

		pm.broadcastUserOplogCore,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return origObj, nil
}

func (pm *ProtocolManager) inupdateUserImg(obj pkgservice.Object, theData pkgservice.UpdateData, oplog *pkgservice.BaseOplog, theOpData pkgservice.OpData) (pkgservice.SyncInfo, error) {

	data, ok := theData.(*UpdateUserImg)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	opData, ok := theOpData.(*UserOpUpdateUserImg)
	if !ok {
		return nil, pkgservice.ErrInvalidData
	}

	// op-data
	opData.Hash = types.Hash([]byte(data.Str))

	// sync-info
	syncInfo := NewEmptySyncUserImgInfo()
	syncInfo.InitWithOplog(oplog.ToStatus(), oplog)

	syncInfo.ImgType = data.ImgType
	syncInfo.Width = data.Width
	syncInfo.Height = data.Height
	syncInfo.Str = data.Str

	return syncInfo, nil
}

func normalizeUserImg(str string) (ImgType, uint16, uint16, string, error) {
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
