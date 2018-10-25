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

/*
ServiceProtocolManager manage service-level operations.

ServiceProtocolManager includes peers-of-services and the corresponding entities.
both are dynamically allocated / deallocated.

When there is a new peer: have all the existing entities trying to register the peer.
When a peer disappear: have all the existing entities trying to unregister the peer.

When there is a new entity: trying to register all the peers.
When a peer disappear: trying to unregister all the peers.
*/
type ServiceProtocolManager interface {
	Start() error
	Stop() error

	Ptt() Ptt
	Service() Service
}

type BaseServiceProtocolManager struct {
	ptt     Ptt
	service Service
}

func NewBaseServiceProtocolManager(ptt Ptt, service Service) (*BaseServiceProtocolManager, error) {
	spm := &BaseServiceProtocolManager{
		ptt: ptt,

		service: service,
	}

	return spm, nil
}

func (b *BaseServiceProtocolManager) Start() error {
	return nil
}

func (b *BaseServiceProtocolManager) Stop() error {
	return nil
}

func (b *BaseServiceProtocolManager) Ptt() Ptt {
	return b.ptt
}

func (b *BaseServiceProtocolManager) Service() Service {
	return b.service
}
