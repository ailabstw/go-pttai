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

package service

import (
	"github.com/ailabstw/go-pttai/rpc"
)

type Service interface {
	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []rpc.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error

	SPM() ServiceProtocolManager

	Name() string

	Ptt() Ptt
}

/*
BaseService implements the base-type of Service
*/
type BaseService struct {
	spm ServiceProtocolManager
	ptt Ptt
}

func NewBaseService(ptt Ptt, spm ServiceProtocolManager) (*BaseService, error) {
	return &BaseService{ptt: ptt, spm: spm}, nil
}

func (svc *BaseService) APIs() []rpc.API {
	return nil
}

func (svc *BaseService) Start() error {
	return svc.SPM().Start()
}

func (svc *BaseService) Stop() error {
	return svc.SPM().Stop()
}

func (svc *BaseService) SPM() ServiceProtocolManager {
	return svc.spm
}

func (svc *BaseService) Ptt() Ptt {
	return svc.ptt
}
