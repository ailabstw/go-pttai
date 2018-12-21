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
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (spm *ServiceProtocolManager) CreateJoinEntity(
	approveJoin *ApproveJoinEntity,
	peer *pkgservice.PttPeer,

	meLog *pkgservice.BaseOplog,

	isStart bool,
	isNew bool,
	isForceNotBroadcast bool,

	isLocked bool,
) (pkgservice.Entity, error) {

	entity, err := spm.BaseServiceProtocolManager.CreateJoinEntity(
		approveJoin.ApproveJoinEntity,
		peer,

		meLog,

		isStart,
		isNew,
		isForceNotBroadcast,

		isLocked,
	)
	if err != nil {
		return nil, err
	}

	pm := entity.PM().(*ProtocolManager)

	// user name
	err = pm.createJoinEntityUserName(approveJoin.UserName)
	if err != nil {
		return nil, err
	}

	// user img
	err = pm.createJoinEntityUserImg(approveJoin.UserImg)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (pm *ProtocolManager) createJoinEntityUserName(userName *UserName) error {
	pm.SetUserNameDB(userName)

	err := userName.Lock()
	if err != nil {
		return err
	}
	defer userName.Unlock()

	origUserName := NewEmptyUserName()
	pm.SetUserNameDB(origUserName)
	origUserName.SetID(userName.GetID())

	err = origUserName.GetByID(true)
	if err == nil {
		if userName.UpdateTS.IsLess(origUserName.UpdateTS) {
			return nil
		}
	}

	userName.Status = types.StatusInternalSync

	userName.Save(true)

	return nil
}

func (pm *ProtocolManager) createJoinEntityUserImg(userImg *UserImg) error {
	pm.SetUserImgDB(userImg)

	err := userImg.Lock()
	if err != nil {
		return err
	}
	defer userImg.Unlock()

	origUserImg := NewEmptyUserImg()
	pm.SetUserImgDB(origUserImg)
	origUserImg.SetID(userImg.GetID())

	err = origUserImg.GetByID(true)
	if err == nil {
		if userImg.UpdateTS.IsLess(origUserImg.UpdateTS) {
			return nil
		}
	}

	userImg.Status = types.StatusInternalSync

	userImg.Save(true)

	return nil
}
