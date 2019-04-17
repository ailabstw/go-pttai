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

package account

import (
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func NewEmptyApproveJoinProfile() *ApproveJoinEntity {
	return &ApproveJoinEntity{
		ApproveJoinEntity: &pkgservice.ApproveJoinEntity{
			Entity: NewEmptyProfile(),
		},
		UserName: NewEmptyUserName(),
		UserImg:  NewEmptyUserImg(),
		NameCard: NewEmptyNameCard(),
	}
}

type ApproveJoinEntity struct {
	*pkgservice.ApproveJoinEntity `json:"b"`
	UserName                      *UserName `json:"n"`
	UserImg                       *UserImg  `json:"i"`
	NameCard                      *NameCard `json:"c"`
}

func (pm *ProtocolManager) ApproveJoin(
	joinEntity *pkgservice.JoinEntity,
	keyInfo *pkgservice.KeyInfo,
	peer *pkgservice.PttPeer,
) (*pkgservice.KeyInfo, interface{}, error) {

	keyInfo, approveJoinEntity, err := pm.BaseProtocolManager.ApproveJoin(joinEntity, keyInfo, peer)
	if err != nil {
		return nil, nil, err
	}

	userID := pm.Entity().GetCreatorID()
	spm := pm.Entity().Service().SPM().(*ServiceProtocolManager)

	userName, err := spm.GetUserNameByID(userID)
	if err != nil {
		return nil, nil, err
	}

	userImg, err := spm.GetUserImgByID(userID)
	if err != nil {
		return nil, nil, err
	}

	nameCard, err := spm.GetNameCardByID(userID)
	if err != nil {
		return nil, nil, err
	}

	data := &ApproveJoinEntity{
		ApproveJoinEntity: approveJoinEntity.(*pkgservice.ApproveJoinEntity),
		UserName:          userName,
		UserImg:           userImg,
		NameCard:          nameCard,
	}

	return keyInfo, data, nil
}
