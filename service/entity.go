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
	"github.com/ailabstw/go-pttai/common/types"
)

type Entity interface {
	GetID() *types.PttID
	GetCreateTS() types.Timestamp
	GetStatus() types.Status

	GetOwnerID() *types.PttID

	Start() error
	Stop() error

	// implemented in BaseEntity
	PM() ProtocolManager
	Ptt() Ptt
	Service() Service

	Name() string
	SetName(name string)
}

type BaseEntity struct {
	pm      ProtocolManager
	name    string
	ptt     Ptt
	service Service
}

func NewBaseEntity(pm ProtocolManager, name string, ptt Ptt, service Service) (*BaseEntity, error) {
	b := &BaseEntity{
		pm:      pm,
		name:    name,
		ptt:     ptt,
		service: service,
	}

	return b, nil
}

func (b *BaseEntity) GetID() *types.PttID {
	return nil
}

func (b *BaseEntity) GetStatus() types.Status {
	return types.StatusInvalid
}

func (b *BaseEntity) GetOwnerID() *types.PttID {
	return nil
}

func (b *BaseEntity) Start() error {
	return StartPM(b.pm)
}

func (b *BaseEntity) Stop() error {
	return StopPM(b.pm)
}

// implemented in BaseEntity

func (b *BaseEntity) PM() ProtocolManager {
	return b.pm
}

func (b *BaseEntity) Ptt() Ptt {
	return b.ptt
}

func (b *BaseEntity) Service() Service {
	return b.service
}

func (b *BaseEntity) Name() string {
	return b.name
}

func (b *BaseEntity) SetName(name string) {
	b.name = name
}
