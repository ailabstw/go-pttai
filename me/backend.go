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
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Backend struct {
	*pkgservice.BaseService

	Config *Config

	accountBackend *account.Backend
	contentBackend *content.Backend
	friendBackend  *friend.Backend
}

func NewBackend(ctx *pkgservice.ServiceContext, cfg *Config, ptt *pkgservice.BasePtt, accountBackend *account.Backend, contentBackend *content.Backend, friendBacked *friend.Backend) (*Backend, error) {
	err := InitMe(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	// init-id
	err = initMyInfo(cfg.ID, ptt.MyNodeID(), cfg.PrivateKey, cfg.NodeType)
	if err != nil {
		return nil, err
	}

	backend := &Backend{
		Config: cfg,

		accountBackend: accountBackend,
		contentBackend: contentBackend,
		friendBackend:  friendBacked,
	}

	spm, err := NewServiceProtocolManager(ptt, backend)
	if err != nil {
		return nil, err
	}

	svc, err := pkgservice.NewBaseService(ptt, spm)
	if err != nil {
		return nil, err
	}
	backend.BaseService = svc

	return backend, nil
}

func (b *Backend) Start() error {
	b.SPM().(*ServiceProtocolManager).Start()
	return nil
}

func (b *Backend) Stop() error {
	b.SPM().(*ServiceProtocolManager).Stop()

	log.Debug("Stop: to TeardownMe")

	TeardownMe()

	log.Debug("Stop: after TeardownMe")

	return nil
}

func (b *Backend) Name() string {
	return "me"
}
