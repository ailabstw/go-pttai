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

package friend

import (
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ethereum/go-ethereum/rpc"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Backend struct {
	*pkgservice.BaseService

	accountBackend *account.Backend
	contentBackend *content.Backend
}

func NewBackend(ctx *pkgservice.ServiceContext, cfg *Config, id *types.PttID, ptt pkgservice.Ptt, accountBackend *account.Backend, contentBackend *content.Backend) (*Backend, error) {
	// init friend
	err := InitFriend(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	// backend
	backend := &Backend{
		accountBackend: accountBackend,
		contentBackend: contentBackend,
	}

	// spm
	spm, err := NewServiceProtocolManager(ptt, backend)
	if err != nil {
		return nil, err
	}

	// base-ptt-service
	b, err := pkgservice.NewBaseService(ptt, spm)
	if err != nil {
		return nil, err
	}
	backend.BaseService = b

	return backend, nil

}

func (b *Backend) Start() error {
	b.SPM().(*ServiceProtocolManager).Start()
	return nil
}

func (b *Backend) Stop() error {
	b.SPM().(*ServiceProtocolManager).Stop()

	TeardownFriend()
	return nil
}

func (b *Backend) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "friend",
			Version:   "1.0",
			Service:   NewPrivateAPI(b),
			Public:    pkgservice.IsPrivateAsPublic,
		},
	}
}

func (b *Backend) Name() string {
	return "friend"
}
