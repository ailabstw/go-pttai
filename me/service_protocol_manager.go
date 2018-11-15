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

package me

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ServiceProtocolManager struct {
	*pkgservice.BaseServiceProtocolManager

	MyID   *types.PttID
	MyInfo *MyInfo

	myPtt pkgservice.MyPtt
}

func NewServiceProtocolManager(myID *types.PttID, ptt pkgservice.MyPtt, service pkgservice.Service, contentBackend *content.Backend) (*ServiceProtocolManager, error) {

	spm := &ServiceProtocolManager{myPtt: ptt, MyID: myID}
	b, err := pkgservice.NewBaseServiceProtocolManager(ptt, service)
	if err != nil {
		return nil, err
	}

	spm.BaseServiceProtocolManager = b

	// load me
	myInfo, myInfos, err := spm.GetMeList(myID, contentBackend, nil, 0)
	if err != nil {
		return nil, err
	}

	for _, eachMyInfo := range myInfos {
		err = eachMyInfo.Init(ptt, service, spm)
		if err != nil {
			log.Error("Unable to Init", "myInfo", eachMyInfo.GetID(), "e", err)
			return nil, err
		}

		err = spm.RegisterEntity(eachMyInfo.ID, eachMyInfo)
		if err != nil {
			return nil, err
		}

	}

	// Me
	spm.MyInfo = myInfo

	return spm, nil
}

func (spm *ServiceProtocolManager) NewEmptyEntity() pkgservice.Entity {
	return NewEmptyMyInfo()
}
