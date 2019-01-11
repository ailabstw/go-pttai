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

func (spm *ServiceProtocolManager) CreateProfile() (*Profile, error) {
	entity, err := spm.CreateEntity(
		nil,
		UserOpTypeCreateProfile,

		spm.NewProfile,
		spm.NewUserOplogWithTS,

		nil,

		nil,
	)

	if err != nil {
		return nil, err
	}

	profile, ok := entity.(*Profile)
	if !ok {
		return nil, pkgservice.ErrInvalidEntity
	}

	pm := profile.PM().(*ProtocolManager)
	err = pm.CreateUserName(nil)
	if err != nil {
		return nil, err
	}

	err = pm.CreateUserImg()
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (spm *ServiceProtocolManager) NewProfile(data pkgservice.CreateData, ptt pkgservice.Ptt, service pkgservice.Service) (pkgservice.Entity, pkgservice.OpData, error) {
	myID := spm.Ptt().GetMyEntity().GetID()

	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}

	profile, err := NewProfile(myID, ts, ptt, service, spm, spm.GetDBLock())
	if err != nil {
		return nil, nil, err
	}

	return profile, &UserOpCreateProfile{}, nil
}
