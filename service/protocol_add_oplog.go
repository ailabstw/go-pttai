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

package service

import (
	"encoding/json"
)

type AddOplog struct {
	Oplog *BaseOplog `json:"O"`
}

type AddOplogs struct {
	Oplogs []*BaseOplog `json:"O"`
}

func (pm *BaseProtocolManager) HandleAddOplog(
	dataBytes []byte,
	handleOplogs func(oplogs []*BaseOplog, p *PttPeer, isUpdateSyncTime bool) error,
	peer *PttPeer) error {

	data := &AddOplog{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return handleOplogs([]*BaseOplog{data.Oplog}, peer, false)
}

func (pm *BaseProtocolManager) HandleAddOplogs(
	dataBytes []byte,
	handleOplogs func(oplogs []*BaseOplog, p *PttPeer, isUpdateSyncTime bool) error,
	peer *PttPeer) error {

	data := &AddOplogs{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return handleOplogs(data.Oplogs, peer, false)
}

func (pm *BaseProtocolManager) HandleAddPendingOplog(
	dataBytes []byte,
	handlePendingOplogs func(oplogs []*BaseOplog, p *PttPeer) error,
	peer *PttPeer) error {

	data := &AddOplog{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return handlePendingOplogs([]*BaseOplog{data.Oplog}, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingOplogs(
	dataBytes []byte,
	handlePendingOplogs func(oplogs []*BaseOplog, p *PttPeer) error,
	peer *PttPeer) error {

	data := &AddOplogs{}
	err := json.Unmarshal(dataBytes, data)
	if err != nil {
		return err
	}

	return handlePendingOplogs(data.Oplogs, peer)
}
