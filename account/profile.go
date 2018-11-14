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
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type Profile struct {
	*pkgservice.BaseEntity `json:"e"`
	UpdateTS               types.Timestamp `json:"UT"`

	MyID *types.PttID `json:"m"`
}

func NewEmptyProfile() *Profile {
	return &Profile{BaseEntity: &pkgservice.BaseEntity{}}
}

func NewProfile(myID *types.PttID, ts types.Timestamp, ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager, dbLock *types.LockMap) (*Profile, error) {

	id, err := pkgservice.NewPttIDWithMyID(myID)
	if err != nil {
		return nil, err
	}

	e := pkgservice.NewBaseEntity(id, ts, myID, types.StatusPending, dbAccount, dbLock)
	e.EntityType = pkgservice.EntityTypePersonal

	p := &Profile{
		BaseEntity: e,
		UpdateTS:   ts,

		MyID: myID,
	}

	log.Debug("NewProfile", "id", id)

	err = p.Init(ptt, service, spm)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Profile) GetUpdateTS() types.Timestamp {
	return p.UpdateTS
}

func (p *Profile) SetUpdateTS(ts types.Timestamp) {
	p.UpdateTS = ts
}

func (p *Profile) Init(ptt pkgservice.Ptt, service pkgservice.Service, spm pkgservice.ServiceProtocolManager) error {

	p.SetDB(dbAccount, spm.GetDBLock())

	err := p.InitPM(ptt, service)
	if err != nil {
		return err
	}

	return nil
}

func (p *Profile) InitPM(ptt pkgservice.Ptt, service pkgservice.Service) error {
	pm, err := NewProtocolManager(p, ptt)
	if err != nil {
		log.Error("InitPM: unable to NewProtocolManager", "e", err)
		return err
	}

	p.BaseEntity.Init(pm, ptt, service)

	return nil
}

func (p *Profile) MarshalKey() ([]byte, error) {
	key := append(DBProfilePrefix, p.ID[:]...)

	return key, nil
}

func (p *Profile) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Profile) Unmarshal(theBytes []byte) error {
	err := json.Unmarshal(theBytes, p)
	if err != nil {
		return err
	}

	// postprocess

	return nil
}

func (p *Profile) Save(isLocked bool) error {
	if !isLocked {
		err := p.Lock()
		if err != nil {
			return err
		}
		defer p.Unlock()
	}

	key, err := p.MarshalKey()
	if err != nil {
		return err
	}

	marshaled, err := p.Marshal()
	if err != nil {
		return err
	}

	err = dbAccountCore.Put(key, marshaled)
	if err != nil {
		return err
	}

	return nil
}
